package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var chain_56_xmsg_68270 = &xmsgtypes.Xmsg{
	Signer: "",
	Index:  "0x541b570182950809f9b9077861a0fc7038af9a14ce8af4e151a83adfa308c7a9",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_PENDING_OUTBOUND,
		StatusMessage:       "",
		LastUpdateTimestamp: 1709145183,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0xd91b507F2A3e2D4A32d0C86Ac19FEAD2D461008D",
		SenderChainId:                86,
		TxOrigin:                     "0xb0C04e07A301927672A8A7a874DB6930576C90B8",
		InboundTxHash:                "0x093f4ca4c1884df0fd9dd59b75979342ded29d3c9b6861644287a2e1417b9a39",
		InboundTxBlockHeight:         1960153,
		InboundTxBallotIndex:         "0x541b570182950809f9b9077861a0fc7038af9a14ce8af4e151a83adfa308c7a9",
		InboundTxFinalizedPellHeight: 0,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0xd91b507F2A3e2D4A32d0C86Ac19FEAD2D461008D",
					Operator: "0xb0C04e07A301927672A8A7a874DB6930576C90B8",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0xb0C04e07A301927672A8A7a874DB6930576C90B8",
			ReceiverChainId:             56,
			OutboundTxTssNonce:          68270,
			OutboundTxGasLimit:          21000,
			OutboundTxGasPrice:          "6000000000",
			OutboundTxHash:              "0xeb2b183ece6638688b9df9223180b13a67208cd744bbdadeab8de0482d7f4e3c",
			OutboundTxBallotIndex:       "0xa4600c952934f797e162d637d70859a611757407908d96bc53e45a81c80b006b",
			OutboundTxExternalHeight:    36537856,
			OutboundTxGasUsed:           21000,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(6000000000),
			OutboundTxEffectiveGasLimit: 21000,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}
