package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestNewMsgAddBlameVoteMsg_ValidateBasic(t *testing.T) {
	keeper.SetConfig(false)
	tests := []struct {
		name  string
		msg   *types.MsgAddBlameVote
		error bool
	}{
		{
			name: "invalid creator",
			msg: types.NewMsgAddBlameVoteMsg(
				"invalid_address",
				1,
				sample.BlameRecordsList_pell(t, 1)[0],
			),
			error: true,
		},
		{
			name: "invalid chain id",
			msg: types.NewMsgAddBlameVoteMsg(
				sample.AccAddress(),
				-1,
				sample.BlameRecordsList_pell(t, 1)[0],
			),
			error: true,
		},
		{
			name: "valid",
			msg: types.NewMsgAddBlameVoteMsg(
				sample.AccAddress(),
				5,
				sample.BlameRecordsList_pell(t, 1)[0],
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

func TestNewMsgAddBlameVoteMsg_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    types.MsgAddBlameVote
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.MsgAddBlameVote{
				Signer: signer,
			},
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.MsgAddBlameVote{
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

func TestNewMsgAddBlameVoteMsg_Type(t *testing.T) {
	msg := types.MsgAddBlameVote{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.TypeMsgAddBlameVote, msg.Type())
}

func TestNewMsgAddBlameVoteMsg_Route(t *testing.T) {
	msg := types.MsgAddBlameVote{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestNewMsgAddBlameVoteMsg_GetSignBytes(t *testing.T) {
	msg := types.MsgAddBlameVote{
		Signer: sample.AccAddress(),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}

func TestNewMsgAddBlameVoteMsg_Digest(t *testing.T) {
	msg := types.MsgAddBlameVote{
		Signer: sample.AccAddress(),
	}

	digest := msg.Digest()
	msg.Signer = ""
	expectedDigest := crypto.Keccak256Hash([]byte(msg.String()))
	require.Equal(t, expectedDigest.Hex(), digest)
}
