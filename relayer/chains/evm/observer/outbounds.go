package observer

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/chains/interfaces"
	pctx "github.com/pell-chain/pellcore/relayer/context"
	"github.com/pell-chain/pellcore/relayer/logs"
	clienttypes "github.com/pell-chain/pellcore/relayer/types"
	xmsgkeeper "github.com/pell-chain/pellcore/x/xmsg/keeper"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// WatchOutbound watches evm chain for outgoing txs status
// TODO(revamp): move ticker function to ticker file
func (ob *ChainClient) WatchOutTx(ctx context.Context) error {
	// get app context
	app, err := pctx.FromContext(ctx)
	if err != nil {
		return err
	}

	ticker, err := clienttypes.NewDynamicTicker(
		fmt.Sprintf("EVM_WatchOutTx_%d", ob.Chain().Id),
		ob.GetChainParams().OutTxTicker,
	)
	if err != nil {
		ob.Logger().Outbound.Error().Err(err).Msg("error creating ticker")
		return err
	}

	ob.Logger().Outbound.Info().Msgf("WatchOutTx started for chain %d", ob.Chain().Id)
	sampledLogger := ob.Logger().Outbound.Sample(&zerolog.BasicSampler{N: 10})
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			if !app.PellCoreContext().IsOutboundObservationEnabled(ob.GetChainParams()) {
				sampledLogger.Info().Msgf("WatchOutTx: outbound observation is disabled for chain %d", ob.Chain().Id)
				continue
			}

			// process outbound trackers
			err := ob.ProcessOutboundTrackers(ctx)
			if err != nil {
				ob.Logger().
					Outbound.Error().
					Err(err).
					Msgf("WatchOutbound: error ProcessOutboundTrackers for chain %d", ob.Chain().Id)
			}

			ticker.UpdateInterval(ob.GetChainParams().OutTxTicker, ob.Logger().Outbound)
		case <-ob.StopChannel():
			ob.Logger().Outbound.Info().Msg("WatchOutTx: stopped")
			return nil
		}
	}
}

// ProcessOutboundTrackers processes outbound trackers
func (ob *ChainClient) ProcessOutboundTrackers(ctx context.Context) error {
	chainID := ob.Chain().Id
	trackers, err := ob.PellcoreClient().GetAllOutTxTrackerByChain(ctx, ob.Chain().Id, interfaces.Ascending)
	if err != nil {
		return errors.Wrap(err, "GetAllOutboundTrackerByChain error")
	}

	// prepare logger fields
	logger := ob.Logger().Outbound.With().
		Str(logs.FieldMethod, "ProcessOutboundTrackers").
		Int64(logs.FieldChain, chainID).
		Logger()

	// process outbound trackers
	for _, tracker := range trackers {
		// go to next tracker if this one already has a confirmed tx
		nonce := tracker.Nonce
		logger.Info().Msgf("processing %d outbound trackers for chain %d nonce %d", len(trackers), chainID, nonce)
		if ob.IsTxConfirmed(nonce) { // Go to next tracker if this one already has a confirmed tx
			continue
		}

		logger.Info().Msgf("processing %d outbound trackers for chain %d nonce %d not confirmed", len(trackers), chainID, nonce)

		// check each txHash and save tx and receipt if it's legit and confirmed
		txCount := 0
		var outtxReceipt *ethtypes.Receipt
		var outtx *ethtypes.Transaction
		for _, txHash := range tracker.HashLists {
			if receipt, tx, ok := ob.checkConfirmedTx(txHash.TxHash, nonce); ok {
				txCount++
				outtxReceipt = receipt
				outtx = tx
				logger.Info().Msgf("confirmed outbound %s for chain %d nonce %d", txHash.TxHash, chainID, nonce)
				if txCount > 1 {
					logger.Error().
						Msgf("checkConfirmedTx passed, txCount %d chain %d nonce %d receipt %v tx %v", txCount, chainID, nonce, receipt, tx)
				}
			} else {
				logger.Info().Msgf("not confirmed outbound %s for chain %d nonce %d", txHash.TxHash, chainID, nonce)
			}
		}

		// should be only one txHash confirmed for each nonce.
		if txCount == 1 { // should be only one txHash confirmed for each nonce.
			ob.SetTxNReceipt(nonce, outtxReceipt, outtx)
		} else if txCount > 1 { // should not happen. We can't tell which txHash is true. It might happen (e.g. glitchy/hacked endpoint)
			// should not happen. We can't tell which txHash is true. It might happen (e.g. bug, glitchy/hacked endpoint)
			ob.Logger().Outbound.Error().Msgf("WatchOutbound: confirmed multiple (%d) outbound for chain %d nonce %d", txCount, chainID, nonce)
		} else {
			if len(tracker.HashLists) == xmsgkeeper.MaxOutTxTrackerHashes {
				ob.Logger().Outbound.Error().Msgf("WatchOutbound: outbound tracker is full of hashes for chain %d nonce %d", chainID, nonce)
			}
		}
	}

	return nil
}

