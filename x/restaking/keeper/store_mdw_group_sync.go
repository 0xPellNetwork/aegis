package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// AddGroupSync adds the quorum sync
func (k *Keeper) AddGroupSync(ctx sdk.Context, txHash string, xmsgIndex string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the existing progress data
	key := types.GroupSyncKey(txHash)
	data := store.Get(key)

	var list types.GroupSyncList
	if data != nil {
		k.cdc.MustUnmarshal(data, &list)
	}

	// Add the new progress
	list.XmsgIndex = append(list.XmsgIndex, xmsgIndex)

	// Store the updated progress list
	store.Set(key, k.cdc.MustMarshal(&list))

	return nil
}

func (k *Keeper) AddGroupSyncs(ctx sdk.Context, txHash string, xmsgIndices []string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the existing progress data
	key := types.GroupSyncKey(txHash)
	data := store.Get(key)

	var list types.GroupSyncList
	if data != nil {
		k.cdc.MustUnmarshal(data, &list)
	}

	// Add the new progresses
	list.XmsgIndex = append(list.XmsgIndex, xmsgIndices...)

	// Store the updated progress list
	store.Set(key, k.cdc.MustMarshal(&list))

	return nil
}

// GetGroupSyncList gets the quorum sync list
func (k *Keeper) GetGroupSyncList(ctx sdk.Context, txHash string) (*types.GroupSyncList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	key := types.GroupSyncKey(txHash)
	data := store.Get(key)
	if data == nil {
		return nil, false
	}

	var list types.GroupSyncList
	k.cdc.MustUnmarshal(data, &list)

	return &list, true
}
