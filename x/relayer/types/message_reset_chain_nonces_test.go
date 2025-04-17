package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestMsgResetChainNonces_ValidateBasic(t *testing.T) {
	externalChainList := chains.ExternalChainList()

	tests := []struct {
		name    string
		msg     types.MsgResetChainNonces
		wantErr bool
	}{
		{
			name: "valid message chain nonce high greater than nonce low",
			msg: types.MsgResetChainNonces{
				Signer:         sample.AccAddress(),
				ChainId:        externalChainList[0].Id,
				ChainNonceLow:  1,
				ChainNonceHigh: 5,
			},
			wantErr: false,
		},
		{
			name: "valid message chain nonce high same as nonce low",
			msg: types.MsgResetChainNonces{
				Signer:         sample.AccAddress(),
				ChainId:        externalChainList[0].Id,
				ChainNonceLow:  1,
				ChainNonceHigh: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			msg: types.MsgResetChainNonces{
				Signer:  "invalid_address",
				ChainId: externalChainList[0].Id,
			},
			wantErr: true,
		},
		{
			name: "invalid chain ID",
			msg: types.MsgResetChainNonces{
				Signer:  sample.AccAddress(),
				ChainId: 999,
			},
			wantErr: true,
		},
		{
			name: "invalid chain nonce low",
			msg: types.MsgResetChainNonces{
				Signer:        sample.AccAddress(),
				ChainId:       externalChainList[0].Id,
				ChainNonceLow: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid chain nonce high",
			msg: types.MsgResetChainNonces{
				Signer:         sample.AccAddress(),
				ChainId:        externalChainList[0].Id,
				ChainNonceLow:  1,
				ChainNonceHigh: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid chain nonce low greater than chain nonce high",
			msg: types.MsgResetChainNonces{
				Signer:         sample.AccAddress(),
				ChainId:        externalChainList[0].Id,
				ChainNonceLow:  1,
				ChainNonceHigh: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgResetChainNonces_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    *types.MsgResetChainNonces
		panics bool
	}{
		{
			name:   "valid signer",
			msg:    types.NewMsgResetChainNonces(signer, 5, 1, 5),
			panics: false,
		},
		{
			name:   "invalid signer",
			msg:    types.NewMsgResetChainNonces("invalid", 5, 1, 5),
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

func TestMsgResetChainNonces_Type(t *testing.T) {
	msg := types.NewMsgResetChainNonces(sample.AccAddress(), 5, 1, 5)
	require.Equal(t, types.TypeMsgResetChainNonces, msg.Type())
}

func TestMsgResetChainNonces_Route(t *testing.T) {
	msg := types.NewMsgResetChainNonces(sample.AccAddress(), 5, 1, 5)
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgResetChainNonces_GetSignBytes(t *testing.T) {
	msg := types.NewMsgResetChainNonces(sample.AccAddress(), 5, 1, 5)
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
