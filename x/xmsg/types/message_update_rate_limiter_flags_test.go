package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestMsgUpdateRateLimiterFlags_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *types.MsgUpdateRateLimiterFlags
		err  error
	}{
		{
			name: "valid message",
			msg:  types.NewMsgUpdateRateLimiterFlags(sample.AccAddress(), sample.RateLimiterFlags_pell()),
		},
		{
			name: "invalid creator address",
			msg:  types.NewMsgUpdateRateLimiterFlags("invalid", sample.RateLimiterFlags_pell()),
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid rate limiter flags",
			msg: types.NewMsgUpdateRateLimiterFlags(sample.AccAddress(), types.RateLimiterFlags{
				Window: -1,
			}),
			err: types.ErrInvalidRateLimiterFlags,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateRateLimiterFlags_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgUpdateRateLimiterFlags
		panics bool
	}{
		{
			name:   "valid signer",
			msg:    types.NewMsgUpdateRateLimiterFlags(signer, sample.RateLimiterFlags_pell()),
			panics: false,
		},
		{
			name:   "invalid signer",
			msg:    types.NewMsgUpdateRateLimiterFlags("invalid", sample.RateLimiterFlags_pell()),
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

func TestMsgUpdateRateLimiterFlags_Type(t *testing.T) {
	msg := types.NewMsgUpdateRateLimiterFlags(sample.AccAddress(), sample.RateLimiterFlags_pell())
	require.Equal(t, types.TypeMsgUpdateRateLimiterFlags, msg.Type())
}

func TestMsgUpdateRateLimiterFlags_Route(t *testing.T) {
	msg := types.NewMsgUpdateRateLimiterFlags(sample.AccAddress(), sample.RateLimiterFlags_pell())
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgUpdateRateLimiterFlags_GetSignBytes(t *testing.T) {
	msg := types.NewMsgUpdateRateLimiterFlags(sample.AccAddress(), sample.RateLimiterFlags_pell())
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
