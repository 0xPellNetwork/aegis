package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// GetOperatorWeightedShareList returns the operator registration list for the given registry router address
func (k Keeper) GetOperatorWeightedShareList(ctx sdk.Context) (*types.LSTOperatorWeightedShareList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTOperatorWeightedShareKey))

	// Get the operator registration list for the given registry router address
	data := store.Get([]byte(types.LSTOperatorWeightedShareKey))
	if len(data) == 0 {
		return nil, false
	}

	list := new(types.LSTOperatorWeightedShareList)
	k.cdc.MustUnmarshal(data, list)

	return list, true
}

// AddOperatorWeightedShare adds a new operator registration to the operator registration list
func (k Keeper) AddOperatorWeightedShare(ctx sdk.Context, operatorData *types.LSTOperatorWeightedShare) {
	// Get existing list if it exists
	existingList, exist := k.GetOperatorWeightedShareList(ctx)
	if !exist {
		// If not found, create a new list with the operatorData
		existingList = &types.LSTOperatorWeightedShareList{
			OperatorWeightedShares: []*types.LSTOperatorWeightedShare{operatorData},
		}
	} else {
		// Append new operatorData to existing list
		existingList.OperatorWeightedShares = append(existingList.OperatorWeightedShares, operatorData)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTOperatorWeightedShareKey))
	key := []byte(types.LSTOperatorWeightedShareKey)
	store.Set(key, k.cdc.MustMarshal(existingList))
}

// SetOperatorWeightedShareList overwrites the operator registration list with the given list
func (k Keeper) SetOperatorWeightedShareList(ctx sdk.Context, list *types.LSTOperatorWeightedShareList) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTOperatorWeightedShareKey))
	key := []byte(types.LSTOperatorWeightedShareKey)
	store.Set(key, k.cdc.MustMarshal(list))
}

// SetLastRoundOperatorWeightedShareList overwrites the last round operator registration list with the given list
// Stores the previously updated weight information, using the same data structure as OperatorShares, to check if the shares information has changed.
func (k Keeper) SetLastRoundOperatorWeightedShareList(ctx sdk.Context, list *types.LSTOperatorWeightedShareList) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTLastRoundOperatorWeightedShareKey))
	key := []byte(types.LSTLastRoundOperatorWeightedShareKey)

	store.Set(key, k.cdc.MustMarshal(list))
}

// GetLastRoundOperatorWeightedShareList returns the last round operator registration list
func (k Keeper) GetLastRoundOperatorWeightedShareList(ctx sdk.Context) (*types.LSTOperatorWeightedShareList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTLastRoundOperatorWeightedShareKey))

	// Get the operator registration list for the given registry router address
	data := store.Get([]byte(types.LSTLastRoundOperatorWeightedShareKey))
	if len(data) == 0 {
		return nil, false
	}

	list := new(types.LSTOperatorWeightedShareList)
	k.cdc.MustUnmarshal(data, list)

	return list, true
}
