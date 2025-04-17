package observer

import (
	"fmt"
	"sort"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/net/context"

	"github.com/pell-chain/pellcore/pkg/bg"
	"github.com/pell-chain/pellcore/pkg/ticker"
	"github.com/pell-chain/pellcore/relayer/config"
	pctx "github.com/pell-chain/pellcore/relayer/context"
	"github.com/pell-chain/pellcore/relayer/metrics"
	"github.com/pell-chain/pellcore/relayer/pellcore"
	clienttypes "github.com/pell-chain/pellcore/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// WatchInbound watches evm chain for incoming txs and post votes to pellcore
// TODO(revamp): move ticker function to a separate file
func (ob *ChainClient) WatchInTx(ctx context.Context) error {
	sampledLogger := ob.Logger().Inbound.Sample(&zerolog.BasicSampler{N: 10})
	interval := ticker.SecondsFromUint64(ob.GetChainParams().InTxTicker)
	task := func(ctx context.Context, t *ticker.Ticker) error {
		return ob.watchInboundOnce(ctx, t, sampledLogger)
	}

	t := ticker.New(interval, task)

	bg.Work(ctx, func(_ context.Context) error {
		<-ob.StopChannel()
		t.Stop()
		ob.Logger().Inbound.Info().Msg("WatchInbound stopped")
		return nil
	})

	ob.Logger().Inbound.Info().Msgf("WatchInbound started")

	return t.Run(ctx)
}

// WatchInTx watches evm chain for incoming txs and post votes to pellcore
func (ob *ChainClient) watchInboundOnce(ctx context.Context, t *ticker.Ticker, sampledLogger zerolog.Logger) error {
	app, err := pctx.FromContext(ctx)
	if err != nil {
		return err
	}

	if !app.PellCoreContext().IsInboundObservationEnabled(ob.GetChainParams()) {
		ob.Logger().Inbound.Warn().Msg("WatchInbound: inbound observation is disabled")
		return nil
	}

	if err := ob.observeInTX(ctx, sampledLogger); err != nil {
		ob.Logger().Inbound.Err(err).Msg("WatchInbound: observeInbound error")
	}

	newInterval := ticker.SecondsFromUint64(ob.GetChainParams().InTxTicker)
	t.SetInterval(newInterval)

	return nil
}

// WatchInboundTracker gets a list of Inbound tracker suggestions from pell-core at each tick and tries to check if the in-tx was confirmed.
// If it was, it tries to broadcast the confirmation vote. If this pell client has previously broadcast the vote, the tx would be rejected
// TODO(revamp): move inbound tracker function to a separate file
func (ob *ChainClient) WatchIntxTracker(ctx context.Context) error {
	app, err := pctx.FromContext(ctx)
	if err != nil {
		return err
	}

	ticker, err := clienttypes.NewDynamicTicker(
		fmt.Sprintf("EVM_WatchInboundTracker_%d", ob.Chain().Id),
		ob.GetChainParams().InTxTicker,
	)
	if err != nil {
		ob.Logger().Inbound.Err(err).Msg("error creating ticker")
		return err
	}
	defer ticker.Stop()

	ob.Logger().Inbound.Info().Msgf("Inbound tracker watcher started for chain %d", ob.Chain().Id)
	for {
		select {
		case <-ticker.C():
			if !app.PellCoreContext().IsInboundObservationEnabled(ob.GetChainParams()) {
				continue
			}
			err := ob.ObserveIntxTrackers(ctx)
			if err != nil {
				ob.Logger().Inbound.Err(err).Msg("ProcessInboundTrackers error")
			}
			ticker.UpdateInterval(ob.GetChainParams().InTxTicker, ob.Logger().Inbound)
		case <-ob.StopChannel():
			ob.Logger().Inbound.Info().Msgf("WatchInboundTracker stopped for chain %d", ob.Chain().Id)
			return nil
		}
	}
}

// ObserveIntxTrackers observes the inbound trackers for the chain
func (ob *ChainClient) ObserveIntxTrackers(ctx context.Context) error {
	trackers, err := ob.PellcoreClient().GetInboundTrackersForChain(ctx, ob.Chain().Id)
	if err != nil {
		return err
	}

	for _, tracker := range trackers {
		// query tx and receipt
		tx, _, err := ob.TransactionByHash(tracker.TxHash)
		if err != nil {
			return errors.Wrapf(
				err,
				"error getting transaction for inbound %s chain %d",
				tracker.TxHash,
				ob.Chain().Id)
		}

		receipt, err := ob.evmClient.TransactionReceipt(context.Background(), ethcommon.HexToHash(tracker.TxHash))
		if err != nil {
			return errors.Wrapf(
				err,
				"error getting receipt for inbound %s chain %d",
				tracker.TxHash,
				ob.Chain().Id,
			)
		}

		ob.Logger().Inbound.Info().Msgf("checking tracker for intx %s chain %d", tracker.TxHash, ob.Chain().Id)

		// check and build inbound vote msg
		msgs, err := ob.evmEventReactor.CheckAndBuildInboundVoteMsg(tx, receipt, ob.LastBlock())
		if err != nil {
			return errors.Wrapf(err, "error checking and building for intx %s chain %d", tx.Hash, ob.Chain().Id)
		}
		// post vote msg
		_, err = ob.PostVoteInboundMsgs(ctx, msgs)
		if err != nil {
			return errors.Wrapf(err, "error posting vote msg for intx %s chain %d", tx.Hash, ob.Chain().Id)
		}
	}
	return nil
}

func (ob *ChainClient) observeInTX(ctx context.Context, sampledLogger zerolog.Logger) error {
	// get and update latest block height
	blockNumber, err := ob.evmClient.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	if blockNumber < ob.LastBlock() {
		return fmt.Errorf("observeInTX: block number should not decrease: current %d last %d", blockNumber, ob.LastBlock())
	}

	ob.WithLastBlock(blockNumber)

	// increment prom counter
	metrics.GetBlockByNumberPerChain.WithLabelValues(ob.Chain().ChainName()).Inc()

	// skip if current height is too low
	if blockNumber < ob.GetChainParams().ConfirmationCount {
		return fmt.Errorf("observeInTX: skipping observer, current block number %d is too low", blockNumber)
	}
	confirmedBlockNum := blockNumber - ob.GetChainParams().ConfirmationCount

	// skip if no new block is confirmed
	lastScanned := ob.LastBlockScanned()
	if lastScanned >= confirmedBlockNum {
		sampledLogger.Info().Msgf("observeInTX: skipping observer, no new block is produced for chain %d", ob.Chain().Id)
		return nil
	}

	// get last scanned block height (we simply use same height for all 3 events PellSent, Deposited, TssRecvd)
	// Note: using different heights for each event incurs more complexity (metrics, db, etc) and not worth it
	startBlock, toBlock := ob.calcBlockRangeToScan(confirmedBlockNum, lastScanned, config.DefaultBlocksPerPeriod)

	// task : query evm chain for pell logs (read at most 100 blocks in on go)
	lastScannedBlock := ob.ObservePellEvents(ctx, startBlock, toBlock)

	// update last scanned block height for only one event (PellSent), ignore db error
	if lastScannedBlock > lastScanned {
		sampledLogger.Info().Msgf("observeInTX: lastScanned PellEvents heights %d", lastScannedBlock)

		ob.WithLastBlockScanned(lastScannedBlock)
		metrics.LastScannedBlockNumber.WithLabelValues(ob.Chain().ChainName()).Set(float64(lastScannedBlock))

		if err := ob.SaveScanInfo(lastScannedBlock, ob.LastInboundBlock()); err != nil {
			ob.Logger().Inbound.Error().Err(err).Msgf("observeInTX: error writing lastScannedLowest %d to db", lastScannedBlock)
		}
		return nil
	}

	return nil
}

// ObservePellEvents queries logs from registered providers and processes them
func (ob *ChainClient) ObservePellEvents(ctx context.Context, startBlock, toBlock uint64) uint64 {
	chainIndex, err := ob.PellcoreClient().GetChainIndex(ctx, ob.Chain().Id)
	if err != nil {
		ob.Logger().Inbound.Error().
			Err(err).
			Uint64("startBlock", startBlock).
			Msg("ObservePellEvents: failed to get chain index")
		return startBlock
	}

	// When the scanning start block height (startBlock) exceeds the current indexed height (chainIndex.CurrHeight)
	// by a certain gap (MaxLatestIndexedBlockGap), it indicates that the relayer's scanning has far exceeded
	// the chain's indexing progress. Continuing to scan may result in invalid votes.
	// Therefore, we return startBlock to stop the relayer's scanning until the chain's index catches up.
	if startBlock > ob.maxLatestIndexedBlockGap+chainIndex.CurrHeight {
		ob.Logger().Inbound.Warn().
			Uint64("start_block", startBlock).
			Uint64("curr_height", chainIndex.CurrHeight).
			Uint64("max_gap", ob.maxLatestIndexedBlockGap).
			Msg("ObservePellEvents: block height gap too large")

		// If no Pell Events are generated within MaxLatestIndexedBlockGap, it's normal for the scanning height to exceed the chain's index height.
		// Therefore, only return startBlock to stop scanning when LastInboundBlock (the last height where Pell Events were scanned) is greater than the index height.
		if ob.LastInboundBlock() > chainIndex.CurrHeight {
			return startBlock
		}
	}

	// process all registered event handlers
	eventStore, lastScannedBlock := ob.evmEventReactor.HandleBlocks(startBlock, toBlock)

	ob.Logger().Inbound.Info().
		Int64("start_block", int64(startBlock)).
		Int64("to_block", int64(toBlock)).
		Int("num_events", len(eventStore)).
		Int64("last_scanned_block", int64(lastScannedBlock)).
		Msgf("ObservePellEvents: scan blocks %d to %d detected %d events", startBlock, toBlock, len(eventStore))
	metrics.LastInboundBlockNumber.WithLabelValues(ob.Chain().ChainName()).Set(float64(lastScannedBlock))

	// vote on all events
	if len(eventStore) > 0 {
		if err := ob.voteEvents(ctx, eventStore); err != nil {
			ob.Logger().Inbound.Err(err).Msg("ObservePellEvents: error voting on events")

			return startBlock
		}
	}
	// update last scanned block height
	ob.WithLastBlock(lastScannedBlock)
	metrics.LastVoteInboundBlockNumber.WithLabelValues(ob.Chain().ChainName()).Set(float64(lastScannedBlock))

	return lastScannedBlock
}

// build vote msg from multiple contract events
// stakerDeposit/delegate
func (ob *ChainClient) voteEvents(ctx context.Context, eventStore map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) error {
	heights := make([]uint64, 0)
	for height := range eventStore {
		heights = append(heights, height)
	}

	sort.SliceStable(heights, func(i, j int) bool {
		return heights[i] < heights[j]
	})

	for _, height := range heights {
		blockProof := xmsgtypes.BlockProof{
			ChainId:         uint64(ob.Chain().Id),
			PrevBlockHeight: ob.LastInboundBlock(),
			BlockHeight:     height,
		}

		events := eventStore[height]
		sort.SliceStable(events, func(i, j int) bool {
			return events[i].EventIndex < events[j].EventIndex
		})

		blockProof.Events = make([]*xmsgtypes.Event, len(events))
		for i, event := range events {
			blockProof.Events[i] = &xmsgtypes.Event{
				Index:     event.EventIndex,
				TxHash:    event.InTxHash,
				PellEvent: event.PellTx,
				Digest:    event.Digest(),
			}
		}

		ob.Logger().Inbound.Info().
			Int64("block_height", int64(blockProof.BlockHeight)).
			Msg("voteEvents: voting on block")

		block := &xmsgtypes.MsgVoteInboundBlock{
			Signer:     ob.PellcoreClient().GetKeys().GetOperatorAddress().String(),
			BlockProof: &blockProof,
		}

		if _, _, err := ob.PellcoreClient().PostVoteInboundBlock(ctx, pellcore.PostVoteInboundGasLimit, pellcore.PostVoteInboundGasLimit*2, block, events); err != nil {
			return err
		}

		ob.WithInboundBlock(blockProof.BlockHeight)
	}

	return nil
}

// PostVoteInboundMsgs posts multiple inbound vote messages to pellcore
func (ob *ChainClient) PostVoteInboundMsgs(ctx context.Context, msgs []*xmsgtypes.MsgVoteOnObservedInboundTx) (string, error) {
	var ret string
	var err error
	for _, msg := range msgs {
		ret, err = ob.PostVoteInbound(ctx, msg, pellcore.PostVoteInboundExecutionGasLimit)
		if err != nil {
			return "", err
		}
	}
	return ret, nil
}

// HasEnoughConfirmations checks if the given receipt has enough confirmations
func (ob *ChainClient) HasEnoughConfirmations(receipt *ethtypes.Receipt, lastHeight uint64) bool {
	confHeight := receipt.BlockNumber.Uint64() + ob.GetChainParams().ConfirmationCount
	return lastHeight >= confHeight
}

// calcBlockRangeToScan calculates the next range of blocks to scan
func (ob *ChainClient) calcBlockRangeToScan(latestConfirmed, lastScanned, batchSize uint64) (uint64, uint64) {
	startBlock := lastScanned + 1
	toBlock := lastScanned + batchSize
	if toBlock > latestConfirmed {
		toBlock = latestConfirmed
	}
	return startBlock, toBlock
}
