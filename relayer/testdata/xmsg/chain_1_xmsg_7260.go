package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var chain_1_xmsg_7260 = &xmsgtypes.Xmsg{
	Signer: "",
	Index:  "0xbebecbf1d8c12016e38c09d095290df503fe29731722d939433fa47e3ed1f986",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "",
		LastUpdateTimestamp: 1709574082,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0xd91b507F2A3e2D4A32d0C86Ac19FEAD2D461008D",
		SenderChainId:                86,
		TxOrigin:                     "0x8E62e3e6FbFF3E21F725395416A20EA4E2DeF015",
		InboundTxHash:                "0x2720e3a98f18c288f4197d412bfce57e58f00dc4f8b31e335ffc0bf7208dd3c5",
		InboundTxBlockHeight:         2031411,
		InboundTxBallotIndex:         "0xbebecbf1d8c12016e38c09d095290df503fe29731722d939433fa47e3ed1f986",
		InboundTxFinalizedPellHeight: 0,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0xd91b507F2A3e2D4A32d0C86Ac19FEAD2D461008D",
					Operator: "0x8E62e3e6FbFF3E21F725395416A20EA4E2DeF015",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0x8E62e3e6FbFF3E21F725395416A20EA4E2DeF015",
			ReceiverChainId:             1,
			OutboundTxTssNonce:          7260,
			OutboundTxGasLimit:          21000,
			OutboundTxGasPrice:          "236882693686",
			OutboundTxHash:              "0xd13b593eb62b5500a00e288cc2fb2c8af1339025c0e6bc6183b8bef2ebbed0d3",
			OutboundTxBallotIndex:       "0x786f7f42a375ec8b3868edbaae11a7c7e04de7330b8e490d46fd4c3de8340486",
			OutboundTxExternalHeight:    19363323,
			OutboundTxGasUsed:           21000,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(236882693686),
			OutboundTxEffectiveGasLimit: 21000,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}
