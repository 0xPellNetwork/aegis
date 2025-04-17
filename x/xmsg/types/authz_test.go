package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestGetAllAuthzPellclientTxTypes(t *testing.T) {
	require.Equal(t, []string{
		"/xmsg.MsgVoteGasPrice",
		"/xmsg.MsgVoteOnObservedInboundTx",
		"/xmsg.MsgVoteOnObservedOutboundTx",
		"/xmsg.MsgVoteInboundBlock",
		"/xmsg.MsgAddToOutTxTracker",
		"/relayer.MsgVoteTSS",
		"/relayer.MsgAddBlameVote",
		"/relayer.MsgVoteBlockHeader",
		"/xmsg.MsgVoteOnPellRecharge",
		"/xmsg.MsgVoteOnGasRecharge",
	},
		xmsgtypes.GetAllAuthzPellclientTxTypes())
}
