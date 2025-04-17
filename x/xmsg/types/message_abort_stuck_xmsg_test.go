package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestMsgAbortStuckXmsg_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *types.MsgAbortStuckXmsg
		err  error
	}{
		{
			name: "invalid address",
			msg:  types.NewMsgAbortStuckXmsg("invalid_address", "xmsg_index"),
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid xmsg index",
			msg:  types.NewMsgAbortStuckXmsg(sample.AccAddress(), "xmsg_index"),
			err:  types.ErrInvalidIndexValue,
		},
		{
			name: "valid",
			msg:  types.NewMsgAbortStuckXmsg(sample.AccAddress(), sample.GetXmsgIndicesFromString_pell("test")),
			err:  nil,
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

func TestMsgAbortStuckXmsg_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgAbortStuckXmsg
		panics bool
	}{
		{
			name:   "valid signer",
			msg:    types.NewMsgAbortStuckXmsg(signer, "xmsg_index"),
			panics: false,
		},
		{
			name:   "invalid signer",
			msg:    types.NewMsgAbortStuckXmsg("invalid", "xmsg_index"),
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

func TestMsgAbortStuckXmsg_Type(t *testing.T) {
	msg := types.NewMsgAbortStuckXmsg(sample.AccAddress(), "xmsg_index")
	require.Equal(t, types.TypeMsgAbortStuckXmsg, msg.Type())
}

func TestMsgAbortStuckXmsg_Route(t *testing.T) {
	msg := types.NewMsgAbortStuckXmsg(sample.AccAddress(), "xmsg_index")
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgAbortStuckXmsg_GetSignBytes(t *testing.T) {
	msg := types.NewMsgAbortStuckXmsg(sample.AccAddress(), "xmsg_index")
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
