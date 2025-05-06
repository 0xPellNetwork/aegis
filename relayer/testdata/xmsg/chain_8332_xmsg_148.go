package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var chain_8332_xmsg_148 = &xmsgtypes.Xmsg{
	Signer: "",
	Index:  "0xb3f5f3cf2ed2e0c3fa64c8297c9e50fbc07351fb2d26d8eae4cfbbd45e47a524",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "",
		LastUpdateTimestamp: 1708608895,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0x13A0c5930C028511Dc02665E7285134B6d11A5f4",
		SenderChainId:                86,
		TxOrigin:                     "0xe99174F08e1186134830f8511De06bd010978533",
		InboundTxHash:                "0x06455013319acb1b027461134853c77b003d8eab162b1f37673da5ad8a50b74f",
		InboundTxBlockHeight:         1870408,
		InboundTxBallotIndex:         "0xb3f5f3cf2ed2e0c3fa64c8297c9e50fbc07351fb2d26d8eae4cfbbd45e47a524",
		InboundTxFinalizedPellHeight: 0,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0x13A0c5930C028511Dc02665E7285134B6d11A5f4",
					Operator: "0xe99174F08e1186134830f8511De06bd010978533",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "bc1qpsdlklfcmlcfgm77c43x65ddtrt7n0z57hsyjp",
			ReceiverChainId:             8332,
			OutboundTxTssNonce:          148,
			OutboundTxGasLimit:          254,
			OutboundTxGasPrice:          "46",
			OutboundTxHash:              "030cd813443f7b70cc6d8a544d320c6d8465e4528fc0f3410b599dc0b26753a0",
			OutboundTxExternalHeight:    150,
			OutboundTxGasUsed:           0,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(0),
			OutboundTxEffectiveGasLimit: 0,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}
