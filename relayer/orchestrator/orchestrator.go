package orchestrator

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/0xPellNetwork/aegis/pkg/bg"
	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/constant"
	pellmath "github.com/0xPellNetwork/aegis/pkg/math"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	appcontext "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
	"github.com/0xPellNetwork/aegis/relayer/outtxprocessor"
	"github.com/0xPellNetwork/aegis/relayer/ratelimiter"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	// evmOutboundLookbackFactor is the factor to determine how many nonces to look back for pending cctxs
	// For example, give OutboundScheduleLookahead of 120, pending NonceLow of 1000 and factor of 1.0,
	// the scheduler need to be able to pick up and schedule any pending cctx with nonce < 880 (1000 - 120 * 1.0)
	// NOTE: 1.0 means look back the same number of cctxs as we look ahead
	evmOutboundLookbackFactor = 1.0

	// sampling rate for sampled orchestrator logger
	loggerSamplingRate = 10
)

var defaultLogSampler = &zerolog.BasicSampler{N: loggerSamplingRate}

type multiLogger struct {
	zerolog.Logger
	Sampled zerolog.Logger
}

// Orchestrator wraps the pellcore client, chain observers and signers.
// This is the high level object used for CCTX scheduling
type Orchestrator struct {
	// pellcore client
	pellcoreClient interfaces.PellCoreBridger

	// signerMap contains the chain signers indexed by chainID
	signerMap map[int64]interfaces.ChainSigner

	// observerMap contains the chain observers indexed by chainID
	observerMap map[int64]interfaces.ChainClient

	// outbound processor
	outboundProc *outtxprocessor.Processor

	// last operator balance
	lastOperatorBalance sdkmath.Int

	// observer & signer props
	tss         interfaces.TSSSigner
	dbDirectory string
	baseLogger  logs.Logger

	// misc
	logger multiLogger
	ts     *metrics.TelemetryServer
	stop   chan struct{}
	mu     sync.RWMutex
}

// New creates a new CoreObserver
func New(
	ctx context.Context,
	pellcoreClient interfaces.PellCoreBridger,
	tss interfaces.TSSSigner,
	dbDirectory string,
	logger logs.Logger,
	ts *metrics.TelemetryServer,
) (*Orchestrator, error) {
	log := multiLogger{
		Logger:  logger.Std.With().Str("module", "orchestrator").Logger(),
		Sampled: logger.Std.With().Str("module", "orchestrator").Logger().Sample(defaultLogSampler),
	}

	// CreateSignerMap: This creates a map of all signers for each chain.
	// Each signer is responsible for signing transactions for a particular chain
	signerMap, err := CreateSignerMap(ctx, tss, logger, ts)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create signer map")
	}

	// Creates a map of all chain observers for each chain.
	// Each chain observer is responsible for observing events on the chain and processing them.
	observerMap, err := CreateChainObserverMap(
		ctx,
		pellcoreClient,
		tss,
		dbDirectory,
		logger,
		ts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create observer map")
	}

	balance, err := pellcoreClient.GetPellHotKeyBalance(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get last balance of the hot key")
	}

	return &Orchestrator{
		pellcoreClient: pellcoreClient,

		signerMap:   signerMap,
		observerMap: observerMap,

		outboundProc:        outtxprocessor.NewOutTxProcessor(logger.Std),
		lastOperatorBalance: balance,

		// observer & signer props
		tss:         tss,
		dbDirectory: dbDirectory,
		baseLogger:  logger,

		logger: log,
		ts:     ts,
		stop:   make(chan struct{}),
	}, nil
}

func (co *Orchestrator) Start(ctx context.Context) error {
	signerAddress, err := co.pellcoreClient.GetKeys().GetAddress()
	if err != nil {
		return errors.Wrap(err, "unable to get signer address")
	}

	co.logger.Info().Str("signer", signerAddress.String()).Msg("Starting orchestrator")

	// go co.startXmsgScheduler(ctx)
	// start cctx scheduler
	bg.Work(ctx, co.runScheduler, bg.WithName("runScheduler"), bg.WithLogger(co.logger.Logger))
	bg.Work(ctx, co.runObserverSignerSync, bg.WithName("runObserverSignerSync"), bg.WithLogger(co.logger.Logger))

	shutdownOrchestrator := func() {
		// now stop orchestrator and all observers
		close(co.stop)
	}

	co.pellcoreClient.OnBeforeStop(shutdownOrchestrator)

	return nil
}

