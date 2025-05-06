// Package signer implements the ChainSigner interface for EVM chains
package signer

import (
	"context"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"

	"github.com/0xPellNetwork/aegis/relayer/chains/evm"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/logs"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
)

// reportToOutTxTracker reports outTxHash to tracker only when tx receipt is available
func (signer *Signer) reportToOutTxTracker(
	ctx context.Context,
	pellBridge interfaces.PellCoreBridger,
	chainID int64,
	nonce uint64,
	outTxHash string,
	logger zerolog.Logger) {
	logger = logger.With().
		Str(logs.FieldMethod, "reportToOutboundTracker").
		Int64(logs.FieldChain, chainID).
		Uint64(logs.FieldNonce, nonce).
		Str(logs.FieldTx, outTxHash).
		Logger()

	// skip if already being reported
	// set being reported flag to avoid duplicate reporting
	alreadySet := signer.SetBeingReportedFlag(outTxHash)
	if alreadySet {
		logger.Info().Msg("outbound is being reported to tracker")
		return
	}

	// report to outTx tracker with goroutine
	// launch a goroutine to monitor tx confirmation status
	go func() {
		defer func() {
			signer.ClearBeingReportedFlag(outTxHash)
		}()

		// try monitoring tx inclusion status for 10 minutes
		var err error
		report := false
		isPending := false
		blockNumber := uint64(0)

		tStart := time.Now()
		for {
			// take a rest between each check
			time.Sleep(10 * time.Second)

			// give up (forget about the tx) after 20 minutes of monitoring, the reasons are:
			// 1. the gas stability pool should have kicked in and replaced the tx by then.
			// 2. even if there is a chance that the tx is included later, more likely it's going to be a false tx hash (either replaced or dropped).
			// 3. we prefer missed tx hash over potentially invalid tx hash.
			if time.Since(tStart) > evm.OutTxInclusionTimeout {
				logger.Info().Msg("timeout waiting tx inclusion")
				break
			}
			// try getting the tx
			_, isPending, err = signer.client.TransactionByHash(context.TODO(), ethcommon.HexToHash(outTxHash))
			if err != nil {
				logger.Info().Err(err).Msg("error getting tx")
				continue
			}
			// if tx is include in a block, try getting receipt
			if !isPending {
				report = true // included
				receipt, err := signer.client.TransactionReceipt(context.TODO(), ethcommon.HexToHash(outTxHash))
				if err != nil {
					logger.Info().Err(err).Msg("error getting receipt")
				}
				if receipt != nil {
					blockNumber = receipt.BlockNumber.Uint64()
				}
				break
			}
			// keep monitoring pending tx
			logger.Info().Msg("tx has not been included yet")
		}

		// try adding to outTx tracker for 10 minutes
		if report {
			tStart := time.Now()
			for {
				// give up after 10 minutes of retrying
				if time.Since(tStart) > evm.OutTxTrackerReportTimeout {
					logger.Info().Msg("timeout adding outtx tracker, please add manually")
					break
				}
				// stop if the xmsg is already finalized
				xmsg, err := pellBridge.GetXmsgByNonce(ctx, chainID, nonce)
				if err != nil {
					logger.Err(err).Msg("error getting xmsg")
				} else if !xmsgkeeper.IsPending(xmsg) {
					logger = logger.With().Str("xmsgIndex", xmsg.Index).Logger()
					logger.Info().Msg("xmsg already finalized")
					break
				}
				// report to outTx tracker
				pellHash, err := pellBridge.PostAddTxHashToOutTxTracker(ctx, chainID, nonce, outTxHash, nil, "", -1)
				if err != nil {
					logger.Err(err).Msg("error adding to outtx tracker")
				} else if pellHash != "" {
					logger.Info().Msgf("added outTxHash to core successful %s block %d", pellHash, blockNumber)
				} else {
					// stop if the tracker contains the outTxHash
					logger.Info().Msg("outtx tracker contains outTxHash")
					break
				}
				// retry otherwise
				time.Sleep(evm.PellBlockTime * 3)
			}
		}
	}()
}
