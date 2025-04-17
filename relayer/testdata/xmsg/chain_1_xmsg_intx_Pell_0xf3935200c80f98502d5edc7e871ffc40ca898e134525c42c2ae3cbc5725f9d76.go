package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var chain_1_xmsg_intx_Pell_0xf393520 = &xmsgtypes.Xmsg{
	Signer: "pell1h8vnu70zgelr5y8feu4wytv655zua8nk7c7myy",
	Index:  "0xb8eab31d44385a9ba4d1492189d06f3d2615e2cc90d5dd91ee907e9503c9c15c",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "Remote omnichain contract call completed",
		LastUpdateTimestamp: 1708490549,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0x2f993766e8e1Ef9288B1F33F6aa244911A0A77a7",
		SenderChainId:                1,
		TxOrigin:                     "0x2f993766e8e1Ef9288B1F33F6aa244911A0A77a7",
		InboundTxHash:                "0xf3935200c80f98502d5edc7e871ffc40ca898e134525c42c2ae3cbc5725f9d76",
		InboundTxBlockHeight:         19273702,
		InboundTxBallotIndex:         "0xb8eab31d44385a9ba4d1492189d06f3d2615e2cc90d5dd91ee907e9503c9c15c",
		InboundTxFinalizedPellHeight: 1851403,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_EXECUTED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0x2f993766e8e1Ef9288B1F33F6aa244911A0A77a7",
					Operator: "0x2f993766e8e1Ef9288B1F33F6aa244911A0A77a7",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0x2f993766e8e1ef9288b1f33f6aa244911a0a77a7",
			ReceiverChainId:             86,
			OutboundTxTssNonce:          0,
			OutboundTxGasLimit:          100000,
			OutboundTxGasPrice:          "",
			OutboundTxHash:              "0x947434364da7c74d7e896a389aa8cb3122faf24bbcba64b141cb5acd7838209c",
			OutboundTxBallotIndex:       "",
			OutboundTxExternalHeight:    1851403,
			OutboundTxGasUsed:           0,
			OutboundTxEffectiveGasPrice: sdkmath.ZeroInt(),
			OutboundTxEffectiveGasLimit: 0,
			TssPubkey:                   "pellpub1addwnpepqdfkcdlnsx2lzfh7k7te55hh8me44r666504kyc0jg3g4juum077xd2vp29",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		},
	},
}
