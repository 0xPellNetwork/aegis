package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/authz"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestMsgVoteGasPrice_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *types.MsgVoteGasPrice
		err  error
	}{
		{
			name: "invalid address",
			msg: types.NewMsgVoteGasPrice(
				"invalid",
				1,
				1,
				"1000",
				1,
			),
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid chain id",
			msg: types.NewMsgVoteGasPrice(
				sample.AccAddress(),
				-1,
				1,
				"1000",
				1,
			),
			err: sdkerrors.ErrInvalidChainID,
		},
		{
			name: "valid address",
			msg: types.NewMsgVoteGasPrice(
				sample.AccAddress(),
				1,
				1,
				"1000",
				1,
			),
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

func TestMsgVoteGasPrice_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    types.MsgVoteGasPrice
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.MsgVoteGasPrice{
				Signer: signer,
			},
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.MsgVoteGasPrice{
				Signer: "invalid",
			},
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

func TestMsgVoteGasPricer_Type(t *testing.T) {
	msg := types.MsgVoteGasPrice{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, authz.GasPriceVoter.String(), msg.Type())
}

func TestMsgVoteGasPrice_Route(t *testing.T) {
	msg := types.MsgVoteGasPrice{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgVoteGasPrice_GetSignBytes(t *testing.T) {
	msg := types.MsgVoteGasPrice{
		Signer: sample.AccAddress(),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
