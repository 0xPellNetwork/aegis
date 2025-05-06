package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestNewMsgMigrateTssFunds_ValidateBasic(t *testing.T) {
	keeper.SetConfig(false)
	tests := []struct {
		name  string
		msg   *types.MsgMigrateTssFunds
		error bool
	}{
		{
			name: "invalid creator",
			msg: types.NewMsgMigrateTssFunds(
				"invalid address",
				chains.ChainsList()[0].Id,
				sdkmath.NewUintFromString("100000"),
			),
			error: true,
		},
		{
			name: "invalid chain id",
			msg: types.NewMsgMigrateTssFunds(
				sample.AccAddress(),
				999,
				sdkmath.NewUintFromString("100000"),
			),
			error: true,
		},
		{
			name: "invalid amount",
			msg: types.NewMsgMigrateTssFunds(
				sample.AccAddress(),
				chains.ChainsList()[0].Id,
				sdkmath.NewUintFromString("0"),
			),
			error: true,
		},
		{
			name: "valid msg",
			msg: types.NewMsgMigrateTssFunds(
				sample.AccAddress(),
				chains.ChainsList()[0].Id,
				sdkmath.NewUintFromString("100000"),
			),
			error: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.error {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewMsgMigrateTssFunds_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    types.MsgMigrateTssFunds
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.MsgMigrateTssFunds{
				Signer:  signer,
				ChainId: chains.ChainsList()[0].Id,
				Amount:  sdkmath.NewUintFromString("100000"),
			},
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.MsgMigrateTssFunds{
				Signer:  "invalid_address",
				ChainId: chains.ChainsList()[0].Id,
				Amount:  sdkmath.NewUintFromString("100000"),
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

func TestNewMsgMigrateTssFunds_Type(t *testing.T) {
	msg := types.MsgMigrateTssFunds{
		Signer:  sample.AccAddress(),
		ChainId: chains.ChainsList()[0].Id,
		Amount:  sdkmath.NewUintFromString("100000"),
	}
	require.Equal(t, types.TypeMsgMigrateTssFunds, msg.Type())
}

func TestNewMsgMigrateTssFunds_Route(t *testing.T) {
	msg := types.MsgMigrateTssFunds{
		Signer:  sample.AccAddress(),
		ChainId: chains.ChainsList()[0].Id,
		Amount:  sdkmath.NewUintFromString("100000"),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestNewMsgMigrateTssFunds_GetSignBytes(t *testing.T) {
	msg := types.MsgMigrateTssFunds{
		Signer:  sample.AccAddress(),
		ChainId: chains.ChainsList()[0].Id,
		Amount:  sdkmath.NewUintFromString("100000"),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
