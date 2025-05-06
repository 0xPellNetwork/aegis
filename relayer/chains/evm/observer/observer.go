package observer

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/onrik/ethrpc"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/aegis/pkg/bg"
	"github.com/0xPellNetwork/aegis/relayer/chains/base"
	"github.com/0xPellNetwork/aegis/relayer/chains/evm"
	"github.com/0xPellNetwork/aegis/relayer/chains/evm/observer/handler"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/config"
	pctx "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/db"
	clientlogs "github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

const (
	// defaultAlertLatency is the default alert latency (in seconds) for unit tests
	defaultAlertLatency = 60

	// defaultConfirmationCount is the default confirmation count for unit tests
	defaultConfirmationCount = 2
)

type TxHashEnvelope struct {
	TxHash string
	Done   chan struct{}
}

var _ interfaces.ChainClient = &ChainClient{}

// ChainClient represents the chain configuration for an EVM chain
// Filled with above constants depending on chain
type ChainClient struct {
	// base.Observer implements the base chain observer
	base.Observer

	// evmClient is the EVM client for the observed chain
	evmClient interfaces.EVMRPCClient

	// evmJSONRPC is the EVM JSON RPC client for the observed chain
	evmJSONRPC interfaces.EVMJSONRPCClient
	// TODO: sync map
	// outTxPendingTransactions is the map to index pending transactions by hash
	outTxPendingTransactions map[string]*ethtypes.Transaction

	// outboundConfirmedReceipts is the map to index confirmed receipts by hash
	outTXConfirmedReceipts map[string]*ethtypes.Receipt

	// outboundConfirmedTransactions is the map to index confirmed transactions by hash
	outTXConfirmedTransactions map[string]*ethtypes.Transaction

	evmEventReactor interfaces.IEVMEventReactor

	forceStartHeight uint64

	maxLatestIndexedBlockGap uint64
}

// NewEVMChainClient returns a new configuration based on supplied target chain
func NewEVMChainClient(
	ctx context.Context,
	config config.EVMConfig,
	chainParams relayertypes.ChainParams,
	evmClient interfaces.EVMRPCClient,
	evmJSONRPC interfaces.EVMJSONRPCClient,
	pellcoreClient interfaces.PellCoreBridger,
	tss interfaces.TSSSigner,
	database *db.DB,
	logger clientlogs.Logger,
	ts *metrics.TelemetryServer,
) (*ChainClient, error) {
	appContext, err := pctx.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app context")
	}

	// create base observer
	baseObserver, err := base.NewObserver(
		appContext.PellCoreContext(),
		config.Chain,
		chainParams,
		pellcoreClient,
		tss,
		base.DefaultBlockCacheSize,
		base.DefaultHeaderCacheSize,
		defaultAlertLatency,
		ts,
		database,
		logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create base observer")
	}

	// create evm observer
	ob := &ChainClient{
		Observer:                   *baseObserver,
		evmClient:                  evmClient,
		evmJSONRPC:                 evmJSONRPC,
		outTxPendingTransactions:   make(map[string]*ethtypes.Transaction),
		outTXConfirmedReceipts:     make(map[string]*ethtypes.Receipt),
		outTXConfirmedTransactions: make(map[string]*ethtypes.Transaction),
		forceStartHeight:           config.ForceStartHeight,
		maxLatestIndexedBlockGap:   config.MaxLatestIndexedBlockGap,
	}

	ob.evmEventReactor = handler.NewEVMEventReactor(ob.evmClient, ob.evmJSONRPC, ob.GetChainParams(), ob.Chain(), ob.PellcoreClient(), ob.Logger())

	// load last block scanned
	if err = ob.LoadLastBlockScanned(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to load last block scanned")
	}

	return ob, nil
}

// WithEvmClient attaches a new evm client to the chain client
func (ob *ChainClient) WithEvmClient(client interfaces.EVMRPCClient) {
	ob.evmClient = client
}

// WithEvmJSONRPC attaches a new evm json rpc client to the chain client
func (ob *ChainClient) WithEvmJSONRPC(client interfaces.EVMJSONRPCClient) {
	ob.evmJSONRPC = client
}

// SetChainParams sets the chain params for the chain client
func (ob *ChainClient) SetChainParams(params relayertypes.ChainParams) {
	ob.WithChainParams(params)
}

// GetChainParams returns the chain params for the chain client
func (ob *ChainClient) GetChainParams() relayertypes.ChainParams {
	return ob.ChainParams()
}

// GetEvmReactor returns the evm event reactor for the chain client
func (ob *ChainClient) GetEvmReactor() interfaces.IEVMEventReactor {
	return ob.evmEventReactor
}

