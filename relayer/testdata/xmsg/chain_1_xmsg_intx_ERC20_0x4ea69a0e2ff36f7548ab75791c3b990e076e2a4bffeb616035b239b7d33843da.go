package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var chain_1_xmsg_intx_ERC20_0x4ea69a0 = &xmsgtypes.Xmsg{
	Signer: "pell1pu5xy7wnwt7ukvt4yvvkldshhh0lhq6q6rhhxp",
	Index:  "0x40f6d13538e2a2a7c8b5d0dcbd6778ef5afeb45a43e946afb4d773eb24fc770a",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "Remote omnichain contract call completed",
		LastUpdateTimestamp: 1709052990,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0x56BF8D4a6E7b59D2C0E40Cba2409a4a30ab4FbE2",
		SenderChainId:                1,
		TxOrigin:                     "0x56BF8D4a6E7b59D2C0E40Cba2409a4a30ab4FbE2",
		InboundTxHash:                "0x4ea69a0e2ff36f7548ab75791c3b990e076e2a4bffeb616035b239b7d33843da",
		InboundTxBlockHeight:         19320188,
		InboundTxBallotIndex:         "0x40f6d13538e2a2a7c8b5d0dcbd6778ef5afeb45a43e946afb4d773eb24fc770a",
		InboundTxFinalizedPellHeight: 1944675,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_EXECUTED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0x56BF8D4a6E7b59D2C0E40Cba2409a4a30ab4FbE2",
					Operator: "0x56BF8D4a6E7b59D2C0E40Cba2409a4a30ab4FbE2",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0x56bf8d4a6e7b59d2c0e40cba2409a4a30ab4fbe2",
			ReceiverChainId:             86,
			OutboundTxTssNonce:          0,
			OutboundTxGasLimit:          1500000,
			OutboundTxGasPrice:          "",
			OutboundTxHash:              "0xf63eaa3e01af477673aa9e86fb634df15d30a00734dab7450cb0fc28dbc9d11b",
			OutboundTxBallotIndex:       "",
			OutboundTxExternalHeight:    1944675,
			OutboundTxGasUsed:           0,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(0),
			OutboundTxEffectiveGasLimit: 0,
			TssPubkey:                   "pellpub1addwnpepqdfkcdlnsx2lzfh7k7te55hh8me44r666504kyc0jg3g4juum077xd2vp29",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		},
	},
}
