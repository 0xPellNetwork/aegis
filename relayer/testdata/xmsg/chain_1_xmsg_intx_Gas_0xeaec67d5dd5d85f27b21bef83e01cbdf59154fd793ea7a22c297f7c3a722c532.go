package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var chain_1_xmsg_intx_Gas_0xeaec67d = &xmsgtypes.Xmsg{
	Signer: "pell1hjct6q7npsspsg3dgvzk3sdf89spmlpf7rqmnw",
	Index:  "0x9e4a98a34b8d795ed172bf4afe2194d22ef6677739fc9fb37804d9fd4988ae51",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "Remote omnichain contract call completed",
		LastUpdateTimestamp: 1709177431,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0xF829fa7069680b8C37A8086b37d4a24697E5003b",
		SenderChainId:                1,
		TxOrigin:                     "0xF829fa7069680b8C37A8086b37d4a24697E5003b",
		InboundTxHash:                "0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532",
		InboundTxBlockHeight:         19330473,
		InboundTxBallotIndex:         "0x9e4a98a34b8d795ed172bf4afe2194d22ef6677739fc9fb37804d9fd4988ae51",
		InboundTxFinalizedPellHeight: 1965579,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_EXECUTED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0xF829fa7069680b8C37A8086b37d4a24697E5003b",
					Operator: "0xF829fa7069680b8C37A8086b37d4a24697E5003b",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0xF829fa7069680b8C37A8086b37d4a24697E5003b",
			ReceiverChainId:             86,
			OutboundTxTssNonce:          0,
			OutboundTxGasLimit:          90000,
			OutboundTxGasPrice:          "",
			OutboundTxHash:              "0x3b8c1dab5aa21ff90ddb569f2f962ff2d4aa8d914c9177900102e745955e6f35",
			OutboundTxBallotIndex:       "",
			OutboundTxExternalHeight:    1965579,
			OutboundTxGasUsed:           0,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(0),
			OutboundTxEffectiveGasLimit: 0,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		},
	},
}
