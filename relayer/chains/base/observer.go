package base

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	corecontext "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/db"
	"github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
	"github.com/0xPellNetwork/aegis/relayer/pellcore"
	clienttypes "github.com/0xPellNetwork/aegis/relayer/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	// EnvVarLatestBlock is the environment variable that forces the observer to scan from the latest block
	EnvVarLatestBlock = "latest"

	// DefaultBlockCacheSize is the default number of blocks that the observer will keep in cache for performance (without RPC calls)
	// Cached blocks can be used to get block information and verify transactions
	DefaultBlockCacheSize = 1000

	// DefaultHeaderCacheSize is the default number of headers that the observer will keep in cache for performance (without RPC calls)
	// Cached headers can be used to get header information
	DefaultHeaderCacheSize = 1000
)

// ObserverLogger contains the loggers for chain observers
type ObserverLogger struct {
	// the parent logger for the chain observer
	Chain zerolog.Logger

	// the logger for inbound transactions
	Inbound zerolog.Logger

	// the logger for outbound transactions
	Outbound zerolog.Logger

	// the logger for the chain's gas price
	GasPrice zerolog.Logger

	// the logger for block headers
	Headers zerolog.Logger

	// the logger for the compliance check
	Compliance zerolog.Logger
}

// Observer is the base structure for chain observers, grouping the common logic for each chain observer client.
// The common logic includes: chain, chainParams, contexts, pellcore client, tss, lastBlock, db, metrics, loggers etc.
type Observer struct {
	// coreContext is the app context
	// todo: use context included appContext
	coreContext *corecontext.PellCoreContext

	// chain contains static information about the observed chain
	chain chains.Chain

	// chainParams contains the dynamic chain parameters of the observed chain
	chainParams relayertypes.ChainParams

	// pellcoreClient is the client to interact with PellChain
	pellcoreClient interfaces.PellCoreBridger

	// tss is the TSS signer
	tss interfaces.TSSSigner

	// lastBlock is the last block height of the observed chain
	lastBlock uint64

	// lastBlockScanned is the last block height scanned by the observer
	lastBlockScanned uint64

	// rpcAlertLatency is the threshold of RPC latency to trigger an alert
	rpcAlertLatency time.Duration

	// TODO: sync from xmsg chain_index
	// previous vote inbound block number.
	lastInboundBlock uint64

	// blockCache is the cache for blocks
	blockCache *lru.Cache

	// headerCache is the cache for headers
	headerCache *lru.Cache

	// db is the database to persist data
	db *db.DB

	// ts is the telemetry server for metrics
	ts *metrics.TelemetryServer

	// logger contains the loggers used by observer
	logger ObserverLogger

	// mu protects fields from concurrent access
	// Note: base observer simply provides the mutex. It's the sub-struct's responsibility to use it to be thread-safe
	mu      *sync.Mutex
	started bool

	// stop is the channel to signal the observer to stop
	stop chan struct{}
}

// NewObserver creates a new base observer.
func NewObserver(
	coreContext *corecontext.PellCoreContext,
	chain chains.Chain,
	chainParams relayertypes.ChainParams,
	pellcoreClient interfaces.PellCoreBridger,
	tss interfaces.TSSSigner,
	blockCacheSize int,
	headerCacheSize int,
	rpcAlertLatency int64,
	ts *metrics.TelemetryServer,
	database *db.DB,
	logger logs.Logger,
) (*Observer, error) {
	ob := Observer{
		coreContext:      coreContext,
		chain:            chain,
		chainParams:      chainParams,
		pellcoreClient:   pellcoreClient,
		tss:              tss,
		lastBlock:        0,
		lastBlockScanned: 0,
		lastInboundBlock: 0,
		rpcAlertLatency:  time.Duration(rpcAlertLatency) * time.Second,
		ts:               ts,
		db:               database,
		mu:               &sync.Mutex{},
		stop:             make(chan struct{}),
	}

	// setup loggers
	ob.WithLogger(logger)

	// create block cache
	var err error
	ob.blockCache, err = lru.New(blockCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "error creating block cache")
	}

	// create header cache
	ob.headerCache, err = lru.New(headerCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "error creating header cache")
	}

	return &ob, nil
}

