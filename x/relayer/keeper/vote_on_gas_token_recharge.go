package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

// VoteOnAddGasTokenBallot votes on the AddGasTokenBallot
func (k Keeper) VoteOnAddGasTokenBallot(ctx sdk.Context, chainId int64, voter string, voteIndex uint64) (bool, bool, error) {
	// makes sure we are getting only supported chains
	senderChain := k.GetSupportedChainFromChainID(ctx, chainId)
	if senderChain == nil {
		return false, false, fmt.Errorf(
			"ChainID %d, Observation %s: %w",
			chainId, types.ObservationType_GAS_TOKEN_RECHARGE.String(), types.ErrSupportedChains,
		)
	}

	// checks the voter is authorized to vote on the observation chain
	if ok := k.IsNonTombstonedObserver(ctx, voter); !ok {
		return false, false, types.ErrNotObserver
	}

	// checks against the supported chains list before querying for Ballot
	ballotIndex := fmt.Sprint(types.AddGasTokenBallotPrefix, voteIndex)
	ballot, isNew, err := k.FindBallot(ctx, ballotIndex, senderChain, types.ObservationType_GAS_TOKEN_RECHARGE)
	if err != nil {
		return false, false, err
	}

	if isNew {
		EmitEventBallotCreated(ctx, ballot, "", senderChain.String(), EventTypeVoteOnGasTokenRecharge)
	}

	// adds a vote and sets the ballot
	ballot, err = k.AddVoteToBallot(ctx, ballot, voter, types.VoteType_SUCCESS_OBSERVATION)
	if err != nil {
		return false, isNew, err
	}

	// checks if the ballot is finalized
	_, isFinalized := k.CheckIfFinalizingVote(ctx, ballot)
	ctx.Logger().Debug(fmt.Sprintf("VoteOnAddGasTokenBallot: processed gas recharge ballot, voteIndex: %d, isFinalized: %t", voteIndex, isFinalized))

	return isFinalized, isNew, nil
}