func (co *Orchestrator) getSigner(chainID int64) (interfaces.ChainSigner, error) {
	co.mu.RLock()
	defer co.mu.RUnlock()

	s, found := co.signerMap[chainID]
	if !found {
		return nil, fmt.Errorf("signer not found for chainID %d", chainID)
	}

	return s, nil
}

// GetUpdatedSigner returns signer with updated chain parameters
func (co *Orchestrator) resolveSigner(coreContext *appcontext.PellCoreContext, chainID int64) (interfaces.ChainSigner, error) {
	signer, err := co.getSigner(chainID)
	if err != nil {
		return nil, err
	}

	// update EVM signer parameters only. BTC signer doesn't use chain parameters for now.
	if chains.IsEVMChain(chainID) {
		params, found := coreContext.GetEVMChainParams(chainID)
		if found {
			pellConnectorAddress := ethcommon.HexToAddress(params.GetConnectorContractAddress())
			if pellConnectorAddress != signer.GetPellConnectorAddress() {
				signer.SetPellConnectorAddress(pellConnectorAddress)
				co.logger.Info().
					Str("signer.pellconnector_address", params.GetConnectorContractAddress()).
					Msgf("update pell connector address for chainID %d, new address: %s", chainID, pellConnectorAddress)
			}
		}
	}
	return signer, nil
}

func (co *Orchestrator) getObserver(chainID int64) (interfaces.ChainClient, error) {
	co.mu.RLock()
	defer co.mu.RUnlock()

	ob, found := co.observerMap[chainID]
	if !found {
		return nil, fmt.Errorf("observer not found for chainID %d", chainID)
	}

	return ob, nil
}

// returns chain observer with updated chain parameters
func (co *Orchestrator) resolveObserver(coreContext *appcontext.PellCoreContext, chainID int64) (interfaces.ChainClient, error) {
	observer, err := co.getObserver(chainID)
	if err != nil {
		return nil, err
	}

	// update chain client chain parameters
	curParams := observer.GetChainParams()
	if chains.IsEVMChain(chainID) {
		evmParams, found := coreContext.GetEVMChainParams(chainID)
		if found && !cmp.Equal(curParams, *evmParams) {
			observer.SetChainParams(*evmParams)
			co.logger.Info().
				Interface("observer.chain_params", *evmParams).
				Msgf("updated chain params for chainID %d, new params: %v", chainID, *evmParams)
		}
	}
	return observer, nil
}

// GetPendingXmsgsWithinRateLimit get pending cctxs across foreign chains within rate limit
func (co *Orchestrator) GetPendingXmsgsWithinRateLimit(ctx context.Context, chainIDs []int64) (
	map[int64][]*types.Xmsg,
	error,
) {
	// get rate limiter flags
	rateLimitFlags, err := co.pellcoreClient.GetRateLimiterFlags(ctx)
	if err != nil {
		return nil, err
	}

	// apply rate limiter or not according to the flags
	rateLimiterUsable := ratelimiter.IsRateLimiterUsable(rateLimitFlags)

	// fallback to non-rate-limited query if rate limiter is not usable
	xmsgsMap := make(map[int64][]*types.Xmsg)
	if !rateLimiterUsable {
		for _, chainID := range chainIDs {
			resp, _, err := co.pellcoreClient.ListPendingXmsg(ctx, chainID)
			if err == nil && resp != nil {
				xmsgsMap[chainID] = resp
			}
		}
		return xmsgsMap, nil
	}

	// query rate limiter input
	resp, err := co.pellcoreClient.GetRateLimiterInput(ctx, rateLimitFlags.Window)
	if err != nil {
		return nil, err
	}
	input, ok := ratelimiter.NewInput(*resp)
	if !ok {
		return nil, fmt.Errorf("failed to create rate limiter input")
	}

	// apply rate limiter
	output := ratelimiter.ApplyRateLimiter(input, rateLimitFlags.Window, rateLimitFlags.Rate)

	// set metrics
	percentage := pellmath.Percentage(output.CurrentWithdrawRate.BigInt(), rateLimitFlags.Rate.BigInt())
	if percentage != nil {
		percentageFloat, _ := percentage.Float64()
		metrics.PercentageOfRateReached.Set(percentageFloat)
		co.logger.Sampled.Info().Msgf("current rate limiter window: %d rate: %s, percentage: %f",
			output.CurrentWithdrawWindow, output.CurrentWithdrawRate.String(), percentageFloat)
	}

	return output.XmsgsMap, nil
}

