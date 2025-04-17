package xmsg

import (
	"github.com/pell-chain/pellcore/pkg/coin"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// XmsgByNonceMap maps the [chainID, nonce] to the cross chain transaction
var XmsgByNonceMap = map[int64]map[uint64]*xmsgtypes.Xmsg{
	// Ethereum mainnet
	1: {
		chain_1_xmsg_6270.GetCurrentOutTxParam().OutboundTxTssNonce: chain_1_xmsg_6270,
		chain_1_xmsg_7260.GetCurrentOutTxParam().OutboundTxTssNonce: chain_1_xmsg_7260,
		chain_1_xmsg_8014.GetCurrentOutTxParam().OutboundTxTssNonce: chain_1_xmsg_8014,
		chain_1_xmsg_9718.GetCurrentOutTxParam().OutboundTxTssNonce: chain_1_xmsg_9718,
	},
	// BSC mainnet
	56: {
		chain_56_xmsg_68270.GetCurrentOutTxParam().OutboundTxTssNonce: chain_56_xmsg_68270,
	},
	// local goerli testnet
	1337: {
		chain_1337_xmsg_14.GetCurrentOutTxParam().OutboundTxTssNonce: chain_1337_xmsg_14,
	},
	// Bitcoin mainnet
	8332: {
		chain_8332_xmsg_148.GetCurrentOutTxParam().OutboundTxTssNonce: chain_8332_xmsg_148,
	},
}

// XmsgByIntxMap maps the [chainID, coinType, intxHash] to the cross chain transaction
var XmsgByIntxMap = map[int64]map[coin.CoinType]map[string]*xmsgtypes.Xmsg{
	// Ethereum mainnet
	1: {
		coin.CoinType_PELL: {
			chain_1_xmsg_intx_Pell_0xf393520.InboundTxParams.InboundTxHash: chain_1_xmsg_intx_Pell_0xf393520,
		},
		coin.CoinType_ERC20: {
			chain_1_xmsg_intx_ERC20_0x4ea69a0.InboundTxParams.InboundTxHash: chain_1_xmsg_intx_ERC20_0x4ea69a0,
		},
		coin.CoinType_GAS: {
			chain_1_xmsg_intx_Gas_0xeaec67d.InboundTxParams.InboundTxHash: chain_1_xmsg_intx_Gas_0xeaec67d,
		},
	},
	// BSC mainnet
	56: {},
	// Bitcoin mainnet
	8332: {},
}