// Start starts the observer. Returns true if the observer was already started (noop).
func (ob *Observer) Start() bool {
	ob.mu.Lock()
	defer ob.Mu().Unlock()

	// noop
	if ob.started {
		return true
	}

	ob.started = true

	return false
}

// Stop notifies all goroutines to stop and closes the database.
func (ob *Observer) Stop() {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.started {
		ob.logger.Chain.Info().Msgf("Observer already stopped")
		return
	}

	ob.logger.Chain.Info().Msgf("Stopping observer")

	close(ob.stop)
	ob.started = false

	// close database
	if err := ob.db.Close(); err != nil {
		ob.Logger().Chain.Error().Err(err).Msgf("unable to close db")
	}

	ob.Logger().Chain.Info().Msgf("observer stopped")
}

// AppContext returns the app context.
func (ob *Observer) AppContext() *corecontext.PellCoreContext {
	return ob.coreContext
}

// Chain returns the chain for the observer.
func (ob *Observer) Chain() chains.Chain {
	return ob.chain
}

// WithChain attaches a new chain to the observer.
func (ob *Observer) WithChain(chain chains.Chain) *Observer {
	ob.chain = chain
	return ob
}

// ChainParams returns the chain params for the observer.
func (ob *Observer) ChainParams() relayertypes.ChainParams {
	return ob.chainParams
}

// WithChainParams attaches a new chain params to the observer.
func (ob *Observer) WithChainParams(params relayertypes.ChainParams) *Observer {
	ob.chainParams = params
	return ob
}

// PellcoreClient returns the pellcore client for the observer.
func (ob *Observer) PellcoreClient() interfaces.PellCoreBridger {
	return ob.pellcoreClient
}

// WithPellcoreClient attaches a new pellcore client to the observer.
func (ob *Observer) WithPellcoreClient(client interfaces.PellCoreBridger) *Observer {
	ob.pellcoreClient = client
	return ob
}

// Tss returns the tss signer for the observer.
func (ob *Observer) TSS() interfaces.TSSSigner {
	return ob.tss
}

// WithTSS attaches a new tss signer to the observer.
func (ob *Observer) WithTSS(tss interfaces.TSSSigner) *Observer {
	ob.tss = tss
	return ob
}

// LastBlock get external last block height.
func (ob *Observer) LastBlock() uint64 {
	return atomic.LoadUint64(&ob.lastBlock)
}

// WithLastBlock set external last block height.
func (ob *Observer) WithLastBlock(lastBlock uint64) *Observer {
	atomic.StoreUint64(&ob.lastBlock, lastBlock)
	return ob
}

// LastBlockScanned get last block scanned (not necessarily caught up with the chain; could be slow/paused).
func (ob *Observer) LastBlockScanned() uint64 {
	height := atomic.LoadUint64(&ob.lastBlockScanned)
	return height
}

// WithLastBlockScanned set last block scanned (not necessarily caught up with the chain; could be slow/paused).
func (ob *Observer) WithLastBlockScanned(blockNumber uint64) *Observer {
	atomic.StoreUint64(&ob.lastBlockScanned, blockNumber)
	metrics.LastScannedBlockNumber.WithLabelValues(ob.chain.ChainName()).Set(float64(blockNumber))
	return ob
}

// LastInboundBlock get external last block height.
func (ob *Observer) LastInboundBlock() uint64 {
	return atomic.LoadUint64(&ob.lastInboundBlock)
}

// WithLastBlock set external last block height.
func (ob *Observer) WithInboundBlock(lastBlock uint64) *Observer {
	atomic.StoreUint64(&ob.lastInboundBlock, lastBlock)
	return ob
}

// BlockCache returns the block cache for the observer.
func (ob *Observer) BlockCache() *lru.Cache {
	return ob.blockCache
}

// WithBlockCache attaches a new block cache to the observer.
func (ob *Observer) WithBlockCache(cache *lru.Cache) *Observer {
	ob.blockCache = cache
	return ob
}