// startXmsgScheduler schedules keysigns for cctxs on each PellChain block (the ticker)
func (co *Orchestrator) runScheduler(ctx context.Context) error {
	appContext, err := appcontext.FromContext(ctx)
	if err != nil {
		return err
	}

	observeTicker := time.NewTicker(3 * time.Second)
	var lastBlockNum int64
	for {
		select {
		case <-co.stop:
			co.logger.Warn().Msg("startXmsgScheduler: stopped")
			return nil
		case <-observeTicker.C:
			{
				bn, err := co.pellcoreClient.GetBlockHeight(ctx)
				if err != nil {
					co.logger.Error().Err(err).Msg("startXmsgScheduler: GetPellBlockHeight fail")
					continue
				}
				if bn < 0 {
					co.logger.Error().Msg("startXmsgScheduler: GetPellBlockHeight returned negative height")
					continue
				}
				if lastBlockNum == 0 {
					lastBlockNum = bn - 1
				}
				if bn > lastBlockNum { // we have a new block
					bn = lastBlockNum + 1
					if bn%10 == 0 {
						co.logger.Debug().Msgf("startXmsgScheduler: PellCore heart beat: %d", bn)
					}

					balance, err := co.pellcoreClient.GetPellHotKeyBalance(ctx)
					if err != nil {
						co.logger.Error().Err(err).Msgf("couldn't get operator balance")
					} else {
						diff := co.lastOperatorBalance.Sub(balance)
						if diff.GT(sdkmath.NewInt(0)) && diff.LT(sdkmath.NewInt(math.MaxInt64)) {
							co.ts.AddFeeEntry(bn, diff.Int64())
							co.lastOperatorBalance = balance
						}
					}

					// set current hot key burn rate
					metrics.HotKeyBurnRate.Set(float64(co.ts.HotKeyBurnRate.GetBurnRate().Int64()))

					// schedule keysign for pending cctxs on each chain
					coreContext := appContext.PellCoreContext()
					supportedChains := coreContext.GetEnabledChains()

					// get chain ids without pell chain
					chainIDs := lo.FilterMap(supportedChains, func(c chains.Chain, _ int) (int64, bool) {
						return c.Id, !c.IsPellChain()
					})

					// query pending cctxs across all external chains within rate limit
					xmsgMap, err := co.GetPendingXmsgsWithinRateLimit(ctx, chainIDs)
					if err != nil {
						co.logger.Error().Err(err).Msgf("runScheduler: GetPendingXmsgsWithinRatelimit failed")
					}

					for _, c := range supportedChains {
						if c.IsPellChain() {
							continue
						}

						// update chain parameters for signer and chain client
						signer, err := co.resolveSigner(coreContext, c.Id)
						if err != nil {
							co.logger.Error().Err(err).Msgf("startXmsgScheduler: getUpdatedSigner failed for chain %d", c.Id)
							continue
						}

						ob, err := co.resolveObserver(coreContext, c.Id)
						if err != nil {
							co.logger.Error().Err(err).Msgf("startXmsgScheduler: getTargetChainOb failed for chain %d", c.Id)
							continue
						}

						// get cctxs from map and set pending transactions prometheus gauge
						xmsgList := xmsgMap[c.Id]

						metrics.PendingTxsPerChain.WithLabelValues(c.ChainName()).Set(float64(len(xmsgList)))

						if len(xmsgList) == 0 {
							continue
						}

						if !coreContext.IsOutboundObservationEnabled(ob.GetChainParams()) {
							continue
						}

						// #nosec G701 range is verified
						pellHeight := uint64(bn)

						if chains.IsEVMChain(c.Id) {
							co.scheduleXmsgEVM(ctx, pellHeight, c.Id, xmsgList, ob, signer)
						} else {
							co.logger.Error().Msgf("startXmsgScheduler: no scheduler found chain %d", c.Id)
							continue
						}
					}

					// update last processed block number
					lastBlockNum = bn
					co.ts.SetCoreBlockNumber(lastBlockNum)
				}
			}
		}
	}
}

