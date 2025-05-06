package keeper

import (
	"encoding/base64"
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

/* ProcessSuccessfulOutbound processes a successful outbound transaction. It does the following things in one function:

	1. Change the status of the Xmsg from
	 - PendingRevert to Reverted
     - PendingOutbound to OutboundMined

	2. Set the finalization status of the current outbound tx to executed

	3. Emit an event for the successful outbound transaction
*/

// This function sets Xmsg status , in cases where the outbound tx is successful, but tx itself fails
// This is done because SaveSuccessfulOutbound does not set the xmsg status
// For cases where the outbound tx is unsuccessful, the xmsg status is automatically set to Aborted in the ProcessFailedOutbound function, so we can just return and error to trigger that

func (k Keeper) ProcessSuccessfulOutbound(ctx sdk.Context, xmsg *types.Xmsg) {
	oldStatus := xmsg.XmsgStatus.Status
	switch oldStatus {
	case types.XmsgStatus_PENDING_REVERT:
		xmsg.SetReverted("Outbound succeeded, revert executed")
	case types.XmsgStatus_PENDING_OUTBOUND:
		xmsg.SetOutBoundMined("Outbound succeeded, mined")
	default:
		return
	}
	xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
	newStatus := xmsg.XmsgStatus.Status.String()
	EmitOutboundSuccess(ctx, oldStatus.String(), newStatus, xmsg.Index)
}

/*
ProcessFailedOutbound processes a failed outbound transaction. It does the following things in one function:

 1. For Admin Tx or a withdrawal from Pell chain, it aborts the Xmsg

 2. For other Xmsg
    - If the Xmsg is in PendingOutbound, it creates a revert tx and sets the finalization status of the current outbound tx to executed
    - If the Xmsg is in PendingRevert, it sets the Status to Aborted

 3. Emit an event for the failed outbound transaction

 4. Set the finalization status of the current outbound tx to executed. If a revert tx is is created, the finalization status is not set, it would get set when the revert is processed via a subsequent transaction
*/

// This function sets Xmsg status , in cases where the outbound tx is successful, but tx itself fails
// This is done because SaveSuccessfulOutbound does not set the xmsg status
// For cases where the outbound tx is unsuccessful, the xmsg status is automatically set to Aborted in the ProcessFailedOutbound function, so we can just return and error to trigger that
func (k Keeper) ProcessFailedOutbound(ctx sdk.Context, xmsg *types.Xmsg) error {
	oldStatus := xmsg.XmsgStatus.Status
	// The following logic is used to handler the mentioned conditions separately. The reason being
	// All admin tx is created using a policy message , there is no associated inbound tx , therefore we do not need any revert logic
	// For transactions which originated from PEVM , we can process the outbound in the same block as there is no TSS signing required for the revert
	// For all other transactions we need to create a revert tx and set the status to pending revert

	if chains.IsPellChain(xmsg.InboundTxParams.SenderChainId) {
		pellSent := xmsg.InboundTxParams.InboundPellTx.GetPellSent()
		if pellSent == nil {
			return fmt.Errorf("no pell data to be sent")
		}

		paramType, err := pevmtypes.PellSentParamTypeFromString(pellSent.PellParams)
		if err != nil {
			return fmt.Errorf("failed to convert PellSentParamType: %s", err.Error())
		}

		switch paramType {
		case pevmtypes.ReceiveCall:
			xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
			xmsg.SetAbort("Outbound failed from pell chain")
		case pevmtypes.RevertableCall:
			if err := k.processFailedOutboundForPEVMTx(ctx, xmsg); err != nil {
				return cosmoserrors.Wrap(err, "ProcessFailedOutboundForPEVMTx")
			}
		case pevmtypes.Transfer:
			xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
			xmsg.SetAbort("Outbound failed from pell chain with transfer pell params")
		default:
			// TODO: Remove this block in the next release
			// This block is maintained for backward compatibility.
			// Previously, Pell parameters were Base64-encoded.
			// Decode the Base64-encoded PellParams to handle legacy data.
			decodedData, err := base64.StdEncoding.DecodeString(pellSent.PellParams)
			if err != nil {
				return fmt.Errorf("failed to decode pellSent.PellParams: %s", err.Error())
			}
			// Ensure the decoded data is not empty.
			if len(decodedData) == 0 {
				return fmt.Errorf("invalid pellSent.PellParams: %s", pellSent.PellParams)
			}
			// The last byte of decoded data determines the type of transaction to process.
			lastByte := decodedData[len(decodedData)-1]
			switch lastByte {
			case 0:
				xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
				xmsg.SetAbort("Outbound failed from pell chain")
			case 1:
				if err := k.processFailedOutboundForPEVMTx(ctx, xmsg); err != nil {
					return cosmoserrors.Wrap(err, "ProcessFailedOutboundForPEVMTx")
				}
			case 2:
				xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
				xmsg.SetAbort("Outbound failed from pell chain with transfer pell params")
			}

			xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
			xmsg.SetAbort("Outbound failed from pell chain with unknown pell params")
		}
	} else {
		xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
		xmsg.SetAbort("Outbound failed from other chain")
	}

	newStatus := xmsg.XmsgStatus.Status.String()
	EmitOutboundFailure(ctx, oldStatus.String(), newStatus, xmsg.Index)
	return nil
}

func (k Keeper) processFailedOutboundForPEVMTx(ctx sdk.Context, xmsg *types.Xmsg) error {
	indexBytes, err := xmsg.GetXmsgIndicesBytes()
	if err != nil {
		// Return err to save the failed outbound ad set to aborted
		return fmt.Errorf("failed reverting GetXmsgIndicesBytes: %s", err.Error())
	}
	// Finalize the older outbound tx
	xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED

	// create new OutboundTxParams for the revert. We use the fixed gas limit for revert when calling pEVM
	err = xmsg.AddRevertOutbound(pevmtypes.PEVMGasLimit.Uint64())
	if err != nil {
		// Return err to save the failed outbound ad set to aborted
		return fmt.Errorf("failed AddRevertOutbound: %s", err.Error())
	}

	// Trying to revert the transaction this would get set to a finalized status in the same block as this does not need a TSS singing
	xmsg.SetPendingRevert("Outbound failed, trying revert")

	// Fetch the original sender and receiver from the Xmsg , since this is a revert the sender with be the receiver in the new tx
	originalSender := ethcommon.HexToAddress(xmsg.InboundTxParams.Sender)
	// This transaction will always have two outbounds, the following logic is just an added precaution.
	// The contract call or token deposit would go the original sender.
	originalReceiver := ethcommon.HexToAddress(xmsg.GetCurrentOutTxParam().Receiver)
	orginalReceiverChainID := xmsg.GetCurrentOutTxParam().ReceiverChainId
	if len(xmsg.OutboundTxParams) == 2 {
		// If there are 2 outbound tx, then the original receiver is the receiver in the first outbound tx
		originalReceiver = ethcommon.HexToAddress(xmsg.OutboundTxParams[0].Receiver)
		orginalReceiverChainID = xmsg.OutboundTxParams[0].ReceiverChainId
	}
	// Call evm to revert the transaction
	_, err = k.pevmKeeper.PELLRevertAndCallContract(ctx,
		originalSender,
		originalReceiver,
		xmsg.InboundTxParams.SenderChainId,
		orginalReceiverChainID, indexBytes)
	// If revert fails, we set it to abort directly there is no way to refund here as the revert failed
	if err != nil {
		return fmt.Errorf("failed PELLRevertAndCallContract: %s", err.Error())
	}
	xmsg.SetReverted("Outbound failed, revert executed")
	if len(ctx.TxBytes()) > 0 {
		// add event for tendermint transaction hash format
		hash := tmbytes.HexBytes(tmtypes.Tx(ctx.TxBytes()).Hash())
		ethTxHash := ethcommon.BytesToHash(hash)
		xmsg.GetCurrentOutTxParam().OutboundTxHash = ethTxHash.String()
		// #nosec G701 always positive
		xmsg.GetCurrentOutTxParam().OutboundTxExternalHeight = uint64(ctx.BlockHeight())
	}
	xmsg.GetCurrentOutTxParam().TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
	return nil
}

// ProcessOutbound processes the finalization of an outbound transaction based on the ballot status
// The state is committed only if the individual steps are successful
func (k Keeper) ProcessOutbound(ctx sdk.Context, xmsg *types.Xmsg, ballotStatus relayertypes.BallotStatus) error {
	tmpCtx, commit := ctx.CacheContext()
	err := func() error {
		switch ballotStatus {
		case relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION:
			k.ProcessSuccessfulOutbound(tmpCtx, xmsg)
		case relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION:
			err := k.ProcessFailedOutbound(tmpCtx, xmsg)
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}
	err = xmsg.Validate()
	if err != nil {
		return err
	}
	commit()
	return nil
}

// processXmsgOutboundResult processes the outbound result of a xmsg. XmsgOutboundResultHook hooks are called here
func (k Keeper) processXmsgOutboundResult(ctx sdk.Context, xmsg *types.Xmsg, ballotStatus relayertypes.BallotStatus) error {
	for _, hook := range k.xmsgResultHooks {
		hook.ProcessXmsgOutboundResult(ctx, xmsg, ballotStatus)
	}

	return nil
}
