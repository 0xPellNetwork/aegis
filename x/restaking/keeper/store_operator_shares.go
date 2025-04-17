package keeper

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// OperatorSharesStore manages nested prefix stores for operator shares
type OperatorSharesStore struct {
	k   Keeper
	ctx sdk.Context
}

// NewOperatorSharesStore creates a new OperatorSharesStore
func (k Keeper) NewOperatorSharesStore(ctx sdk.Context) OperatorSharesStore {
	return OperatorSharesStore{k: k, ctx: ctx}
}

// Store returns the base store for operator shares
func (s OperatorSharesStore) Store() prefix.Store {
	return prefix.NewStore(s.ctx.KVStore(s.k.storeKey), types.KeyPrefix(types.KeyOperatorShareStore))
}

// ChainStore returns a store for a specific chain ID
func (s OperatorSharesStore) ChainStore(chainId uint64) prefix.Store {
	return prefix.NewStore(s.Store(), types.KeyPrefix(fmt.Sprint(chainId)))
}

// OperatorStore returns a store for a specific operator within a chain
func (s OperatorSharesStore) OperatorStore(chainId uint64, operator string) prefix.Store {
	return prefix.NewStore(s.ChainStore(chainId), []byte(operator))
}

// SetOperatorShares stores operator shares
func (k Keeper) SetOperatorShares(ctx sdk.Context, chainId uint64, operator, strategy string, shares sdkmath.Int) {
	store := k.NewOperatorSharesStore(ctx)
	operatorStore := store.OperatorStore(chainId, operator)

	value := &types.OperatorShares{
		ChainId:  chainId,
		Operator: operator,
		Strategy: strategy,
		Shares:   shares,
	}

	bz := k.cdc.MustMarshal(value)
	operatorStore.Set([]byte(strategy), bz)
}

// GetOperatorShares retrieves operator shares
func (k Keeper) GetOperatorShares(ctx sdk.Context, chainId uint64, operator, strategy string) *types.OperatorShares {
	store := k.NewOperatorSharesStore(ctx)
	operatorStore := store.OperatorStore(chainId, operator)

	bz := operatorStore.Get([]byte(strategy))
	if bz == nil {
		return nil
	}

	var shares types.OperatorShares
	k.cdc.MustUnmarshal(bz, &shares)
	return &shares
}

// GetSharesByChain gets all shares for a specific chain
func (k Keeper) GetSharesByChain(ctx sdk.Context, chainId uint64) []*types.OperatorShares {
	store := k.NewOperatorSharesStore(ctx)
	chainStore := store.ChainStore(chainId)

	var shares []*types.OperatorShares
	iterator := chainStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var share types.OperatorShares
		k.cdc.MustUnmarshal(iterator.Value(), &share)
		shares = append(shares, &share)
	}

	return shares
}

// GetSharesByChainAndOperator gets all shares for a specific operator
func (k Keeper) GetSharesByChainAndOperator(ctx sdk.Context, chainId uint64, operator string) []*types.OperatorShares {
	store := k.NewOperatorSharesStore(ctx)
	operatorStore := store.OperatorStore(chainId, operator)

	var shares []*types.OperatorShares
	iterator := operatorStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var share types.OperatorShares
		k.cdc.MustUnmarshal(iterator.Value(), &share)
		shares = append(shares, &share)
	}

	return shares
}

// IterateAllShares iterates over all shares across all chains and operators
func (k Keeper) IterateAllShares(ctx sdk.Context, cb func(*types.OperatorShares) bool) {
	store := k.NewOperatorSharesStore(ctx).Store()
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var share types.OperatorShares
		k.cdc.MustUnmarshal(iterator.Value(), &share)
		if cb(&share) {
			break
		}
	}
}

func (k Keeper) GetAllShares(ctx sdk.Context) []*types.OperatorShares {
	var shares []*types.OperatorShares
	k.IterateAllShares(ctx, func(share *types.OperatorShares) bool {
		shares = append(shares, share)
		return false
	})

	return shares
}

// DeleteAllShares deletes all shares
func (k Keeper) DeleteAllShares(ctx sdk.Context) {
	store := k.NewOperatorSharesStore(ctx).Store()
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	// Delete all key-value pairs in the store
	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

// set operator shares changed snapshot
func (k Keeper) SetChangedOperatorSharesSnapshot(ctx sdk.Context, epochNumber uint64, sharesChange []*types.OperatorShares) {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	k.Logger(ctx).Info("SetChangedOperatorSharesSnapshot", "epochNumber", epochNumber, "sharesChange", sharesChange, "xmsgIndex", inXmsgIndices)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyEpochOperatorSharesSnapshot)))

	b := k.cdc.MustMarshal(&types.ChangedOperatorSharesSnapshot{
		EpochNumber:    epochNumber,
		OperatorShares: sharesChange,
	})

	store.Set(types.KeyPrefix(fmt.Sprint(epochNumber)), b)
}