// // getAllPendingXmsgWithRatelimit get pending cctxs across all foreign chains with rate limit
// func (co *Orchestrator) getAllPendingXmsgWithRatelimit() (map[int64][]*types.Xmsg, int64, string, error) {
// 	cctxList, totalPending, withdrawWindow, withdrawRate, rateLimitExceeded, err := co.pellcoreClient.ListPendingXmsgWithinRatelimit()
// 	if err != nil {
// 		return nil, 0, "", err
// 	}
// 	if rateLimitExceeded {
// 		co.logger.Warn().Msgf("rate limit exceeded, fetched %d cctxs out of %d", len(cctxList), totalPending)
// 	}

// 	// classify pending cctxs by chain id
// 	cctxMap := make(map[int64][]*types.Xmsg)
// 	for _, cctx := range cctxList {
// 		chainID := cctx.GetCurrentOutTxParam().ReceiverChainId
// 		if _, found := cctxMap[chainID]; !found {
// 			cctxMap[chainID] = make([]*types.Xmsg, 0)
// 		}
// 		cctxMap[chainID] = append(cctxMap[chainID], cctx)
// 	}

// 	return cctxMap, withdrawWindow, withdrawRate, nil
// }

// scheduleXmsgEVM schedules evm outtx keysign on each PellChain block (the ticker)
func (co *Orchestrator) scheduleXmsgEVM(
	ctx context.Context,
	pellHeight uint64,
	chainID int64,
	xmsgList []*types.Xmsg,
	observer interfaces.ChainClient,
	signer interfaces.ChainSigner,
) {
	res, err := co.pellcoreClient.GetAllOutTxTrackerByChain(ctx, chainID, interfaces.Ascending)
	if err != nil {
		co.logger.Warn().Err(err).Msgf("scheduleXmsgEVM: GetAllOutTxTrackerByChain failed for chain %d", chainID)
		return
	}
	trackerMap := make(map[uint64]bool)
	for _, v := range res {
		trackerMap[v.Nonce] = true
	}

	outboundScheduleLookahead := observer.GetChainParams().OutboundTxScheduleLookahead
	// #nosec G115 always in range
	outboundScheduleLookback := uint64(float64(outboundScheduleLookahead) * evmOutboundLookbackFactor)
	// #nosec G115 positive
	outboundScheduleInterval := uint64(observer.GetChainParams().OutboundTxScheduleInterval)

	criticalInterval := uint64(10)                      // for critical pending outbound we reduce re-try interval
	nonCriticalInterval := outboundScheduleInterval * 2 // for non-critical pending outbound we increase re-try interval

	for idx, xsmg := range xmsgList {
		params := xsmg.GetCurrentOutTxParam()
		nonce := params.OutboundTxTssNonce
		outTxID := outtxprocessor.ToOutTxID(xsmg.Index, params.ReceiverChainId, nonce)

		co.logger.Info().Msgf("ScheduleXmsgEVM: outtx %s chainid %d nonce %d xmsg %s", outTxID, params.ReceiverChainId, nonce, xsmg.Index)

		if params.ReceiverChainId != chainID {
			co.logger.Error().Msgf("ScheduleXmsgEVM: outtx %s chainid mismatch: want %d, got %d", outTxID, chainID, params.ReceiverChainId)
			continue
		}
		if params.OutboundTxTssNonce > xmsgList[0].GetCurrentOutTxParam().OutboundTxTssNonce+outboundScheduleLookback {
			co.logger.Error().Msgf("ScheduleXmsgEVM: nonce too high: signing %d, earliest pending %d, chain %d",
				params.OutboundTxTssNonce, xmsgList[0].GetCurrentOutTxParam().OutboundTxTssNonce, chainID)
			break
		}

		// try confirming the outtx
		included, _, err := observer.IsOutboundProcessed(ctx, xsmg, co.baseLogger.Std)
		if err != nil {
			co.logger.Error().Err(err).Msgf("ScheduleXmsgEVM: VoteOutboundIfConfirmed failed for chain %d nonce %d", chainID, nonce)
			continue
		}
		if included {
			co.logger.Info().Msgf("ScheduleXmsgEVM: outbound %s already processed; do not schedule keysign", outTxID)
			continue
		}

		// determining critical outbound; if it satisfies following criteria
		// 1. it's the first pending outbound for this chain
		// 2. the following 5 nonces have been in tracker
		if nonce%criticalInterval == pellHeight%criticalInterval {
			count := 0
			for i := nonce + 1; i <= nonce+10; i++ {
				if _, found := trackerMap[i]; found {
					count++
				}
			}
			if count >= 5 {
				outboundScheduleInterval = criticalInterval
			}
		}
		// if it's already in tracker, we increase re-try interval
		if _, ok := trackerMap[nonce]; ok {
			outboundScheduleInterval = nonCriticalInterval
		}

		// otherwise, the normal inooerval is used
		// otherwise, the normal interval is used
		if nonce%outboundScheduleInterval == pellHeight%outboundScheduleInterval &&
			!co.outboundProc.IsOutTxActive(outTxID) {
			co.outboundProc.StartTryProcess(outTxID)
			co.logger.Debug().Msgf("ScheduleXmsgEVM: sign outbound %s", outTxID)
			go signer.TryProcessOutTx(
				ctx,
				xsmg,
				co.outboundProc,
				outTxID,
				observer,
				co.pellcoreClient,
				pellHeight)
		}

		// #nosec G701 always in range
		if int64(idx) >= outboundScheduleLookahead-1 { // only look at 'lookahead' cctxs per chain
			break
		}
	}
}

