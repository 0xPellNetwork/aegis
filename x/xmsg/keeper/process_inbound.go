package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// ProcessInbound processes the inbound Xmsg.
// It does a conditional dispatch to ProcessPEVMDeposit or ProcessCrosschainMsgPassing based on the receiver chain.
// The internal functions handle the state changes and error handling.
func (k Keeper) ProcessInbound(ctx sdk.Context, xmsg *types.Xmsg) {
	if chains.IsPellChain(xmsg.GetCurrentOutTxParam().ReceiverChainId) {
		if xmsg.IsCrossChainPellTx() {
			k.processPEVMEvents(ctx, xmsg)
			ctx.Logger().Debug("ProcessInbound: processPEVMEvents completed for xmsg ", xmsg.Index)
		} else {
			xmsg.SetAbort(fmt.Sprintf("invalid xmsg[%s]", xmsg.Index))
			ctx.Logger().Debug("ProcessInbound: invalid xmsg ", xmsg.Index)
		}
	} else {
		xmsg.SetAbort(fmt.Sprintf("invalid receiver chainID %d", xmsg.GetCurrentOutTxParam().ReceiverChainId))
	}
}

/*
processPEVMEvents processes the EVM events Xmsg. Every event is a xmsg which has Pellchain as the receiver chain.It trasnsitions state according to the following rules:
  - If the deposit is successful, the Xmsg status is changed to OutboundMined.
  - If the deposit returns an internal error i.e if HandleEVMDeposit() returns an error, but isContractReverted is false, the Xmsg status is changed to Aborted.
  - If the deposit is reverted, the function tries to create a revert xmsg with status PendingRevert.
  - If the creation of revert tx also fails it changes the status to Aborted.

Note : Aborted Xmsgs are not refunded in this function. The refund is done using a separate refunding mechanism.
We do not return an error from this function , as all changes need to be persisted to the state.
Instead we use a temporary context to make changes and then commit the context on for the happy path ,i.e xmsg is set to OutboundMined.
*/
func (k Keeper) processPEVMEvents(ctx sdk.Context, xmsg *types.Xmsg) {
	tmpCtx, commit := ctx.CacheContext()
	isContractReverted, err := k.HandleEVMEvents(tmpCtx, xmsg)

	if err != nil { // exceptional case; internal error; should abort Xmsg
		if !isContractReverted {
			xmsg.SetAbort(err.Error())
			return
		}
		// TODO: set pending reverted
		xmsg.SetAbort(err.Error())
		return

		// contract call reverted; should refund via a revert tx
		//revertMessage := err.Error()
		//senderChain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, xmsg.InboundTxParams.SenderChainId)
		//if senderChain == nil {
		//	xmsg.SetAbort(fmt.Sprintf("invalid sender chain id %d", xmsg.InboundTxParams.SenderChainId))
		//	return
		//}
		//gasLimit, err := k.GetRevertGasLimit(ctx, *xmsg)
		//if err != nil {
		//	xmsg.SetAbort(fmt.Sprintf("revert gas limit error: %s", err.Error()))
		//	return
		//}
		//
		//if gasLimit == 0 {
		//	// use same gas limit of outbound as a fallback -- should not be required
		//	gasLimit = xmsg.GetCurrentOutTxParam().OutboundTxGasLimit
		//}
		//
		//err = xmsg.AddRevertOutbound(gasLimit)
		//if err != nil {
		//	xmsg.SetAbort(fmt.Sprintf("revert outbound error: %s", err.Error()))
		//	return
		//}
		//// we create a new cached context, and we don't commit the previous one with EVM deposit
		//tmpCtxRevert, commitRevert := ctx.CacheContext()
		//err = func() error {
		//	err := k.PayGasAndUpdateXmsg(
		//		tmpCtxRevert,
		//		senderChain.Id,
		//		xmsg,
		//	)
		//	if err != nil {
		//		return err
		//	}
		//	// Update nonce using senderchain id as this is a revert tx and would go back to the original sender
		//	return k.UpdateNonce(tmpCtxRevert, senderChain.Id, xmsg)
		//}()
		//if err != nil {
		//	xmsg.SetAbort(fmt.Sprintf("processInbound revert message: %s err : %s", revertMessage, err.Error()))
		//	return
		//}
		//commitRevert()
		//xmsg.SetPendingRevert(revertMessage)
		//return
	}
	// successful HandleEVMDeposit;
	commit()
	xmsg.SetOutBoundMined("Remote omnichain contract call completed")
}
