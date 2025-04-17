package keeper

import (
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// MigrationStore is used to migrate GroupOperatorRegistrationList from v1.1 to v1.2
// Mainly for the data of operator and group registration relationships in dvs, the old version had a bug that used uint64 data type to store BLS public key GPoint information
// The new version has changed to sdk.Int type
func (k Keeper) MigrationStore(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating GroupOperatorRegistrationList from v1.1 to v1.2")

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var oldList types.GroupOperatorRegistrationList
		if err := k.cdc.Unmarshal(iterator.Value(), &oldList); err != nil {
			ctx.Logger().Error("Failed to unmarshal old GroupOperatorRegistrationList", "error", err)
			continue
		}

		newList := groupOperatorRegistrationListV1ToV2(oldList)

		store.Set(iterator.Key(), k.cdc.MustMarshal(newList))
		ctx.Logger().Info("Migrated GroupOperatorRegistrationList", "registryRouterAddr", common.BytesToAddress(iterator.Key()), "oldCount", len(oldList.OperatorRegisteredInfos), "newCount", len(newList.OperatorRegisteredInfos))
	}

	ctx.Logger().Info("Migration complete")
	return nil
}

func groupOperatorRegistrationListV1ToV2(oldList types.GroupOperatorRegistrationList) *types.GroupOperatorRegistrationListV2 {
	newList := &types.GroupOperatorRegistrationListV2{
		OperatorRegisteredInfos: make([]*types.GroupOperatorRegistrationV2, 0, len(oldList.OperatorRegisteredInfos)),
	}

	for _, oldRegistration := range oldList.OperatorRegisteredInfos {
		// Skip if oldRegistration or PubkeyParams is nil
		if oldRegistration == nil || oldRegistration.PubkeyParams == nil {
			continue
		}

		// Skip if PubkeyG1 or PubkeyG2 is nil
		if oldRegistration.PubkeyParams.PubkeyG1 == nil || oldRegistration.PubkeyParams.PubkeyG2 == nil {
			continue
		}

		newRegistration := &types.GroupOperatorRegistrationV2{
			Operator:     oldRegistration.Operator,
			OperatorId:   oldRegistration.OperatorId,
			GroupNumbers: oldRegistration.GroupNumbers,
			Socket:       oldRegistration.Socket,
			PubkeyParams: &types.PubkeyRegistrationParamsV2{
				PubkeyG1: &types.G1PointV2{
					X: sdkmath.NewIntFromUint64(oldRegistration.PubkeyParams.PubkeyG1.X),
					Y: sdkmath.NewIntFromUint64(oldRegistration.PubkeyParams.PubkeyG1.Y),
				},
				PubkeyG2: &types.G2PointV2{
					X: convertUint64SliceToInt(oldRegistration.PubkeyParams.PubkeyG2.X),
					Y: convertUint64SliceToInt(oldRegistration.PubkeyParams.PubkeyG2.Y),
				},
			},
		}
		newList.OperatorRegisteredInfos = append(newList.OperatorRegisteredInfos, newRegistration)
	}

	return newList
}

func convertUint64SliceToInt(arr []uint64) []sdkmath.Int {
	intArr := make([]sdkmath.Int, len(arr))
	for i, v := range arr {
		intArr[i] = sdkmath.NewIntFromUint64(v)
	}

	return intArr
}

// AddGroupOperatorRegistrationV1 adds a group operator registration to the list for the given registry router address
// This function is the old version, mainly used for unit tests to write the old version of the operator registration relationship data, because the new code does not retain the old version of the write method
// This function must be deleted in the next version of the code
// TODO: This function should be removed after the upgrade
func (k *Keeper) AddGroupOperatorRegistrationV1(ctx sdk.Context, registryRouterAddr common.Address, registration *types.GroupOperatorRegistration) error {
	ctx.Logger().Info("AddGroupOperatorRegistrationV1", "registryRouterAddr", registryRouterAddr, "registration", registration)

	existingList := &types.GroupOperatorRegistrationList{
		OperatorRegisteredInfos: []*types.GroupOperatorRegistration{registration},
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.GroupOperatorKey(registryRouterAddr)
	store.Set(key, k.cdc.MustMarshal(existingList))

	return nil
}

// GetGroupOperatorRegistrationListV1 returns the group operator registration list for the given registry router address
// This function is the old version, mainly used for unit tests to read the old version of the operator registration relationship data, because the new code does not retain the old version of the read method
// This function must be deleted in the next version of the code
// TODO: This function should be removed after the upgrade
func (k *Keeper) GetGroupOperatorRegistrationListV1(ctx sdk.Context, registryRouterAddr common.Address) (*types.GroupOperatorRegistrationList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the operator registration list for the given registry router address
	key := types.GroupOperatorKey(registryRouterAddr)
	data := store.Get(key)
	if data == nil {
		return nil, false
	}

	var registrationList types.GroupOperatorRegistrationList
	k.cdc.MustUnmarshal(data, &registrationList)

	return &registrationList, true
}
