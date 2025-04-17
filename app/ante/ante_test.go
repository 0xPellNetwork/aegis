package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/app"
	"github.com/pell-chain/pellcore/app/ante"
	"github.com/pell-chain/pellcore/testutil/sample"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var _ sdk.AnteHandler = (&MockAnteHandler{}).AnteHandle

// MockAnteHandler mocks an AnteHandler
type MockAnteHandler struct {
	WasCalled bool
	CalledCtx sdk.Context
}

// AnteHandle implements AnteHandler
func (mah *MockAnteHandler) AnteHandle(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	mah.WasCalled = true
	mah.CalledCtx = ctx
	return ctx, nil
}

func TestIsSystemTx(t *testing.T) {
	// system tx types:
	// *xmsgtypes.MsgVoteGasPrice,
	// *xmsgtypes.MsgVoteOnObservedInboundTx,
	// *xmsgtypes.MsgVoteOnObservedOutboundTx,
	// *xmsgtypes.MsgAddToOutTxTracker,
	// *xmsgtypes.MsgVoteInboundBlock,
	// *relayertypes.MsgVoteBlockHeader,
	// *relayertypes.MsgVoteTSS,
	// *relayertypes.MsgAddBlameVote,
	// *xmsgtypes.MsgVoteOnGasRecharge,
	// *xmsgtypes.MsgVoteOnPellRecharge:

	buildTxFromMsg := func(msg sdk.Msg) sdk.Tx {
		txBuilder := app.MakeEncodingConfig().TxConfig.NewTxBuilder()
		txBuilder.SetMsgs(msg)
		return txBuilder.GetTx()
	}

	buildAuthzTxFromMsg := func(msg sdk.Msg) sdk.Tx {
		txBuilder := app.MakeEncodingConfig().TxConfig.NewTxBuilder()
		msgExec := authz.NewMsgExec(sample.Bech32AccAddress(), []sdk.Msg{msg})
		txBuilder.SetMsgs(&msgExec)
		return txBuilder.GetTx()
	}

	isAuthorized := func(_ string) bool {
		return true
	}
	isAuthorizedFalse := func(_ string) bool {
		return false
	}

	tests := []struct {
		name         string
		tx           sdk.Tx
		isAuthorized func(string) bool
		wantIs       bool
	}{
		{
			"MsgVoteTSS",
			buildTxFromMsg(&observertypes.MsgVoteTSS{
				Signer:    sample.AccAddress(),
				TssPubkey: "pubkey1234",
			}),
			isAuthorizedFalse,
			false,
		},
		{
			"MsgVoteTSS",
			buildTxFromMsg(&observertypes.MsgVoteTSS{
				Signer:    sample.AccAddress(),
				TssPubkey: "pubkey1234",
			}),
			isAuthorized,
			true,
		},
		{
			"MsgExec{MsgVoteTSS}",
			buildAuthzTxFromMsg(&observertypes.MsgVoteTSS{
				Signer:    sample.AccAddress(),
				TssPubkey: "pubkey1234",
			}),
			isAuthorized,

			true,
		},
		{
			"MsgSend",
			buildTxFromMsg(&banktypes.MsgSend{}),
			isAuthorized,

			false,
		},
		{
			"MsgExec{MsgSend}",
			buildAuthzTxFromMsg(&banktypes.MsgSend{}),
			isAuthorized,

			false,
		},
		{
			"MsgCreateValidator",
			buildTxFromMsg(&stakingtypes.MsgCreateValidator{}),
			isAuthorized,

			false,
		},

		{
			"MsgVoteOnObservedInboundTx",
			buildTxFromMsg(&xmsgtypes.MsgVoteOnObservedInboundTx{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgExec{MsgVoteOnObservedInboundTx}",
			buildAuthzTxFromMsg(&xmsgtypes.MsgVoteOnObservedInboundTx{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},

		{
			"MsgVoteOnObservedOutboundTx",
			buildTxFromMsg(&xmsgtypes.MsgVoteOnObservedOutboundTx{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgExec{MsgVoteOnObservedOutboundTx}",
			buildAuthzTxFromMsg(&xmsgtypes.MsgVoteOnObservedOutboundTx{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgAddToOutTxTracker",
			buildTxFromMsg(&xmsgtypes.MsgAddToOutTxTracker{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgExec{MsgAddToOutTxTracker}",
			buildAuthzTxFromMsg(&xmsgtypes.MsgAddToOutTxTracker{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgVoteTSS",
			buildTxFromMsg(&observertypes.MsgVoteTSS{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgExec{MsgVoteTSS}",
			buildAuthzTxFromMsg(&observertypes.MsgVoteTSS{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgVoteBlockHeader",
			buildTxFromMsg(&observertypes.MsgVoteBlockHeader{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgExec{MsgVoteBlockHeader}",
			buildAuthzTxFromMsg(&observertypes.MsgVoteBlockHeader{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgAddBlameVote",
			buildTxFromMsg(&observertypes.MsgAddBlameVote{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgExec{MsgAddBlameVote}",
			buildAuthzTxFromMsg(&observertypes.MsgAddBlameVote{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgVoteOnGasRecharge",
			buildTxFromMsg(&xmsgtypes.MsgVoteOnGasRecharge{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
		{
			"MsgVoteOnPellRecharge",
			buildTxFromMsg(&xmsgtypes.MsgVoteOnPellRecharge{
				Signer: sample.AccAddress(),
			}),
			isAuthorized,

			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := ante.IsSystemTx(tt.tx, tt.isAuthorized)
			require.Equal(t, tt.wantIs, is)
		})
	}
}
