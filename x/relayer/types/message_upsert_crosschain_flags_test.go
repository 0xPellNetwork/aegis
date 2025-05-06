package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestMsgUpsertCrosschainFlags_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgUpsertCrosschainFlags
		err  error
	}{
		{
			name: "invalid address",
			msg: types.MsgUpsertCrosschainFlags{
				Signer: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid gas price increase flags",
			msg: types.MsgUpsertCrosschainFlags{
				Signer: sample.AccAddress(),
				GasPriceIncreaseFlags: &types.GasPriceIncreaseFlags{
					EpochLength:             -1,
					RetryInterval:           1,
					GasPriceIncreasePercent: 1,
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "valid address",
			msg: types.MsgUpsertCrosschainFlags{
				Signer: sample.AccAddress(),
				GasPriceIncreaseFlags: &types.GasPriceIncreaseFlags{
					EpochLength:             1,
					RetryInterval:           1,
					GasPriceIncreasePercent: 1,
				},
			},
		},
		{
			name: "gas price increase flags can be nil",
			msg: types.MsgUpsertCrosschainFlags{
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

func TestGasPriceIncreaseFlags_Validate(t *testing.T) {
	tests := []struct {
		name        string
		gpf         types.GasPriceIncreaseFlags
		errContains string
	}{
		{
			name: "invalid epoch length",
			gpf: types.GasPriceIncreaseFlags{
				EpochLength:             -1,
				RetryInterval:           1,
				GasPriceIncreasePercent: 1,
			},
			errContains: "epoch length must be positive",
		},
		{
			name: "invalid retry interval",
			gpf: types.GasPriceIncreaseFlags{
				EpochLength:             1,
				RetryInterval:           -1,
				GasPriceIncreasePercent: 1,
			},
			errContains: "retry interval must be positive",
		},
		{
			name: "valid",
			gpf: types.GasPriceIncreaseFlags{
				EpochLength:             1,
				RetryInterval:           1,
				GasPriceIncreasePercent: 1,
			},
		},
		{
			name: "percent can be 0",
			gpf: types.GasPriceIncreaseFlags{
				EpochLength:             1,
				RetryInterval:           1,
				GasPriceIncreasePercent: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gpf.Validate()
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgUpsertCrosschainFlags_GetRequiredPolicyType(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgUpsertCrosschainFlags
		want authoritytypes.PolicyType
	}{
		{
			name: "disabling outbound and inbound allows group 1",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:                       sample.AccAddress(),
				IsInboundEnabled:             false,
				IsOutboundEnabled:            false,
				BlockHeaderVerificationFlags: nil,
				GasPriceIncreaseFlags:        nil,
			},
			want: authoritytypes.PolicyType_GROUP_EMERGENCY,
		},
		{
			name: "disabling outbound and inbound and block header verification allows group 1",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:            sample.AccAddress(),
				IsInboundEnabled:  false,
				IsOutboundEnabled: false,
				BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
					IsEthTypeChainEnabled: false,
					IsBtcTypeChainEnabled: false,
				},
				GasPriceIncreaseFlags: nil,
			},
			want: authoritytypes.PolicyType_GROUP_EMERGENCY,
		},
		{
			name: "updating gas price increase flags asserts group 2",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:            sample.AccAddress(),
				IsInboundEnabled:  false,
				IsOutboundEnabled: false,
				BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
					IsEthTypeChainEnabled: false,
					IsBtcTypeChainEnabled: false,
				},
				GasPriceIncreaseFlags: &types.GasPriceIncreaseFlags{
					EpochLength:             1,
					RetryInterval:           1,
					GasPriceIncreasePercent: 1,
					MaxPendingXmsgs:         100,
				},
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
		{
			name: "enabling inbound asserts group 2",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:            sample.AccAddress(),
				IsInboundEnabled:  true,
				IsOutboundEnabled: false,
				BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
					IsEthTypeChainEnabled: false,
					IsBtcTypeChainEnabled: false,
				},
				GasPriceIncreaseFlags: nil,
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
		{
			name: "enabling outbound asserts group 2",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:            sample.AccAddress(),
				IsInboundEnabled:  false,
				IsOutboundEnabled: true,
				BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
					IsEthTypeChainEnabled: false,
					IsBtcTypeChainEnabled: false,
				},
				GasPriceIncreaseFlags: nil,
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
		{
			name: "enabling eth header verification asserts group 2",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:            sample.AccAddress(),
				IsInboundEnabled:  false,
				IsOutboundEnabled: false,
				BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
					IsEthTypeChainEnabled: true,
					IsBtcTypeChainEnabled: false,
				},
				GasPriceIncreaseFlags: nil,
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
		{
			name: "enabling btc header verification asserts group 2",
			msg: types.MsgUpsertCrosschainFlags{
				Signer:            sample.AccAddress(),
				IsInboundEnabled:  false,
				IsOutboundEnabled: false,
				BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
					IsEthTypeChainEnabled: false,
					IsBtcTypeChainEnabled: true,
				},
				GasPriceIncreaseFlags: nil,
			},
			want: authoritytypes.PolicyType_GROUP_OPERATIONAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualValues(t, tt.want, tt.msg.GetRequiredPolicyType())
		})
	}
}

func TestMsgUpsertCrosschainFlags_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgUpsertCrosschainFlags
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.NewMsgUpsertCrosschainFlags(
				signer,
				true,
				true,
			),
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.NewMsgUpsertCrosschainFlags(
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

func TestMsgUpsertCrosschainFlags_Type(t *testing.T) {
	msg := types.MsgUpsertCrosschainFlags{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.TypeMsgUpsertCrosschainFlags, msg.Type())
}

func TestMsgUpsertCrosschainFlags_Route(t *testing.T) {
	msg := types.MsgUpsertCrosschainFlags{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgUpsertCrosschainFlags_GetSignBytes(t *testing.T) {
	msg := types.MsgUpsertCrosschainFlags{
		Signer: sample.AccAddress(),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