// Start all observation routines for the evm chain
func (ob *ChainClient) Start(ctx context.Context) {
	if noop := ob.Observer.Start(); noop {
		ob.Logger().Chain.Info().Msgf("observer is already started for chain %d", ob.Chain().Id)
		return
	}

	ob.Logger().Chain.Info().Msgf("observer is starting for chain %d", ob.Chain().Id)

	bg.Work(ctx, ob.WatchInTx, bg.WithName("WatchInbound"), bg.WithLogger(ob.Logger().Inbound))
	bg.Work(ctx, ob.WatchOutTx, bg.WithName("WatchOutbound"), bg.WithLogger(ob.Logger().Outbound))
	bg.Work(ctx, ob.WatchGasPrice, bg.WithName("WatchGasPrice"), bg.WithLogger(ob.Logger().GasPrice))
	bg.Work(ctx, ob.WatchIntxTracker, bg.WithName("WatchInboundTracker"), bg.WithLogger(ob.Logger().Inbound))
	bg.Work(ctx, ob.WatchRPCStatus, bg.WithName("WatchRPCStatus"), bg.WithLogger(ob.Logger().Chain))
	bg.Work(ctx, ob.WatchPellToken, bg.WithName("WatchPellToken"), bg.WithLogger(ob.Logger().Outbound))
	bg.Work(ctx, ob.WatchGasToken, bg.WithName("watchGasToken"), bg.WithLogger(ob.Logger().Outbound))
}

// SetPendingTx sets the pending transaction in memory
func (ob *ChainClient) SetPendingTx(nonce uint64, transaction *ethtypes.Transaction) {
	ob.Mu().Lock()
	defer ob.Mu().Unlock()
	ob.outTxPendingTransactions[ob.OutboundID(nonce)] = transaction
}

// GetPendingTx gets the pending transaction from memory
func (ob *ChainClient) GetPendingTx(nonce uint64) *ethtypes.Transaction {
	ob.Mu().Lock()
	defer ob.Mu().Unlock()
	return ob.outTxPendingTransactions[ob.OutboundID(nonce)]
}

// SetTxNReceipt sets the receipt and transaction in memory
func (ob *ChainClient) SetTxNReceipt(nonce uint64, receipt *ethtypes.Receipt, transaction *ethtypes.Transaction) {
	ob.Mu().Lock()
	defer ob.Mu().Unlock()
	delete(ob.outTxPendingTransactions, ob.OutboundID(nonce)) // remove pending transaction, if any
	ob.outTXConfirmedReceipts[ob.OutboundID(nonce)] = receipt
	ob.outTXConfirmedTransactions[ob.OutboundID(nonce)] = transaction
}

// GetTxNReceipt gets the receipt and transaction from memory
func (ob *ChainClient) GetTxNReceipt(nonce uint64) (*ethtypes.Receipt, *ethtypes.Transaction) {
	ob.Mu().Lock()
	defer ob.Mu().Unlock()
	receipt := ob.outTXConfirmedReceipts[ob.OutboundID(nonce)]
	transaction := ob.outTXConfirmedTransactions[ob.OutboundID(nonce)]
	return receipt, transaction
}

// IsTxConfirmed returns true if there is a confirmed tx for 'nonce'
func (ob *ChainClient) IsTxConfirmed(nonce uint64) bool {
	ob.Mu().Lock()
	defer ob.Mu().Unlock()
	return ob.outTXConfirmedReceipts[ob.OutboundID(nonce)] != nil &&
		ob.outTXConfirmedTransactions[ob.OutboundID(nonce)] != nil
}

// CheckTxInclusion returns nil only if tx is included at the position indicated by the receipt ([block, index])
func (ob *ChainClient) CheckTxInclusion(tx *ethtypes.Transaction, receipt *ethtypes.Receipt) error {
	block, err := ob.GetBlockByNumberCached(receipt.BlockNumber.Uint64())
	if err != nil {
		return errors.Wrapf(err, "GetBlockByNumberCached error for block %d txHash %s nonce %d",
			receipt.BlockNumber.Uint64(), tx.Hash(), tx.Nonce())
	}
	// #nosec G701 non negative value
	if receipt.TransactionIndex >= uint(len(block.Transactions)) {
		return fmt.Errorf("transaction index %d out of range [0, %d), txHash %s nonce %d block %d",
			receipt.TransactionIndex, len(block.Transactions), tx.Hash(), tx.Nonce(), receipt.BlockNumber.Uint64())
	}
	txAtIndex := block.Transactions[receipt.TransactionIndex]
	if !strings.EqualFold(txAtIndex.Hash, tx.Hash().Hex()) {
		ob.RemoveCachedBlock(receipt.BlockNumber.Uint64()) // clean stale block from cache
		return fmt.Errorf("transaction at index %d has different hash %s, txHash %s nonce %d block %d",
			receipt.TransactionIndex, txAtIndex.Hash, tx.Hash(), tx.Nonce(), receipt.BlockNumber.Uint64())
	}
	return nil
}

