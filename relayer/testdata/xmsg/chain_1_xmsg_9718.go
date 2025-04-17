package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var chain_1_xmsg_9718 = &xmsgtypes.Xmsg{
	Signer: "",
	Index:  "0xbf7a214cf9868e1c618123ab4df0081da87bade74eeb5aef37843e35f25e67b7",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "",
		LastUpdateTimestamp: 1712336965,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0xF0a3F93Ed1B126142E61423F9546bf1323Ff82DF",
		SenderChainId:                86,
		TxOrigin:                     "0x87257C910a19a3fe64AfFAbFe8cF9AAF2ab148BF",
		InboundTxHash:                "0xb136652cd58fb6a537b0a1677965983059a2004d98919cdacd52551f877cc44f",
		InboundTxBlockHeight:         2492552,
		InboundTxBallotIndex:         "0xbf7a214cf9868e1c618123ab4df0081da87bade74eeb5aef37843e35f25e67b7",
		InboundTxFinalizedPellHeight: 0,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0xF0a3F93Ed1B126142E61423F9546bf1323Ff82DF",
					Operator: "0x87257C910a19a3fe64AfFAbFe8cF9AAF2ab148BF",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0x30735c88fa430f11499b0edcfcc25246fb9182e3",
			ReceiverChainId:             1,
			OutboundTxTssNonce:          9718,
			OutboundTxGasLimit:          90000,
			OutboundTxGasPrice:          "112217884384",
			OutboundTxHash:              "0x81342051b8a85072d3e3771c1a57c7bdb5318e8caf37f5a687b7a91e50a7257f",
			OutboundTxBallotIndex:       "0xff07eaa34ca02a08bca1558e5f6220cbfc734061f083622b24923e032f0c480f",
			OutboundTxExternalHeight:    19590894,
			OutboundTxGasUsed:           64651,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(112217884384),
			OutboundTxEffectiveGasLimit: 100000,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}
