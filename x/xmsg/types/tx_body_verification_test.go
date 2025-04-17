package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/coin"
	"github.com/pell-chain/pellcore/testutil/sample"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestVerifyInTxBody(t *testing.T) {
	sampleTo := sample.EthAddress()
	sampleEthTx, sampleEthTxBytes := sample.EthTx(t, chains.EthChain().Id, sampleTo, 42)

	// NOTE: errContains == "" means no error
	for _, tc := range []struct {
		desc        string
		msg         types.MsgAddToInTxTracker
		txBytes     []byte
		chainParams observertypes.ChainParams
		tss         observertypes.QueryGetTssAddressResponse
		errContains string
	}{

		{
			desc: "txBytes can't be unmarshaled",
			msg: types.MsgAddToInTxTracker{
				ChainId: chains.EthChain().Id,
			},
			txBytes:     []byte("invalid"),
			errContains: "failed to unmarshal transaction",
		},
		{
			desc: "txHash doesn't correspond",
			msg: types.MsgAddToInTxTracker{
				ChainId: chains.EthChain().Id,
				TxHash:  sample.Hash().Hex(),
			},
			txBytes:     sampleEthTxBytes,
			errContains: "invalid hash",
		},
		{
			desc: "chain id doesn't correspond",
			msg: types.MsgAddToInTxTracker{
				ChainId: chains.SepoliaChain().Id,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			txBytes:     sampleEthTxBytes,
			errContains: "invalid chain id",
		},
		{
			desc: "invalid tx event",
			msg: types.MsgAddToInTxTracker{
				ChainId:  chains.EthChain().Id,
				TxHash:   sampleEthTx.Hash().Hex(),
				CoinType: coin.CoinType(1000),
			},
			txBytes:     sampleEthTxBytes,
			chainParams: observertypes.ChainParams{StrategyManagerContractAddress: sample.EthAddress().Hex(), DelegationManagerContractAddress: sample.EthAddress().Hex()},
			errContains: "tx event is not supported",
		},
		{
			desc: "tx event is stakerdeposited",
			msg: types.MsgAddToInTxTracker{
				ChainId:  chains.EthChain().Id,
				TxHash:   sampleEthTx.Hash().Hex(),
				CoinType: coin.CoinType_PELL,
			},
			txBytes:     sampleEthTxBytes,
			chainParams: observertypes.ChainParams{StrategyManagerContractAddress: sampleTo.Hex()},
		},
		{
			desc: "tx event is stakerdelegated",
			msg: types.MsgAddToInTxTracker{
				ChainId:  chains.EthChain().Id,
				TxHash:   sampleEthTx.Hash().Hex(),
				CoinType: coin.CoinType_PELL,
			},
			txBytes:     sampleEthTxBytes,
			chainParams: observertypes.ChainParams{DelegationManagerContractAddress: sampleTo.Hex()},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := types.VerifyInTxBody(tc.msg, tc.txBytes, tc.chainParams, tc.tss)
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

func TestVerifyOutTxBody(t *testing.T) {

	sampleTo := sample.EthAddress()
	sampleEthTx, sampleEthTxBytes, sampleFrom := sample.EthTxSigned(t, chains.EthChain().Id, sampleTo, 42)
	_, sampleEthTxBytesNonSigned := sample.EthTx(t, chains.EthChain().Id, sampleTo, 42)

	// NOTE: errContains == "" means no error
	for _, tc := range []struct {
		desc        string
		msg         types.MsgAddToOutTxTracker
		txBytes     []byte
		tss         observertypes.QueryGetTssAddressResponse
		errContains string
	}{
		{
			desc: "invalid chain id",
			msg: types.MsgAddToOutTxTracker{
				ChainId: int64(1000),
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			txBytes:     sample.Bytes(),
			errContains: "cannot verify outTx body for chain",
		},
		{
			desc: "txBytes can't be unmarshaled",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			txBytes:     []byte("invalid"),
			errContains: "failed to unmarshal transaction",
		},
		{
			desc: "can't recover sender address",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			txBytes:     sampleEthTxBytesNonSigned,
			errContains: "failed to recover sender",
		},
		{
			desc: "tss address not found",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			tss:         observertypes.QueryGetTssAddressResponse{},
			txBytes:     sampleEthTxBytes,
			errContains: "tss address not found",
		},
		{
			desc: "tss address is wrong",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			tss:         observertypes.QueryGetTssAddressResponse{Eth: sample.EthAddress().Hex()},
			txBytes:     sampleEthTxBytes,
			errContains: "sender is not tss address",
		},
		{
			desc: "chain id doesn't correspond",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.SepoliaChain().Id,
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			tss:         observertypes.QueryGetTssAddressResponse{Eth: sampleFrom.Hex()},
			txBytes:     sampleEthTxBytes,
			errContains: "invalid chain id",
		},
		{
			desc: "nonce doesn't correspond",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   100,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			tss:         observertypes.QueryGetTssAddressResponse{Eth: sampleFrom.Hex()},
			txBytes:     sampleEthTxBytes,
			errContains: "invalid nonce",
		},
		{
			desc: "tx hash doesn't correspond",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   42,
				TxHash:  sample.Hash().Hex(),
			},
			tss:         observertypes.QueryGetTssAddressResponse{Eth: sampleFrom.Hex()},
			txBytes:     sampleEthTxBytes,
			errContains: "invalid tx hash",
		},
		{
			desc: "valid out tx body",
			msg: types.MsgAddToOutTxTracker{
				ChainId: chains.EthChain().Id,
				Nonce:   42,
				TxHash:  sampleEthTx.Hash().Hex(),
			},
			tss:     observertypes.QueryGetTssAddressResponse{Eth: sampleFrom.Hex()},
			txBytes: sampleEthTxBytes,
		},
		// TODO: Implement tests for verifyOutTxBodyBTC
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := types.VerifyOutTxBody(tc.msg, tc.txBytes, tc.tss)
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}
