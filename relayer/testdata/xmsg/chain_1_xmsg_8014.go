package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var chain_1_xmsg_8014 = &xmsgtypes.Xmsg{
	Signer: "",
	Index:  "0x5a100fdb426da35ad4c95520d7a4f1fd2f38c88067c9e80ba209d3a655c6e06e",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "",
		LastUpdateTimestamp: 1710834402,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0x7c8dDa80bbBE1254a7aACf3219EBe1481c6E01d7",
		SenderChainId:                86,
		TxOrigin:                     "0x8d8D67A8B71c141492825CAE5112Ccd8581073f2",
		InboundTxHash:                "0x114ed9d327b6afc068c3fa891b82f7c7f2d42ae25a571f7dc004c05e77af592a",
		InboundTxBlockHeight:         2241077,
		InboundTxBallotIndex:         "0x5a100fdb426da35ad4c95520d7a4f1fd2f38c88067c9e80ba209d3a655c6e06e",
		InboundTxFinalizedPellHeight: 0,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0x7c8dDa80bbBE1254a7aACf3219EBe1481c6E01d7",
					Operator: "0x8d8D67A8B71c141492825CAE5112Ccd8581073f2",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0x8d8D67A8B71c141492825CAE5112Ccd8581073f2",
			ReceiverChainId:             1,
			OutboundTxTssNonce:          8014,
			OutboundTxGasLimit:          100000,
			OutboundTxGasPrice:          "58619665744",
			OutboundTxHash:              "0xd2eba7ac3da1b62800165414ea4bcaf69a3b0fb9b13a0fc32f4be11bfef79146",
			OutboundTxBallotIndex:       "0x4213f2c335758301b8bbb09d9891949ed6ffeea5dd95e5d9eaa8d410baaa0884",
			OutboundTxExternalHeight:    19467367,
			OutboundTxGasUsed:           60625,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(58619665744),
			OutboundTxEffectiveGasLimit: 100000,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}
