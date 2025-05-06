package types_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/coin"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestMsgAddToInTxTracker_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *types.MsgAddToInTxTracker
		err  error
	}{
		{
			name: "invalid address",
			msg: types.NewMsgAddToInTxTracker(
				"invalid_address",
				chains.GoerliChain().Id,
				coin.CoinType_GAS,
				"hash",
			),
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid chain id",
			msg: types.NewMsgAddToInTxTracker(
				sample.AccAddress(),
				42,
				coin.CoinType_GAS,
				"hash",
			),
			err: errorsmod.Wrapf(types.ErrInvalidChainID, "chain id (%d)", 42),
		},
		// TODO: current there are no non-EVM chains
		//{
		//	name: "invalid proof",
		//	msg: &types.MsgAddToInTxTracker{
		//		Signer:   sample.AccAddress(),
		//		ChainId:  chains.PellTestnetChain().Id,
		//		CoinType: coin.CoinType_GAS,
		//		Proof:    &proofs.Proof{},
		//	},
		//	err: errorsmod.Wrapf(types.ErrProofVerificationFail, "chain id %d does not support proof-based trackers", chains.PellTestnetChain().Id),
		//},
		{
			name: "invalid coin type",
			msg: &types.MsgAddToInTxTracker{
				Signer:   sample.AccAddress(),
				ChainId:  chains.PellTestnetChain().Id,
				CoinType: 5,
			},
			err: errorsmod.Wrapf(types.ErrProofVerificationFail, "coin-type not supported"),
		},
		{
			name: "valid",
			msg: types.NewMsgAddToInTxTracker(
				sample.AccAddress(),
				chains.GoerliChain().Id,
				coin.CoinType_GAS,
				"hash",
			),
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgAddToInTxTracker_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgAddToInTxTracker
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.NewMsgAddToInTxTracker(
				signer,
				chains.GoerliChain().Id,
				coin.CoinType_GAS,
				"hash",
			),
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.NewMsgAddToInTxTracker(
				"invalid_address",
				chains.GoerliChain().Id,
				coin.CoinType_GAS,
				"hash",
			),
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.panics {
				signers := tt.msg.GetSigners()
				require.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(signer)}, signers)
			} else {
				require.Panics(t, func() {
					tt.msg.GetSigners()
				})
			}
		})
	}
}

func TestMsgAddToInTxTracker_Type(t *testing.T) {
	msg := types.NewMsgAddToInTxTracker(
		sample.AccAddress(),
		chains.GoerliChain().Id,
		coin.CoinType_GAS,
		"hash",
	)
	require.Equal(t, types.TypeMsgAddToInTxTracker, msg.Type())
}

func TestMsgAddToInTxTracker_Route(t *testing.T) {
	msg := types.NewMsgAddToInTxTracker(
		sample.AccAddress(),
		chains.GoerliChain().Id,
		coin.CoinType_GAS,
		"hash",
	)
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgAddToInTxTracker_GetSignBytes(t *testing.T) {
	msg := types.NewMsgAddToInTxTracker(
		sample.AccAddress(),
		chains.GoerliChain().Id,
		coin.CoinType_GAS,
		"hash",
	)
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
