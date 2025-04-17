package keeper

import (
	"math/big"

	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// GetGroupDataList gets the group data list
func (k Keeper) GetGroupDataList(ctx sdk.Context, registryRouterAddress ethcommon.Address) (*types.GroupList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the Group array for the given registry router address
	key := types.GroupDataKey(registryRouterAddress)
	data := store.Get(key)
	if data == nil {
		return nil, false
	}

	var groupList types.GroupList
	k.cdc.MustUnmarshal(data, &groupList)

	return &groupList, true
}

// GetGroupByGroupNumber queries a single Group by its group number.
func (k Keeper) GetGroupByGroupNumber(
	ctx sdk.Context,
	registryRouterAddress ethcommon.Address,
	groupNumber uint64,
) (*types.Group, error) {
	groupList, exists := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exists {
		return nil, cosmoserrors.Wrapf(types.ErrInvalidData, "no group data for address %s", registryRouterAddress.Hex())
	}

	for _, group := range groupList.Groups {
		if group.GroupNumber == groupNumber {
			return group, nil
		}
	}

	return nil, cosmoserrors.Wrapf(types.ErrInvalidData, "group record %d not found", groupNumber)
}

// AddGroupData adds the group data
func (k Keeper) AddGroupData(ctx sdk.Context, registryRouterAddress ethcommon.Address, group *types.Group) error {
	// Get existing list if it exists
	existingList, exist := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exist {
		// If not found, create a new list with the group
		existingList = &types.GroupList{
			Groups: []*types.Group{group},
		}
	} else {
		// Append new group to existing list
		existingList.Groups = append(existingList.Groups, group)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupDataKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(existingList))

	return nil
}

// SetOperatorSetParams sets the operator set params
func (k Keeper) SetOperatorSetParams(ctx sdk.Context, registryRouterAddress ethcommon.Address, groupNumber uint8, setParam types.OperatorSetParam) error {
	// Get existing group list
	groupList, exists := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exists {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group data not found")
	}

	// Find the specific group
	var targetGroup *types.Group
	for _, group := range groupList.Groups {
		if group.GroupNumber == uint64(groupNumber) {
			targetGroup = group
			break
		}
	}

	if targetGroup == nil {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group not found")
	}

	// Update operator set params
	targetGroup.OperatorSetParam = &setParam

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupDataKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(groupList))

	return nil
}

// SetGroupEjectionParams sets the group ejection params
func (k Keeper) SetGroupEjectionParams(ctx sdk.Context, registryRouterAddress ethcommon.Address, groupNumber uint8, ejectionParams types.GroupEjectionParam) error {
	// Get existing group list
	groupList, exists := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exists {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group data not found")
	}

	// Find the specific group
	var targetGroup *types.Group
	for _, group := range groupList.Groups {
		if group.GroupNumber == uint64(groupNumber) {
			targetGroup = group
			break
		}
	}

	if targetGroup == nil {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group not found")
	}

	// Update ejection params
	targetGroup.GroupEjectionParam = &ejectionParams

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupDataKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(groupList))

	return nil
}

// AddPools adds the strategies
func (k Keeper) AddPools(ctx sdk.Context, registryRouterAddress ethcommon.Address, groupNumber uint8,
	poolParams []*types.PoolParams) error {

	// Get existing group list
	groupList, exists := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exists {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group data not found in store")
	}

	// Find the specific group
	var targetGroup *types.Group
	for _, group := range groupList.Groups {
		if group.GroupNumber == uint64(groupNumber) {
			targetGroup = group
			break
		}
	}

	if targetGroup == nil {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group not found in store")
	}

	// Add new strategies to the group
	targetGroup.PoolParams = append(targetGroup.PoolParams, poolParams...)

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupDataKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(groupList))

	return nil
}

func (k Keeper) RemovePools(ctx sdk.Context, registryRouterAddress ethcommon.Address, groupNumber uint8, IndicesToRemove []*big.Int) error {
	// Get existing group list
	groupList, exists := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exists {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group data not found")
	}

	// Find the specific group
	var targetGroup *types.Group
	for _, group := range groupList.Groups {
		if group.GroupNumber == uint64(groupNumber) {
			targetGroup = group
			break
		}
	}

	if targetGroup == nil {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group not found")
	}

	// Create new slice without removed indices
	newPoolParams := make([]*types.PoolParams, 0)
	for i, strategy := range targetGroup.PoolParams {
		remove := false
		for _, index := range IndicesToRemove {
			if index.Int64() == int64(i) {
				remove = true
				break
			}
		}
		if !remove {
			newPoolParams = append(newPoolParams, strategy)
		}
	}

	targetGroup.PoolParams = newPoolParams

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupDataKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(groupList))

	return nil
}

// ModifyPoolParams modifies the strategy params
func (k Keeper) ModifyPoolParams(ctx sdk.Context, registryRouterAddress ethcommon.Address, groupNumber uint8, Indices, newMultipliers []*big.Int) error {
	// Validate input lengths match
	if len(Indices) != len(newMultipliers) {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "indices and multipliers length mismatch")
	}

	// Get existing group list
	groupList, exists := k.GetGroupDataList(ctx, registryRouterAddress)
	if !exists {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group data not found")
	}

	// Find the specific group
	var targetGroup *types.Group
	for _, group := range groupList.Groups {
		if group.GroupNumber == uint64(groupNumber) {
			targetGroup = group
			break
		}
	}

	if targetGroup == nil {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "group not found")
	}

	// Update multipliers for specified indices
	for i, index := range Indices {
		idx := index.Int64()
		if idx < 0 || idx >= int64(len(targetGroup.PoolParams)) {
			return cosmoserrors.Wrapf(types.ErrInvalidContract, "index out of range")
		}
		targetGroup.PoolParams[idx].Multiplier = newMultipliers[i].Uint64()
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupDataKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(groupList))

	return nil

}