func (k Keeper) GetChangedOperatorSharesSnapshot(ctx sdk.Context, epochNumber uint64) (*types.ChangedOperatorSharesSnapshot, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyEpochOperatorSharesSnapshot)))

	b := store.Get(types.KeyPrefix(fmt.Sprint(epochNumber)))
	if b == nil {
		return nil, false
	}

	var epochSharesChange types.ChangedOperatorSharesSnapshot
	k.cdc.MustUnmarshal(b, &epochSharesChange)

	return &epochSharesChange, true
}

func (k Keeper) DeleteEpochSharesModified(ctx sdk.Context, epochNumber uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyEpochOperatorSharesSnapshot)))
	store.Delete(types.KeyPrefix(fmt.Sprint(epochNumber)))
}

// get operator shares changed snapshot by epoch range [startEpoch, endEpoch]
func (k Keeper) GetChangedOperatorSharesSnapshotByEpochRange(ctx sdk.Context, startEpoch uint64, endEpoch uint64) *types.ChangedOperatorSharesSnapshot {
	var epochSnapshotList [][]*types.OperatorShares

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		shareChange, exist := k.GetChangedOperatorSharesSnapshot(ctx, epoch)
		if exist {
			epochSnapshotList = append(epochSnapshotList, shareChange.OperatorShares)
		}
	}

	k.Logger(ctx).Info("GetChangedOperatorSharesSnapshotByEpochRange", "epochSnapshotList", epochSnapshotList, "startEpoch", startEpoch, "endEpoch", endEpoch)

	shareChanges := selectLatestOperatorSharesFromSnapshots(epochSnapshotList)

	k.Logger(ctx).Info("GetChangedOperatorSharesSnapshotByEpochRange", "shareChanges", shareChanges)

	return &types.ChangedOperatorSharesSnapshot{
		EpochNumber:    endEpoch,
		OperatorShares: shareChanges,
	}
}

// mergeMultipleShares merges multiple arrays of OperatorShares into a single sorted array
// using a divide-and-conquer approach
func selectLatestOperatorSharesFromSnapshots(sharesList [][]*types.OperatorShares) []*types.OperatorShares {
	if len(sharesList) == 0 {
		return nil
	}
	if len(sharesList) == 1 {
		return sharesList[0]
	}

	// Divide the array into two halves and process recursively
	mid := len(sharesList) / 2
	left := selectLatestOperatorSharesFromSnapshots(sharesList[:mid])
	right := selectLatestOperatorSharesFromSnapshots(sharesList[mid:])

	return mergeTwoSortedShares(left, right)
}

// mergeTwoSortedShares combines two arrays of OperatorShares into a single sorted array
// If duplicate keys are found, it keeps the record from array b (which has a higher index)
func mergeTwoSortedShares(a, b []*types.OperatorShares) []*types.OperatorShares {
	// Sort input arrays first
	sortShares(a)
	sortShares(b)

	result := make([]*types.OperatorShares, 0, len(a)+len(b))
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		comp := compareOperator(a[i], b[j])
		if comp < 0 {
			// Record from array 'a' is smaller, add it
			result = append(result, a[i])
			i++
		} else if comp > 0 {
			// Record from array 'b' is smaller, add it
			result = append(result, b[j])
			j++
		} else {
			// Same key found, keep the record from 'b' (higher index)
			result = append(result, b[j])
			i++
			j++
		}
	}

	// Add remaining records
	for ; i < len(a); i++ {
		result = append(result, a[i])
	}
	for ; j < len(b); j++ {
		result = append(result, b[j])
	}

	return result
}

// sortShares sorts a single array of OperatorShares based on ChainId, Operator, and Strategy
func sortShares(shares []*types.OperatorShares) {
	if len(shares) <= 1 {
		return
	}

	sort.SliceStable(shares, func(i, j int) bool {
		return compareOperator(shares[i], shares[j]) < 0
	})
}

// compareOperator compares two OperatorShares records based on their ChainId, Operator, and Strategy
// Returns:
//
//	-1 if a < b
//	 0 if a == b
//	 1 if a > b
func compareOperator(a, b *types.OperatorShares) int {
	if a.ChainId != b.ChainId {
		if a.ChainId < b.ChainId {
			return -1
		}
		return 1
	}
	if a.Operator != b.Operator {
		if a.Operator < b.Operator {
			return -1
		}
		return 1
	}
	if a.Strategy != b.Strategy {
		if a.Strategy < b.Strategy {
			return -1
		}
		return 1
	}
	return 0
}
