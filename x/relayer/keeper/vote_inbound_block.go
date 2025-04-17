package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// VoteOnInboundBlockBallot VoteOnInboundBallot casts a vote on an inbound transaction observed on a connected chain. If this
// is the first vote, a new ballot is created. When a threshold of votes is
// reached, the ballot is finalized.
func (k Keeper) VoteOnInboundBlockBallot(
	ctx sdk.Context,
	chainId int64,
	voter string,
	ballotIndex string,
	blockHash string,
) (bool, bool, error) {
	if !k.IsInboundEnabled(ctx) {
		return false, false, types.ErrInboundDisabled
	}

	// makes sure we are getting only supported chains
	// if a chain support has been turned on using gov proposal
	// this function returns nil
	fromChain := k.GetSupportedChainFromChainID(ctx, chainId)
	if fromChain == nil {
		return false, false, fmt.Errorf(
			"ChainID %d, Observation %s. : %s",
			chainId,
			types.ObservationType_IN_BOUND_BLOCK.String(), types.ErrSupportedChains.Error(),
		)
	}

	// checks the voter is authorized to vote on the observation chain
	if ok := k.IsNonTombstonedObserver(ctx, voter); !ok {
		return false, false, types.ErrNotObserver
	}

	// checks against the supported chains list before querying for Ballot
	ballot, isNew, err := k.FindBallot(ctx, ballotIndex, fromChain, types.ObservationType_IN_BOUND_BLOCK)
	if err != nil {
		return false, false, err
	}
	if isNew {
		EmitEventBallotCreated(ctx, ballot, blockHash, fromChain.String(), EventTypeVoteInboundBlock)
	}

	// adds a vote and sets the ballot
	ballot, err = k.AddVoteToBallot(ctx, ballot, voter, types.VoteType_SUCCESS_OBSERVATION)
	if err != nil {
		return false, isNew, err
	}

	// checks if the ballot is finalized
	_, isFinalized := k.CheckIfFinalizingVote(ctx, ballot)
	return isFinalized, isNew, nil
}
