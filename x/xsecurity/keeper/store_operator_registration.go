package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// GetOperatorRegistrationList returns the operator registration list for the given registry router address
func (k Keeper) GetOperatorRegistrationList(ctx sdk.Context) (*types.LSTOperatorRegistrationList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTOperatorRegistrationListKey))

	// Get the operator registration list for the given registry router address
	data := store.Get([]byte(types.LSTOperatorRegistrationListKey))
	if len(data) == 0 {
		return nil, false
	}

	var registrationList types.LSTOperatorRegistrationList
	k.cdc.MustUnmarshal(data, &registrationList)

	return &registrationList, true
}

// AddOperatorRegistration adds a new operator registration to the operator registration list
func (k Keeper) AddOperatorRegistration(ctx sdk.Context, registration *types.LSTOperatorRegistration) {
	// Get existing list if it exists
	existingList, exist := k.GetOperatorRegistrationList(ctx)
	if !exist {
		// If not found, create a new list with the registration
		existingList = &types.LSTOperatorRegistrationList{
			OperatorRegistrations: []*types.LSTOperatorRegistration{registration},
		}
	} else {
		// Append new registration to existing list
		existingList.OperatorRegistrations = append(existingList.OperatorRegistrations, registration)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTOperatorRegistrationListKey))
	key := []byte(types.LSTOperatorRegistrationListKey)
	store.Set(key, k.cdc.MustMarshal(existingList))
}
