package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var chain_1_xmsg_6270 = &xmsgtypes.Xmsg{
	Signer: "",
	Index:  "0xe930f363591b348a07e0a6d309b4301b84f702e3e81e0d0902340c7f7da4b5af",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_OUTBOUND_MINED,
		StatusMessage:       "",
		LastUpdateTimestamp: 1708464433,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0xd91b507F2A3e2D4A32d0C86Ac19FEAD2D461008D",
		SenderChainId:                86,
		TxOrigin:                     "0x18D0E2c38b4188D8Ae07008C3BeeB1c80748b41c",
		InboundTxHash:                "0x8bd0df31e512c472e3162a41281b740b518216cc8eb787c2eb59c81e0cffbe89",
		InboundTxBlockHeight:         1846989,
		InboundTxBallotIndex:         "0xe930f363591b348a07e0a6d309b4301b84f702e3e81e0d0902340c7f7da4b5af",
		InboundTxFinalizedPellHeight: 0,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_NOT_FINALIZED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0xd91b507F2A3e2D4A32d0C86Ac19FEAD2D461008D",
					Operator: "0x18D0E2c38b4188D8Ae07008C3BeeB1c80748b41c",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0x18D0E2c38b4188D8Ae07008C3BeeB1c80748b41c",
			ReceiverChainId:             1,
			OutboundTxTssNonce:          6270,
			OutboundTxGasLimit:          21000,
			OutboundTxGasPrice:          "69197693654",
			OutboundTxHash:              "0x20104d41e042db754cf7908c5441914e581b498eedbca40979c9853f4b7f8460",
			OutboundTxBallotIndex:       "0x346a1d00a4d26a2065fe1dc7d5af59a49ad6a8af25853ae2ec976c07349f48c1",
			OutboundTxExternalHeight:    19271550,
			OutboundTxGasUsed:           21000,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(69197693654),
			OutboundTxEffectiveGasLimit: 21000,
			TssPubkey:                   "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}
