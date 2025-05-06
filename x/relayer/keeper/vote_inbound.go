package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/coin"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// VoteOnInboundBallot casts a vote on an inbound transaction observed on a connected chain. If this
// is the first vote, a new ballot is created. When a threshold of votes is
// reached, the ballot is finalized.
func (k Keeper) VoteOnInboundBallot(
	ctx sdk.Context,
	senderChainID int64,
	receiverChainID int64,
	coinType coin.CoinType,
	voter string,
	ballotIndex string,
	inTxHash string,
) (bool, bool, error) {
	if !k.IsInboundEnabled(ctx) {
		return false, false, types.ErrInboundDisabled
	}

	// makes sure we are getting only supported chains
	// if a chain support has been turned on using gov proposal
	// this function returns nil
	senderChain := k.GetSupportedChainFromChainID(ctx, senderChainID)
	if senderChain == nil {
		return false, false, fmt.Errorf(fmt.Sprintf(
			"ChainID %d, Observation %s",
			senderChain,
			types.ObservationType_IN_BOUND_TX.String()), types.ErrSupportedChains.Error(),
		)
	}

	// checks the voter is authorized to vote on the observation chain
	if ok := k.IsNonTombstonedObserver(ctx, voter); !ok {
		return false, false, types.ErrNotObserver
	}

	// makes sure we are getting only supported chains
	receiverChain := k.GetSupportedChainFromChainID(ctx, receiverChainID)
	if receiverChain == nil {
		return false, false, fmt.Errorf(fmt.Sprintf(
			"ChainID %d, Observation %s",
			receiverChainID,
			types.ObservationType_OUT_BOUND_TX.String()), types.ErrSupportedChains.Error(),
		)
	}

	// checks against the supported chains list before querying for Ballot
	ballot, isNew, err := k.FindBallot(ctx, ballotIndex, senderChain, types.ObservationType_IN_BOUND_TX)
	if err != nil {
		return false, false, err
	}
	if isNew {
		EmitEventBallotCreated(ctx, ballot, inTxHash, senderChain.String(), EventTypeVoteInbound)
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
