package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestMsgUpsertChainParams_ValidateBasic(t *testing.T) {
	externalChainList := chains.ExternalChainList()

	tests := []struct {
		name string
		msg  *types.MsgUpsertChainParams
		err  error
	}{
		{
			name: "valid message",
			msg: types.NewMsgUpsertChainParams(
				sample.AccAddress(),
				sample.ChainParams_pell(externalChainList[0].Id),
			),
		},
		{
			name: "invalid address",
			msg: types.NewMsgUpsertChainParams(
				"invalid_address",
				sample.ChainParams_pell(externalChainList[0].Id),
			),
			err: sdkerrors.ErrInvalidAddress,
		},

		{
			name: "invalid chain params (nil)",
			msg: types.NewMsgUpsertChainParams(
				sample.AccAddress(),
				nil,
			),
			err: types.ErrInvalidChainParams,
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

func TestMsgUpsertChainParams_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    types.MsgUpsertChainParams
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.MsgUpsertChainParams{
				Signer: signer,
			},
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.MsgUpsertChainParams{
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

func TestMsgUpsertChainParams_Type(t *testing.T) {
	msg := types.MsgUpsertChainParams{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.TypeMsgUpsertChainParams, msg.Type())
}

func TestMsgUpsertChainParams_Route(t *testing.T) {
	msg := types.MsgUpsertChainParams{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgUpsertChainParams_GetSignBytes(t *testing.T) {
	msg := types.MsgUpsertChainParams{
		Signer: sample.AccAddress(),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
