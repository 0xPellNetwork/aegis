package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

// VoteOnOutboundBallot casts a vote on an outbound transaction observed on a connected chain (after
// it has been broadcasted to and finalized on a connected chain). If this is
// the first vote, a new ballot is created. When a threshold of votes is
// reached, the ballot is finalized.
// returns if the vote is finalized, if the ballot is new, the ballot status and the name of the observation chain
func (k Keeper) VoteOnOutboundBallot(
	ctx sdk.Context,
	ballotIndex string,
	outTxChainID int64,
	receiveStatus chains.ReceiveStatus,
	voter string,
) (isFinalized bool, isNew bool, ballot relayertypes.Ballot, observationChainName string, err error) {
	// Observer Chain already checked then inbound is created
	/* EDGE CASE : Params updated in during the finalization process
	   i.e Inbound has been finalized but outbound is still pending
	*/
	observationChain := k.GetSupportedChainFromChainID(ctx, outTxChainID)
	if observationChain == nil {
		return false, false, ballot, "", relayertypes.ErrSupportedChains
	}
	if relayertypes.CheckReceiveStatus(receiveStatus) != nil {
		return false, false, ballot, "", relayertypes.ErrInvalidStatus
	}

	// check if voter is authorized
	if ok := k.IsNonTombstonedObserver(ctx, voter); !ok {
		return false, false, ballot, "", relayertypes.ErrNotObserver
	}

	// fetch or create ballot
	ballot, isNew, err = k.FindBallot(ctx, ballotIndex, observationChain, relayertypes.ObservationType_OUT_BOUND_TX)
	if err != nil {
		return false, false, ballot, "", err
	}

	// add vote to ballot
	ballot, err = k.AddVoteToBallot(ctx, ballot, voter, relayertypes.ConvertReceiveStatusToVoteType(receiveStatus))
	if err != nil {
		return false, false, ballot, "", err
	}

	ballot, isFinalizedInThisBlock := k.CheckIfFinalizingVote(ctx, ballot)
	return isFinalizedInThisBlock, isNew, ballot, observationChain.String(), nil
}
