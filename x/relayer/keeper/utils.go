package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) AddVoteToBallot(ctx sdk.Context, ballot types.Ballot, address string, observationType types.VoteType) (types.Ballot, error) {
	ballot, err := ballot.AddVote(address, observationType)
	if err != nil {
		return ballot, err
	}
	ctx.Logger().Info(fmt.Sprintf("Vote Added | Voter :%s, ballot identifier %s", address, ballot.BallotIdentifier))
	k.SetBallot(ctx, &ballot)
	return ballot, nil
}

// CheckIfFinalizingVote checks if the ballot is finalized in this block and if it is, it sets the ballot in the store
// This function with only return true if the ballot moves for pending to success or failed status with this vote.
// If the ballot is already finalized in the previous vote , it will return false
func (k Keeper) CheckIfFinalizingVote(ctx sdk.Context, ballot types.Ballot) (types.Ballot, bool) {
	ballot, isFinalized := ballot.IsFinalizingVote()
	if !isFinalized {
		return ballot, false
	}
	k.SetBallot(ctx, &ballot)
	return ballot, true
}

// IsNonTombstonedObserver checks whether a signer is authorized to sign
// This function checks if the signer is present in the observer set
// and also checks if the signer is not tombstoned
func (k Keeper) IsNonTombstonedObserver(ctx sdk.Context, address string) bool {
	isPresentInMapper := k.IsAddressPartOfObserverSet(ctx, address)
	if !isPresentInMapper {
		return false
	}
	isTombstoned, err := k.IsOperatorTombstoned(ctx, address)
	if err != nil || isTombstoned {
		return false
	}
	return true
}

// FindBallot finds the ballot for the given index
// If the ballot is not found, it creates a new ballot and returns it
func (k Keeper) FindBallot(
	ctx sdk.Context,
	index string,
	chain *chains.Chain,
	observationType types.ObservationType,
) (ballot types.Ballot, isNew bool, err error) {
	isNew = false
	ballot, found := k.GetBallot(ctx, index)
	if !found {
		observerSet, _ := k.GetObserverSet(ctx)

		cp, found := k.GetChainParamsByChainID(ctx, chain.Id)
		if !found || cp == nil || !cp.IsSupported {
			return types.Ballot{}, false, types.ErrSupportedChains
		}

		ballot = types.Ballot{
			Index:                "",
			BallotIdentifier:     index,
			VoterList:            observerSet.RelayerList,
			Votes:                types.CreateVotes(len(observerSet.RelayerList)),
			ObservationType:      observationType,
			BallotThreshold:      cp.BallotThreshold,
			BallotStatus:         types.BallotStatus_BALLOT_IN_PROGRESS,
			BallotCreationHeight: ctx.BlockHeight(),
		}
		isNew = true
		k.AddBallotToList(ctx, ballot)
	}
	return
}

func (k Keeper) IsValidator(ctx sdk.Context, creator string) error {
	valAddress, err := types.GetOperatorAddressFromAccAddress(creator)
	if err != nil {
		return err
	}
	validator, err := k.stakingKeeper.GetValidator(ctx, valAddress)
	if err != nil {
		return types.ErrNotValidator
	}

	if validator.Jailed || !validator.IsBonded() {
		return types.ErrValidatorStatus
	}
	return nil
}

func (k Keeper) IsOperatorTombstoned(ctx sdk.Context, creator string) (bool, error) {
	valAddress, err := types.GetOperatorAddressFromAccAddress(creator)
	if err != nil {
		return false, err
	}
	validator, err := k.stakingKeeper.GetValidator(ctx, valAddress)
	if err != nil {
		return false, types.ErrNotValidator
	}

	consAddress, err := validator.GetConsAddr()
	if err != nil {
		return false, err
	}
	return k.slashingKeeper.IsTombstoned(ctx, consAddress), nil
}

func (k Keeper) CheckObserverSelfDelegation(ctx context.Context, accAddress string) error {
	selfdelAddr, err := sdk.AccAddressFromBech32(accAddress)
	if err != nil {
		return err
	}
	valAddress, err := types.GetOperatorAddressFromAccAddress(accAddress)
	if err != nil {
		return err
	}
	validator, err := k.stakingKeeper.GetValidator(ctx, valAddress)
	if err != nil {
		return types.ErrNotValidator
	}

	delegation, err := k.stakingKeeper.GetDelegation(ctx, selfdelAddr, valAddress)
	if err != nil {
		// TODO: How to handle this case?
		k.RemoveObserverFromSet(ctx, accAddress)
		return nil
		//return types.ErrSelfDelegation
	}

	minDelegation, err := types.GetMinObserverDelegationDec()
	if err != nil {
		return err
	}
	tokens := validator.TokensFromShares(delegation.Shares)
	if tokens.LT(minDelegation) {
		k.RemoveObserverFromSet(ctx, accAddress)
	}
	return nil
}
