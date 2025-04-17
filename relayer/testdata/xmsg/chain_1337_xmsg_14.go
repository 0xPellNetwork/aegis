package xmsg

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// This xmsg was generated in local e2e tests, see original json text attached at end of file
var chain_1337_xmsg_14 = &xmsgtypes.Xmsg{
	Signer: "pell1mj7pr9yxrhmtr9am0hwtqx7hv8jwq9mazrsyy6",
	Index:  "0x85d06ac908823d125a919164f0596e3496224b206ebe8125ffe7b4ab766f85df",
	XmsgStatus: &xmsgtypes.Status{
		Status:              xmsgtypes.XmsgStatus_REVERTED,
		StatusMessage:       "Outbound failed, start revert : Outbound succeeded, revert executed",
		LastUpdateTimestamp: 1712705995,
	},
	InboundTxParams: &xmsgtypes.InboundTxParams{
		Sender:                       "0xBFF76e77D56B3C1202107f059425D56f0AEF87Ed",
		SenderChainId:                1337,
		TxOrigin:                     "0x5cC2fBb200A929B372e3016F1925DcF988E081fd",
		InboundTxHash:                "0xa5589bf24eca8f108ca35048adc9d5582a303d416c01319391159269ae7e4e6f",
		InboundTxBlockHeight:         177,
		InboundTxBallotIndex:         "0x85d06ac908823d125a919164f0596e3496224b206ebe8125ffe7b4ab766f85df",
		InboundTxFinalizedPellHeight: 150,
		TxFinalizationStatus:         xmsgtypes.TxFinalizationStatus_EXECUTED,
		InboundPellTx: &xmsgtypes.InboundPellEvent{
			PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
				StakerDelegated: &xmsgtypes.StakerDelegated{
					Staker:   "0xBFF76e77D56B3C1202107f059425D56f0AEF87Ed",
					Operator: "0x5cC2fBb200A929B372e3016F1925DcF988E081fd",
				},
			},
		},
	},
	OutboundTxParams: []*xmsgtypes.OutboundTxParams{
		{
			Receiver:                    "0xbff76e77d56b3c1202107f059425d56f0aef87ed",
			ReceiverChainId:             1337,
			OutboundTxTssNonce:          13,
			OutboundTxGasLimit:          250000,
			OutboundTxGasPrice:          "18",
			OutboundTxHash:              "0x19f99459da6cb08f917f9b0ee2dac94a7be328371dff788ad46e64a24e8c06c9",
			OutboundTxExternalHeight:    187,
			OutboundTxGasUsed:           67852,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(18),
			OutboundTxEffectiveGasLimit: 250000,
			TssPubkey:                   "pellpub1addwnpepqggky6z958k7hhxs6k5quuvs27uv5vtmlv330ppt2362p8ejct88w4g64jv",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
		{
			Receiver:                    "0xBFF76e77D56B3C1202107f059425D56f0AEF87Ed",
			ReceiverChainId:             1337,
			OutboundTxTssNonce:          14,
			OutboundTxGasLimit:          250000,
			OutboundTxGasPrice:          "18",
			OutboundTxHash:              "0x1487e6a31dd430306667250b72bf15b8390b73108b69f3de5c1b2efe456036a7",
			OutboundTxBallotIndex:       "0xc36c689fdaf09a9b80a614420cd4fea4fec15044790df60080cdefca0090a9dc",
			OutboundTxExternalHeight:    201,
			OutboundTxGasUsed:           76128,
			OutboundTxEffectiveGasPrice: sdkmath.NewInt(18),
			OutboundTxEffectiveGasLimit: 250000,
			TssPubkey:                   "pellpub1addwnpepqggky6z958k7hhxs6k5quuvs27uv5vtmlv330ppt2362p8ejct88w4g64jv",
			TxFinalizationStatus:        xmsgtypes.TxFinalizationStatus_EXECUTED,
		},
	},
}

// Here is the original xmsg json data used to create above chain_1337_xmsg_14
/*
{
  "creator": "pell1mj7pr9yxrhmtr9am0hwtqx7hv8jwq9mazrsyy6",
  "index": "0x85d06ac908823d125a919164f0596e3496224b206ebe8125ffe7b4ab766f85df",
  "pell_fees": "4000000000009027082",
  "relayed_message": "bgGCGUux3roBhJr9PgNaC3DOfLBp5ILuZjUZx2z1abQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQ==",
  "xmsg_status": {
    "status": 5,
    "status_message": "Outbound failed, start revert : Outbound succeeded, revert executed",
    "lastUpdate_timestamp": 1712705995
  },
  "inbound_tx_params": {
    "sender": "0xBFF76e77D56B3C1202107f059425D56f0AEF87Ed",
    "sender_chain_id": 1337,
    "tx_origin": "0x5cC2fBb200A929B372e3016F1925DcF988E081fd",
    "amount": "10000000000000000000",
    "inbound_tx_observed_hash": "0xa5589bf24eca8f108ca35048adc9d5582a303d416c01319391159269ae7e4e6f",
    "inbound_tx_observed_external_height": 177,
    "inbound_tx_ballot_index": "0x85d06ac908823d125a919164f0596e3496224b206ebe8125ffe7b4ab766f85df",
    "inbound_tx_finalized_pell_height": 150,
    "tx_finalization_status": 2
  },
  "outbound_tx_params": [
    {
      "receiver": "0xbff76e77d56b3c1202107f059425d56f0aef87ed",
      "receiver_chainId": 1337,
      "amount": "7999999999995486459",
      "outbound_tx_tss_nonce": 13,
      "outbound_tx_gas_limit": 250000,
      "outbound_tx_gas_price": "18",
      "outbound_tx_hash": "0x19f99459da6cb08f917f9b0ee2dac94a7be328371dff788ad46e64a24e8c06c9",
      "outbound_tx_observed_external_height": 187,
      "outbound_tx_gas_used": 67852,
      "outbound_tx_effective_gas_price": "18",
      "outbound_tx_effective_gas_limit": 250000,
      "tss_pubkey": "pellpub1addwnpepqggky6z958k7hhxs6k5quuvs27uv5vtmlv330ppt2362p8ejct88w4g64jv",
      "tx_finalization_status": 2
    },
    {
      "receiver": "0xBFF76e77D56B3C1202107f059425D56f0AEF87Ed",
      "receiver_chainId": 1337,
      "amount": "5999999999990972918",
      "outbound_tx_tss_nonce": 14,
      "outbound_tx_gas_limit": 250000,
      "outbound_tx_gas_price": "18",
      "outbound_tx_hash": "0x1487e6a31dd430306667250b72bf15b8390b73108b69f3de5c1b2efe456036a7",
      "outbound_tx_ballot_index": "0xc36c689fdaf09a9b80a614420cd4fea4fec15044790df60080cdefca0090a9dc",
      "outbound_tx_observed_external_height": 201,
      "outbound_tx_gas_used": 76128,
      "outbound_tx_effective_gas_price": "18",
      "outbound_tx_effective_gas_limit": 250000,
      "tss_pubkey": "pellpub1addwnpepqggky6z958k7hhxs6k5quuvs27uv5vtmlv330ppt2362p8ejct88w4g64jv",
      "tx_finalization_status": 2
    }
  ]
}
*/