// HeaderCache returns the header cache for the observer.
func (ob *Observer) HeaderCache() *lru.Cache {
	return ob.headerCache
}

// WithHeaderCache attaches a new header cache to the observer.
func (ob *Observer) WithHeaderCache(cache *lru.Cache) *Observer {
	ob.headerCache = cache
	return ob
}

// OutboundID returns a unique identifier for the outbound transaction.
// The identifier is now used as the key for maps that store outbound related data (e.g. transaction, receipt, etc).
func (ob *Observer) OutboundID(nonce uint64) string {
	// all chains uses EVM address as part of the key except bitcoin
	tssAddress := ob.tss.EVMAddress().String()
	return fmt.Sprintf("%d-%s-%d", ob.chain.Id, tssAddress, nonce)
}

// DB returns the database for the observer.
func (ob *Observer) DB() *db.DB {
	return ob.db
}

// WithTelemetryServer attaches a new telemetry server to the observer.
func (ob *Observer) WithTelemetryServer(ts *metrics.TelemetryServer) *Observer {
	ob.ts = ts
	return ob
}

// TelemetryServer returns the telemetry server for the observer.
func (ob *Observer) TelemetryServer() *metrics.TelemetryServer {
	return ob.ts
}

// Logger returns the logger for the observer.
func (ob *Observer) Logger() *ObserverLogger {
	return &ob.logger
}

// WithLogger attaches a new logger to the observer.
func (ob *Observer) WithLogger(logger logs.Logger) *Observer {
	chainLogger := logger.Std.With().Int64(logs.FieldChain, ob.chain.Id).Logger()
	ob.logger = ObserverLogger{
		Chain:      chainLogger,
		Inbound:    chainLogger.With().Str(logs.FieldModule, logs.ModNameInbound).Logger(),
		Outbound:   chainLogger.With().Str(logs.FieldModule, logs.ModNameOutbound).Logger(),
		GasPrice:   chainLogger.With().Str(logs.FieldModule, logs.ModNameGasPrice).Logger(),
		Headers:    chainLogger.With().Str(logs.FieldModule, logs.ModNameHeaders).Logger(),
		Compliance: logger.Compliance,
	}
	return ob
}

// Mu returns the mutex for the observer.
func (ob *Observer) Mu() *sync.Mutex {
	return ob.mu
}

// StopChannel returns the stop channel for the observer.
func (ob *Observer) StopChannel() chan struct{} {
	return ob.stop
}

// LoadBlockScanInfo loads last scanned block from environment variable or from database.
// The last scanned block is the height from which the observer should continue scanning.
func (ob *Observer) LoadBlockScanInfo(logger zerolog.Logger) error {
	// get environment variable
	envvar := EnvVarLatestBlockByChain(ob.chain)
	scanFromBlock := os.Getenv(envvar)

	// load from environment variable if set
	if scanFromBlock != "" {
		logger.Info().
			Msgf("LoadLastBlockScanned: envvar %s is set; scan from  block %s", envvar, scanFromBlock)
		if scanFromBlock == EnvVarLatestBlock {
			return nil
		}
		blockNumber, err := strconv.ParseUint(scanFromBlock, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "unable to parse block number from ENV %s=%s", envvar, scanFromBlock)
		}
		ob.WithLastBlockScanned(blockNumber)
		return nil
	}

	// load from DB otherwise. If not found, start from latest block
	blockNumber, lastInboundBlock, err := ob.ReadBlockScanInfoFromDB()
	if err != nil {
		logger.Info().Msgf("LoadLastBlockScanned: last scanned block not found in db")
		return nil
	}

	ob.WithLastBlockScanned(blockNumber)
	ob.WithInboundBlock(lastInboundBlock)
	return nil
}

// SaveScanInfo saves the last scanned block to memory and database.
func (ob *Observer) SaveScanInfo(blockNumber, lastInboundBlock uint64) error {
	ob.WithLastBlockScanned(blockNumber)
	return ob.db.Client().Save(clienttypes.ToLastBlockSQLType(blockNumber, lastInboundBlock)).Error
}