// runObserverSignerSync runs a blocking ticker that observes chain changes from pellcore
// and optionally (de)provisions respective observers and signers.
func (co *Orchestrator) runObserverSignerSync(ctx context.Context) error {
	// sync observers and signers right away to speed up pellclient startup
	if err := co.syncObserverSigner(ctx); err != nil {
		co.logger.Error().Err(err).Msg("runObserverSignerSync: syncObserverSigner failed for initial sync")
	}

	// sync observer and signer every 10 blocks (approx. 1 minute)
	const cadence = 10 * constant.PellBlockTime

	ticker := time.NewTicker(cadence)
	defer ticker.Stop()

	for {
		select {
		case <-co.stop:
			co.logger.Warn().Msg("runObserverSignerSync: stopped")
			return nil
		case <-ticker.C:
			if err := co.syncObserverSigner(ctx); err != nil {
				co.logger.Error().Err(err).Msg("runObserverSignerSync: syncObserverSigner failed")
			}
		}
	}
}

// syncs and provisions observers & signers.
// Note that zctx.AppContext Update is a responsibility of another component
// See pellcore.Client{}.UpdateAppContextWorker
func (co *Orchestrator) syncObserverSigner(ctx context.Context) error {
	co.mu.Lock()
	defer co.mu.Unlock()

	client := co.pellcoreClient

	added, removed, err := syncObserverMap(ctx, client, co.tss, co.dbDirectory, co.baseLogger, co.ts, &co.observerMap)
	if err != nil {
		return errors.Wrap(err, "syncObserverMap failed")
	}

	if added+removed > 0 {
		co.logger.Info().
			Int("observer.added", added).
			Int("observer.removed", removed).
			Msg("synced observers")
	}

	added, removed, err = syncSignerMap(ctx, co.tss, co.baseLogger, co.ts, &co.signerMap)
	if err != nil {
		return errors.Wrap(err, "syncSignerMap failed")
	}

	if added+removed > 0 {
		co.logger.Info().
			Int("signers.added", added).
			Int("signers.removed", removed).
			Msg("synced signers")
	}

	return nil
}
