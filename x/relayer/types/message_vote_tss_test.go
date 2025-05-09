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

func TestMsgVoteTSS_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *types.MsgVoteTSS
		err  error
	}{
		{
			name: "valid message",
			msg:  types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_SUCCESS),
		},
		{
			name: "valid message with receive status failed",
			msg:  types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_FAILED),
		},
		{
			name: "invalid creator address",
			msg:  types.NewMsgVoteTSS("invalid", "pubkey", 1, chains.ReceiveStatus_SUCCESS),
			err:  sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid observation status",
			msg:  types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_CREATED),
			err:  sdkerrors.ErrInvalidRequest,
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

func TestMsgVoteTSS_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgVoteTSS
		panics bool
	}{
		{
			name:   "valid signer",
			msg:    types.NewMsgVoteTSS(signer, "pubkey", 1, chains.ReceiveStatus_SUCCESS),
			panics: false,
		},
		{
			name:   "invalid signer",
			msg:    types.NewMsgVoteTSS("invalid", "pubkey", 1, chains.ReceiveStatus_SUCCESS),
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

func TestMsgVoteTSS_Type(t *testing.T) {
	msg := types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_SUCCESS)
	require.Equal(t, types.TypeMsgVoteTSS, msg.Type())
}

func TestMsgVoteTSS_Route(t *testing.T) {
	msg := types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_SUCCESS)
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgVoteTSS_GetSignBytes(t *testing.T) {
	msg := types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_SUCCESS)
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}

func TestMsgVoteTSS_Digest(t *testing.T) {
	msg := types.NewMsgVoteTSS(sample.AccAddress(), "pubkey", 1, chains.ReceiveStatus_SUCCESS)
	require.Equal(t, "1-tss-keygen", msg.Digest())
}
