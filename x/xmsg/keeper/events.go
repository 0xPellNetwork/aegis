package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func EmitEventInboundFinalized(ctx sdk.Context, xmsg *types.Xmsg) {
	senderChain, _ := chains.GetChainByChainId(xmsg.InboundTxParams.SenderChainId)
	currentOutParam := xmsg.GetCurrentOutTxParam()
	recvChain, _ := chains.GetChainByChainId(currentOutParam.ReceiverChainId)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventInboundFinalized{
		MsgTypeUrl:    sdk.MsgTypeURL(&types.MsgVoteOnObservedInboundTx{}),
		XmsgIndex:     xmsg.Index,
		Sender:        xmsg.InboundTxParams.Sender,
		SenderChain:   senderChain.ChainName(),
		TxOrgin:       xmsg.InboundTxParams.TxOrigin,
		InTxHash:      xmsg.InboundTxParams.InboundTxHash,
		InBlockHeight: strconv.FormatUint(xmsg.InboundTxParams.InboundTxBlockHeight, 10),
		Receiver:      currentOutParam.Receiver,
		ReceiverChain: recvChain.ChainName(),
		NewStatus:     xmsg.XmsgStatus.Status.String(),
		StatusMessage: xmsg.XmsgStatus.StatusMessage,
	}); err != nil {
		ctx.Logger().Error("Error emitting EventInboundFinalized :", err)
	}
}

func EmitOutboundSuccess(ctx sdk.Context, oldStatus string, newStatus string, xmsgIndex string) {
	err := ctx.EventManager().EmitTypedEvents(&types.EventOutboundSuccess{
		MsgTypeUrl: sdk.MsgTypeURL(&types.MsgVoteOnObservedOutboundTx{}),
		XmsgIndex:  xmsgIndex,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
	})
	if err != nil {
		ctx.Logger().Error("Error emitting MsgVoteOnObservedOutboundTx :", err)
	}

}

func EmitOutboundFailure(ctx sdk.Context, oldStatus string, newStatus string, xmsgIndex string) {
	err := ctx.EventManager().EmitTypedEvents(&types.EventOutboundFailure{
		MsgTypeUrl: sdk.MsgTypeURL(&types.MsgVoteOnObservedOutboundTx{}),
		XmsgIndex:  xmsgIndex,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
	})
	if err != nil {
		ctx.Logger().Error("Error emitting MsgVoteOnObservedOutboundTx :", err)
	}
}

func EmitEventPellSent(ctx sdk.Context, xmsg types.Xmsg) {
	currentOutParam := xmsg.GetCurrentOutTxParam()
	senderChain, _ := chains.GetChainByChainId(xmsg.InboundTxParams.SenderChainId)
	recvChain, _ := chains.GetChainByChainId(currentOutParam.ReceiverChainId)

	pellSent := xmsg.InboundTxParams.InboundPellTx.GetPellSent()
	err := ctx.EventManager().EmitTypedEvents(&types.EventPellSent{
		MsgTypeUrl:          "/pellchain.pellcore.xmsg.internal.PellSent",
		XmsgIndex:           xmsg.Index,
		Sender:              xmsg.InboundTxParams.Sender,
		SenderChain:         senderChain.ChainName(),
		InTxHash:            xmsg.InboundTxParams.InboundTxHash,
		Receiver:            currentOutParam.Receiver,
		ReceiverChain:       recvChain.ChainName(),
		PellTxOrigin:        pellSent.TxOrigin,
		PellSender:          pellSent.Sender,
		PellReceiverChainId: pellSent.ReceiverChainId,
		PellReceiver:        pellSent.Receiver,
		PellMessage:         pellSent.Message,
		PellParams:          pellSent.PellParams,
		NewStatus:           xmsg.XmsgStatus.Status.String(),
	})
	if err != nil {
		ctx.Logger().Error("Error emitting EventPellSent :", err)
	}
}

func EmitEventChainIndex(ctx sdk.Context, chainId uint64, height uint64) {
	if err := ctx.EventManager().EmitTypedEvents(&types.EventChainIndex{
		ChainId:    chainId,
		CurrHeight: height,
	}); err != nil {
		ctx.Logger().Error("Error emitting EventChainIndex :", err)
	}
}

func EmitEventStatusNode(ctx sdk.Context, statusNode *types.EventStatusNode) {
	err := ctx.EventManager().EmitTypedEvents(&types.EventStatusNode{
		PrevEventIndex:    statusNode.PrevEventIndex,
		NextEventIndex:    statusNode.NextEventIndex,
		EventIndexInBlock: statusNode.EventIndexInBlock,
		Status:            statusNode.Status,
	})
	if err != nil {
		ctx.Logger().Error("Error emitting EventChainIndex :", err)
	}
}
