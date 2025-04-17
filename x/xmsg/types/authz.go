package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
)

// GetAllAuthzPellclientTxTypes returns all the authz types for required for pellclient
func GetAllAuthzPellclientTxTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgVoteGasPrice{}),
		sdk.MsgTypeURL(&MsgVoteOnObservedInboundTx{}),
		sdk.MsgTypeURL(&MsgVoteOnObservedOutboundTx{}),
		sdk.MsgTypeURL(&MsgVoteInboundBlock{}),
		sdk.MsgTypeURL(&MsgAddToOutTxTracker{}),
		sdk.MsgTypeURL(&relayertypes.MsgVoteTSS{}),
		sdk.MsgTypeURL(&relayertypes.MsgAddBlameVote{}),
		sdk.MsgTypeURL(&relayertypes.MsgVoteBlockHeader{}),
		sdk.MsgTypeURL(&MsgVoteOnPellRecharge{}),
		sdk.MsgTypeURL(&MsgVoteOnGasRecharge{}),
	}
}
