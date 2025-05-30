package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

func TestMsgUpdateVerificationFlags_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgUpdateVerificationFlags
		err  error
	}{
		{
			name: "invalid address",
			msg: types.MsgUpdateVerificationFlags{
				Signer: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address",
			msg: types.MsgUpdateVerificationFlags{
				Signer: sample.AccAddress(),
				VerificationFlags: types.VerificationFlags{
					EthTypeChainEnabled: true,
					BtcTypeChainEnabled: true,
				},
			},
		},
		{
			name: "verification flags can be false",
			msg: types.MsgUpdateVerificationFlags{
				Signer: sample.AccAddress(),
			},
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

func TestMsgUpdateVerificationFlags_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgUpdateVerificationFlags
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.NewMsgUpdateVerificationFlags(
				signer,
				true,
				true,
			),
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.NewMsgUpdateVerificationFlags(
				"invalid",
				true,
				true,
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

func TestMsgUpdateVerificationFlags_Type(t *testing.T) {
	msg := types.MsgUpdateVerificationFlags{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.TypeMsgUpdateVerificationFlags, msg.Type())
}

func TestMsgUpdateVerificationFlags_Route(t *testing.T) {
	msg := types.MsgUpdateVerificationFlags{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgUpdateVerificationFlags_GetSignBytes(t *testing.T) {
	msg := types.MsgUpdateVerificationFlags{
		Signer: sample.AccAddress(),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}

func TestMsgUpdateVerificationFlags_GetRequireGroup(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgUpdateVerificationFlags
		want authoritytypes.PolicyType
	}{
		{
			name: "groupEmergency",
			msg: types.MsgUpdateVerificationFlags{
				VerificationFlags: types.VerificationFlags{
					EthTypeChainEnabled: false,
					BtcTypeChainEnabled: false,
				},
			},
			want: authoritytypes.PolicyType_GROUP_EMERGENCY,
		},
		{
			name: "groupOperational",
			msg: types.MsgUpdateVerificationFlags{
				VerificationFlags: types.VerificationFlags{
					EthTypeChainEnabled: true,
					BtcTypeChainEnabled: false,
				},
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
		{
			name: "groupOperational",
			msg: types.MsgUpdateVerificationFlags{
				VerificationFlags: types.VerificationFlags{
					EthTypeChainEnabled: false,
					BtcTypeChainEnabled: true,
				},
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
		{
			name: "groupOperational",
			msg: types.MsgUpdateVerificationFlags{
				VerificationFlags: types.VerificationFlags{
					EthTypeChainEnabled: true,
					BtcTypeChainEnabled: true,
				},
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.GetRequireGroup()
			require.Equal(t, tt.want, got)
		})
	}
}