// PostVoteOutbound posts vote to pellcore for the confirmed outtx
func (ob *ChainClient) PostVoteOutbound(
	ctx context.Context,
	xmsgIndex string,
	receipt *ethtypes.Receipt,
	transaction *ethtypes.Transaction,
	receiveStatus chains.ReceiveStatus,
	nonce uint64,
	logger zerolog.Logger,
) {
	chainID := ob.Chain().Id
	failedReasonMsg := ob.filterOutTxFailedReasonMsg(receipt)
	if failedReasonMsg != "" {
		receiveStatus = chains.ReceiveStatus_FAILED
	}

	pellTxHash, ballot, err := ob.PellcoreClient().PostVoteOutbound(
		ctx,
		xmsgIndex,
		receipt.TxHash.Hex(),
		receipt.BlockNumber.Uint64(),
		receipt.GasUsed,
		transaction.GasPrice(),
		transaction.Gas(),
		receiveStatus,
		failedReasonMsg,
		ob.Chain(),
		nonce,
	)
	if err != nil {
		logger.Error().Err(err).Msgf("PostVoteOutbound: error posting vote for chain %d nonce %d outtx %s ", chainID, nonce, receipt.TxHash)
	} else if pellTxHash != "" {
		logger.Info().Msgf("PostVoteOutbound: posted vote for chain %d nonce %d outtx %s vote %s ballot %s", chainID, nonce, receipt.TxHash, pellTxHash, ballot)
	}
}

// filterOutTxFailedEvents filters out tx failed events from receipt.
// PellMessageFailed
func (ob *ChainClient) filterOutTxFailedReasonMsg(receipt *ethtypes.Receipt) string {
	res := ""
	for _, log := range receipt.Logs {
		if log.Address.String() != ob.ChainParams().ConnectorContractAddress {
			continue
		}

		if len(log.Topics) > 0 && log.Topics[0] != ConnectorContractABI.Events[PellMessageFailedEventName].ID {
			continue
		}

		pellMessageFailedEvent := new(pellconnector.PellConnectorPellMessageFailed)
		if err := ConnectorContractABI.UnpackIntoInterface(pellMessageFailedEvent, PellMessageFailedEventName, log.Data); err != nil {
			continue
		}

		if res != "" {
			res += "-"
		}

		reason := hex.EncodeToString(pellMessageFailedEvent.Reason)
		if reason == "" {
			reason = "unknownReason"
		}

		res += reason
	}

	ob.Logger().Outbound.Info().Msgf("filterOutTxFailedReasonMsg: res %s", res)
	return res
}

// IsOutboundProcessed checks outtx status and returns (isIncluded, isConfirmed, error)
// It also posts vote to pellcore if the tx is confirmed
func (ob *ChainClient) IsOutboundProcessed(ctx context.Context, xmsg *xmsgtypes.Xmsg, logger zerolog.Logger) (bool, bool, error) {
	if xmsg.IsCrossChainPellTx() {
		return ob.IsOutboundProcessed_i(ctx, xmsg, logger)
	} else {
		return false, false, fmt.Errorf("IsOutboundProcessed: don't support this xmsg %s", xmsg.Index)
	}
}

