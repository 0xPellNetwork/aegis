package keeper

import (
	"encoding/hex"

	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// GetGroupOperatorRegistrationList returns the group operator registration list for the given registry router address
func (k *Keeper) GetGroupOperatorRegistrationList(ctx sdk.Context, registryRouterAddr ethcommon.Address) (*types.GroupOperatorRegistrationListV2, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the operator registration list for the given registry router address
	key := types.GroupOperatorKey(registryRouterAddr)
	data := store.Get(key)
	if data == nil {
		return nil, false
	}

	var registrationList types.GroupOperatorRegistrationListV2
	k.cdc.MustUnmarshal(data, &registrationList)

	return &registrationList, true
}

// AddGroupOperatorRegistration adds a group operator registration to the list for the given registry router address
func (k *Keeper) AddGroupOperatorRegistration(ctx sdk.Context, registryRouterAddr ethcommon.Address, registration *types.GroupOperatorRegistrationV2) error {
	// Get existing list if it exists
	existingList, exist := k.GetGroupOperatorRegistrationList(ctx, registryRouterAddr)
	if !exist {
		// If not found, create a new list with the registration
		existingList = &types.GroupOperatorRegistrationListV2{
			OperatorRegisteredInfos: []*types.GroupOperatorRegistrationV2{registration},
		}
	} else {
		// Append new registration to existing list
		existingList.OperatorRegisteredInfos = append(existingList.OperatorRegisteredInfos, registration)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupOperatorKey(registryRouterAddr)
	store.Set(key, k.cdc.MustMarshal(existingList))

	return nil
}

// RemoveGroupOperatorRegistration removes a group operator registration from the list for the given registry router address
func (k *Keeper) RemoveGroupOperatorRegistration(ctx sdk.Context, registryRouterAddr, operatorAddr ethcommon.Address) error {
	// Get existing list
	existingList, exist := k.GetGroupOperatorRegistrationList(ctx, registryRouterAddr)
	if !exist {
		return cosmoserrors.Wrapf(types.ErrContractNotFound, "group operator registration list not found for registry router address %s", registryRouterAddr)
	}

	// Find and remove the registration for the given operator address
	found := false
	for i, reg := range existingList.OperatorRegisteredInfos {
		if reg.Operator == operatorAddr.String() {
			// Remove the item by shifting elements
			existingList.OperatorRegisteredInfos = append(
				existingList.OperatorRegisteredInfos[:i],
				existingList.OperatorRegisteredInfos[i+1:]...,
			)
			found = true
			break
		}
	}

	if !found {
		return cosmoserrors.Wrapf(types.ErrContractNotFound, "group operator registration not found for operator address %s", operatorAddr)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupOperatorKey(registryRouterAddr)
	store.Set(key, k.cdc.MustMarshal(existingList))

	return nil
}

// RemoveGroupOperatorRegistrationByIds removes a group operator registration from the list for the given registry router address
func (k *Keeper) RemoveGroupOperatorRegistrationByIds(ctx sdk.Context, registryRouterAddr ethcommon.Address, operatorIds [][][32]byte) error {
	// Get existing list
	existingList, exist := k.GetGroupOperatorRegistrationList(ctx, registryRouterAddr)
	if !exist {
		return cosmoserrors.Wrapf(types.ErrContractNotFound, "group operator registration list not found for registry router address %s", registryRouterAddr)
	}

	// Filter out registrations with matching operator IDs
	newRegistrations := make([]*types.GroupOperatorRegistrationV2, 0)
	for _, reg := range existingList.OperatorRegisteredInfos {
		matchFound := false
		for _, groupOperatorIds := range operatorIds {
			for _, operatorId := range groupOperatorIds {
				if hex.EncodeToString(reg.OperatorId) == hex.EncodeToString(operatorId[:]) {
					matchFound = true
					break
				}
			}
			if matchFound {
				break
			}
		}
		if !matchFound {
			newRegistrations = append(newRegistrations, reg)
		}
	}

	// Update the list with filtered registrations
	existingList.OperatorRegisteredInfos = newRegistrations

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupOperatorKey(registryRouterAddr)
	store.Set(key, k.cdc.MustMarshal(existingList))

	return nil
}