// TransactionByHash query transaction by hash via JSON-RPC
func (ob *ChainClient) TransactionByHash(txHash string) (*ethrpc.Transaction, bool, error) {
	tx, err := ob.evmJSONRPC.EthGetTransactionByHash(txHash)
	if err != nil {
		return nil, false, err
	}
	err = evm.ValidateEvmTransaction(tx)
	if err != nil {
		return nil, false, err
	}
	return tx, tx.BlockNumber == nil, nil
}

func (ob *ChainClient) GetBlockHeaderCached(ctx context.Context, blockNumber uint64) (*ethtypes.Header, error) {
	if header, ok := ob.HeaderCache().Get(blockNumber); ok {
		return header.(*ethtypes.Header), nil
	}
	header, err := ob.evmClient.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return nil, err
	}
	ob.HeaderCache().Add(blockNumber, header)
	return header, nil
}

// GetBlockByNumberCached get block by number from cache
// returns block, ethrpc.Block, isFallback, isSkip, error
func (ob *ChainClient) GetBlockByNumberCached(blockNumber uint64) (*ethrpc.Block, error) {
	if block, ok := ob.BlockCache().Get(blockNumber); ok {
		if block, ok := block.(*ethrpc.Block); ok {
			return block, nil
		}
		return nil, errors.New("cached value is not of type *ethrpc.Block")
	}
	if blockNumber > math.MaxInt32 {
		return nil, fmt.Errorf("block number %d is too large", blockNumber)
	}
	// #nosec G701 always in range, checked above
	block, err := ob.BlockByNumber(int(blockNumber))
	if err != nil {
		return nil, err
	}
	ob.BlockCache().Add(blockNumber, block)
	return block, nil
}

// RemoveCachedBlock remove block from cache
func (ob *ChainClient) RemoveCachedBlock(blockNumber uint64) {
	ob.BlockCache().Remove(blockNumber)
}

// BlockByNumber query block by number via JSON-RPC
func (ob *ChainClient) BlockByNumber(blockNumber int) (*ethrpc.Block, error) {
	block, err := ob.evmJSONRPC.EthGetBlockByNumber(blockNumber, true)
	if err != nil {
		return nil, err
	}
	for i := range block.Transactions {
		err := evm.ValidateEvmTransaction(&block.Transactions[i])
		if err != nil {
			return nil, err
		}
	}
	return block, nil
}

// LoadLastBlockScanned loads the last scanned block from the database
// TODO(revamp): move to a db file
func (ob *ChainClient) LoadLastBlockScanned(ctx context.Context) error {
	if ob.forceStartHeight != 0 {
		ob.Logger().Chain.Info().Msgf("chain %d starts scanning from block %d", ob.Chain().Id, ob.forceStartHeight)

		ob.WithLastBlockScanned(ob.forceStartHeight)
		ob.WithInboundBlock(ob.forceStartHeight)
		return nil
	}

	if err := ob.Observer.LoadBlockScanInfo(ob.Logger().Chain); err != nil {
		return errors.Wrapf(err, "error LoadLastBlockScanned for chain %d", ob.Chain().Id)
	}

	// observer will scan from the last block when 'lastBlockScanned == 0', this happens when:
	// 1. environment variable is set explicitly to "latest"
	// 2. environment variable is empty and last scanned block is not found in DB
	if ob.LastBlockScanned() == 0 {
		chainIndex, err := ob.PellcoreClient().GetChainIndex(ctx, ob.Chain().Id)
		if err != nil {
			return errors.Wrapf(err, "error LastInboundBlock for chain %d", ob.Chain().Id)
		}

		height := ob.ChainParams().StartBlockHeight
		if chainIndex.CurrHeight > height {
			height = chainIndex.CurrHeight
		}

		ob.WithLastBlockScanned(height)
	}

	// If there is no local scan record, then use the on-chain data directly
	if ob.LastInboundBlock() == 0 {
		chainIndex, err := ob.PellcoreClient().GetChainIndex(ctx, ob.Chain().Id)
		if err != nil {
			return errors.Wrapf(err, "error LastInboundBlock for chain %d", ob.Chain().Id)

		}

		ob.WithInboundBlock(chainIndex.CurrHeight)
	}

	ob.Logger().Chain.Info().Msgf("chain %d starts scanning from block %d", ob.Chain().Id, ob.LastBlockScanned())

	return nil
}