// IsOutboundProcessed checks outtx status and returns (isIncluded, isConfirmed, error)
// It also posts vote to pellcore if the tx is confirmed
func (ob *ChainClient) IsOutboundProcessed_i(ctx context.Context, xmsg *xmsgtypes.Xmsg, logger zerolog.Logger) (bool, bool, error) {
	// skip if outtx is not confirmed
	nonce := xmsg.GetCurrentOutTxParam().OutboundTxTssNonce
	if !ob.IsTxConfirmed(nonce) {
		return false, false, nil
	}
	receipt, transaction := ob.GetTxNReceipt(nonce)
	sendID := fmt.Sprintf("%d-%d", ob.Chain().Id, nonce)
	logger = logger.With().Str("sendID", sendID).Logger()

	// define a few common variables
	receiveStatus := chains.ReceiveStatus_FAILED
	if receipt.Status == ethtypes.ReceiptStatusSuccessful {
		receiveStatus = chains.ReceiveStatus_SUCCESS
	}

	// post vote to pellcore
	ob.PostVoteOutbound(ctx, xmsg.Index, receipt, transaction, receiveStatus, nonce, logger)
	return true, true, nil
}

// checkConfirmedTx checks if a txHash is confirmed
// returns (receipt, transaction, true) if confirmed or (nil, nil, false) otherwise
func (ob *ChainClient) checkConfirmedTx(txHash string, nonce uint64) (*ethtypes.Receipt, *ethtypes.Transaction, bool) {
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errLog := log.Error().
		Uint64("nonce", nonce).
		Int64("chain", ob.Chain().Id).
		Str("txHash", txHash)

	// query transaction
	transaction, isPending, err := ob.evmClient.TransactionByHash(ctxt, ethcommon.HexToHash(txHash))
	if err != nil {
		errLog.Err(err).Msgf("confirmTxByHash: error getting transaction for outtx %s nonce %d chain %d", txHash, nonce, ob.Chain().Id)
		return nil, nil, false
	}
	if transaction == nil { // should not happen
		errLog.Msgf("confirmTxByHash: transaction is nil for outtx %s nonce %d chain %d", txHash, nonce, ob.Chain().Id)
		return nil, nil, false
	}

	// check tx sender and nonce
	signer := ethtypes.NewLondonSigner(big.NewInt(ob.Chain().Id))
	from, err := signer.Sender(transaction)
	if err != nil {
		log.Error().Err(err).Msgf("confirmTxByHash: local recovery of sender address failed for outtx %s chain %d", transaction.Hash().Hex(), ob.Chain().Id)
		return nil, nil, false
	}
	if from != ob.TSS().EVMAddress() { // must be TSS address
		errLog.Msgf("confirmTxByHash: sender %s for outtx %s nonce %d chain %d is not TSS address %s",
			from.Hex(), transaction.Hash().Hex(), nonce, ob.Chain().Id, ob.TSS().EVMAddress().Hex())
		return nil, nil, false
	}
	if transaction.Nonce() != nonce { // must match xmsg nonce
		errLog.Msgf("confirmTxByHash: outtx %s nonce mismatch: wanted %d, got tx nonce %d chain %d", txHash, nonce, transaction.Nonce(), ob.Chain().Id)
		return nil, nil, false
	}

	// save pending transaction
	if isPending {
		ob.SetPendingTx(nonce, transaction)
		log.Info().Msgf("confirmTxByHash: outtx %s nonce %d is pending for chain %d", txHash, nonce, ob.Chain().Id)
		return nil, nil, false
	}

	// query receipt
	receipt, err := ob.evmClient.TransactionReceipt(ctxt, ethcommon.HexToHash(txHash))
	if err != nil {
		if err != ethereum.NotFound {
			log.Warn().Err(err).Msgf("confirmTxByHash: TransactionReceipt error, txHash %s nonce %d chain %d", txHash, nonce, ob.Chain().Id)
		}
		return nil, nil, false
	}
	if receipt == nil { // should not happen
		errLog.Msgf("confirmTxByHash: receipt is nil for outtx %s nonce %d chain %d", txHash, nonce, ob.Chain().Id)
		return nil, nil, false
	}

	// check confirmations
	if !ob.HasEnoughConfirmations(receipt, ob.LastBlock()) {
		log.Info().Msgf("confirmTxByHash: txHash %s nonce %d included but not confirmed: receipt block %d, current block %d",
			txHash, nonce, receipt.BlockNumber, ob.LastBlock())
		return nil, nil, false
	}

	// cross-check tx inclusion against the block
	// Note: a guard for false BlockNumber in receipt. The blob-carrying tx won't come here
	err = ob.CheckTxInclusion(transaction, receipt)
	if err != nil {
		errLog.Err(err).Msgf("confirmTxByHash: checkTxInclusion error for outtx %s nonce %d chain %d", txHash, nonce, ob.Chain().Id)
		return nil, nil, false
	}

	return receipt, transaction, true
}
