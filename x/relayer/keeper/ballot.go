package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) SetBallot(ctx sdk.Context, ballot *types.Ballot) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.VoterKey))
	ballot.Index = ballot.BallotIdentifier
	b := k.cdc.MustMarshal(ballot)
	store.Set([]byte(ballot.Index), b)
}

func (k Keeper) SetBallotList(ctx sdk.Context, ballotlist *types.BallotListForHeight) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BallotListKey))
	b := k.cdc.MustMarshal(ballotlist)
	store.Set(types.BallotListKeyPrefix(ballotlist.Height), b)
}

// TODO: remove this after the next upgrade
func (k Keeper) DeleteBallot(ctx sdk.Context, index string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.VoterKey))
	store.Delete(types.KeyPrefix(index))
}

func (k Keeper) GetBallot(ctx sdk.Context, index string) (val types.Ballot, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.VoterKey))
	b := store.Get(types.KeyPrefix(index))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetBallotList(ctx sdk.Context, height int64) (val types.BallotListForHeight, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BallotListKey))
	b := store.Get(types.BallotListKeyPrefix(height))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetAllBallots(ctx sdk.Context) (voters []*types.Ballot) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.VoterKey))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var val types.Ballot
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		voters = append(voters, &val)
	}
	return
}

// AddBallotToList adds a ballot to the list of ballots for a given height.
func (k Keeper) AddBallotToList(ctx sdk.Context, ballot types.Ballot) {
	list, found := k.GetBallotList(ctx, ballot.BallotCreationHeight)
	if !found {
		list = types.BallotListForHeight{Height: ballot.BallotCreationHeight, BallotsIndexList: []string{}}
	}
	list.BallotsIndexList = append(list.BallotsIndexList, ballot.BallotIdentifier)
	k.SetBallotList(ctx, &list)
}

// GetMaturedBallotList Returns a list of ballots which are matured at current height
func (k Keeper) GetMaturedBallotList(ctx sdk.Context) []string {
	maturityBlocks := k.GetParamsIfExists(ctx).BallotMaturityBlocks
	list, found := k.GetBallotList(ctx, ctx.BlockHeight()-maturityBlocks)
	if !found {
		return []string{}
	}
	return list.BallotsIndexList
}