// ReadBlockScanInfoFromDB reads the last scanned block from the database.
func (ob *Observer) ReadBlockScanInfoFromDB() (uint64, uint64, error) {
	var lastBlock clienttypes.LastBlockSQLType
	if err := ob.db.Client().First(&lastBlock, clienttypes.LastBlockNumID).Error; err != nil {
		// record not found
		return 0, 0, err
	}
	return lastBlock.Num, lastBlock.LastInboundBlock, nil
}

// PostVoteInbound posts a vote for the given vote message
func (ob *Observer) PostVoteInbound(
	ctx context.Context,
	msg *xmsgtypes.MsgVoteOnObservedInboundTx,
	retryGasLimit uint64,
) (string, error) {
	txHash := msg.InTxHash
	pellHash, ballot, err := ob.pellcoreClient.
		PostVoteInboundEvents(ctx, pellcore.PostVoteInboundGasLimit, retryGasLimit, []*xmsgtypes.MsgVoteOnObservedInboundTx{msg})
	if err != nil {
		ob.logger.Inbound.Err(err).
			Msgf("PostVoteInbound: error posting vote event tx %s", txHash)
		return "", err
	} else if pellHash != "" {
		ob.logger.Inbound.Info().Msgf("PostVoteInbound: event tx %s post vote %s ballot %s", txHash, pellHash, ballot)
	} else {
		ob.logger.Inbound.Info().Msgf("PostVoteInbound: event tx %s already post vote ballot %s", txHash, ballot)
	}

	return ballot, err
}

// PostVoteInboundBlock posts a vote for the given inbound block and events
// aggregates information from a block, including the block proof and event messages, into a single transaction for submission to the blockchain.
func (ob *Observer) PostVoteInboundBlock(
	ctx context.Context,
	gasLimit, retryGasLimit uint64,
	blockProof *xmsgtypes.MsgVoteInboundBlock,
	events []*xmsgtypes.MsgVoteOnObservedInboundTx,
) ([]string, []string, error) {
	txHash, ballot, err := ob.pellcoreClient.PostVoteInboundBlock(ctx, gasLimit, retryGasLimit, blockProof, events)
	if err != nil {
		ob.logger.Inbound.Err(err).
			Msgf("PostVoteInboundBlock: error posting vote in block %s tx hash %s", blockProof.BlockProof.BlockHash, txHash)
		return nil, nil, err
	} else if len(txHash) > 0 {
		ob.logger.Inbound.Info().Msgf("PostVoteInboundBlock: post vote %s ballot %s", txHash, ballot)
	} else {
		ob.logger.Inbound.Info().Msgf("PostVoteInboundBlock: already post vote ballot %s", ballot)
	}

	return txHash, ballot, nil
}

// AlertOnRPCLatency prints an alert if the RPC latency exceeds the threshold.
// Returns true if the RPC latency is too high.
func (ob *Observer) AlertOnRPCLatency(latestBlockTime time.Time, defaultAlertLatency time.Duration) bool {
	// use configured alert latency if set
	alertLatency := defaultAlertLatency
	if ob.rpcAlertLatency > 0 {
		alertLatency = ob.rpcAlertLatency
	}

	// latest block should not be too old
	elapsedTime := time.Since(latestBlockTime)
	if elapsedTime > alertLatency {
		ob.logger.Chain.Error().
			Msgf("RPC is stale: latest block is %.0f seconds old, RPC down or chain stuck (check explorer)?", elapsedTime.Seconds())
		return true
	}

	ob.logger.Chain.Info().Msgf("RPC is OK: latest block is %.0f seconds old", elapsedTime.Seconds())
	return false
}

// EnvVarLatestBlockByChain returns the environment variable for the last block by chain.
func EnvVarLatestBlockByChain(chain chains.Chain) string {
	return fmt.Sprintf("CHAIN_%d_SCAN_FROM_BLOCK", chain.Id)
}

// EnvVarLatestTxByChain returns the environment variable for the last tx by chain.
func EnvVarLatestTxByChain(chain chains.Chain) string {
	return fmt.Sprintf("CHAIN_%d_SCAN_FROM_TX", chain.Id)
}
