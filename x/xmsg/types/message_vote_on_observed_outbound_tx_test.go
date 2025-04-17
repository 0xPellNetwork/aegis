package types_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/authz"
	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestMsgVoteOnObservedOutboundTx_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *types.MsgVoteOnObservedOutboundTx
		err  error
	}{
		{
			name: "valid message",
			msg: types.NewMsgVoteOnObservedOutboundTx(
				sample.AccAddress(),
				sample.String(),
				sample.String(),
				42,
				42,
				math.NewInt(42),
				42,
				chains.ReceiveStatus_CREATED,
				"",
				42,
				42,
			),
		},
		{
			name: "invalid address",
			msg: types.NewMsgVoteOnObservedOutboundTx(
				"invalid_address",
				sample.String(),
				sample.String(),
				42,
				42,
				math.NewInt(42),
				42,
				chains.ReceiveStatus_CREATED,
				"",
				42,
				42,
			),
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid chain ID",
			msg: types.NewMsgVoteOnObservedOutboundTx(
				sample.AccAddress(),
				sample.String(),
				sample.String(),
				42,
				42,
				math.NewInt(42),
				42,
				chains.ReceiveStatus_CREATED,
				"",
				-1,
				42,
			),
			err: types.ErrInvalidChainID,
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

func TestMsgVoteOnObservedOutboundTx_Digest(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	msg := types.MsgVoteOnObservedOutboundTx{
		Signer:                         sample.AccAddress(),
		XmsgHash:                       sample.String(),
		ObservedOutTxHash:              sample.String(),
		ObservedOutTxBlockHeight:       42,
		ObservedOutTxGasUsed:           42,
		ObservedOutTxEffectiveGasPrice: math.NewInt(42),
		ObservedOutTxEffectiveGasLimit: 42,
		Status:                         chains.ReceiveStatus_CREATED,
		OutTxChain:                     42,
		OutTxTssNonce:                  42,
	}
	hash := msg.Digest()
	require.NotEmpty(t, hash, "hash should not be empty")

	// creator not used
	msg2 := msg
	msg2.Signer = sample.AccAddress()
	hash2 := msg2.Digest()
	require.Equal(t, hash, hash2, "creator should not change hash")

	// status not used
	msg2 = msg
	msg2.Status = chains.ReceiveStatus_FAILED
	hash2 = msg2.Digest()
	require.Equal(t, hash, hash2, "status should not change hash")

	// xmsg hash used
	msg2 = msg
	msg2.XmsgHash = sample.StringRandom(r, 32)
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "xmsg hash should change hash")

	// observed outbound tx hash used
	msg2 = msg
	msg2.ObservedOutTxHash = sample.StringRandom(r, 32)
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "observed outbound tx hash should change hash")

	// observed outbound tx block height used
	msg2 = msg
	msg2.ObservedOutTxBlockHeight = 43
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "observed outbound tx block height should change hash")

	// observed outbound tx gas used used
	msg2 = msg
	msg2.ObservedOutTxGasUsed = 43
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "observed outbound tx gas used should change hash")

	// observed outbound tx effective gas price used
	msg2 = msg
	msg2.ObservedOutTxEffectiveGasPrice = math.NewInt(43)
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "observed outbound tx effective gas price should change hash")

	// observed outbound tx effective gas limit used
	msg2 = msg
	msg2.ObservedOutTxEffectiveGasLimit = 43
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "observed outbound tx effective gas limit should change hash")

	// out tx chain used
	msg2 = msg
	msg2.OutTxChain = 43
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "out tx chain should change hash")

	// out tx tss nonce used
	msg2 = msg
	msg2.OutTxTssNonce = 43
	hash2 = msg2.Digest()
	require.NotEqual(t, hash, hash2, "out tx tss nonce should change hash")
}

func TestMsgVoteOnObservedOutboundTx_GetSigners(t *testing.T) {
	signer := sample.AccAddress()
	tests := []struct {
		name   string
		msg    types.MsgVoteOnObservedOutboundTx
		panics bool
	}{
		{
			name: "valid signer",
			msg: types.MsgVoteOnObservedOutboundTx{
				Signer: signer,
			},
			panics: false,
		},
		{
			name: "invalid signer",
			msg: types.MsgVoteOnObservedOutboundTx{
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

func TestMsgVoteOnObservedOutboundTx_Type(t *testing.T) {
	msg := types.MsgVoteOnObservedOutboundTx{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, authz.OutboundVoter.String(), msg.Type())
}

func TestMsgVoteOnObservedOutboundTx_Route(t *testing.T) {
	msg := types.MsgVoteOnObservedOutboundTx{
		Signer: sample.AccAddress(),
	}
	require.Equal(t, types.RouterKey, msg.Route())
}

func TestMsgVoteOnObservedOutboundTx_GetSignBytes(t *testing.T) {
	msg := types.MsgVoteOnObservedOutboundTx{
		Signer: sample.AccAddress(),
	}
	require.NotPanics(t, func() {
		msg.GetSignBytes()
	})
}
