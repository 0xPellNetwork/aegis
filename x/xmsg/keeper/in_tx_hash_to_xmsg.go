package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// SetInTxHashToXmsg set a specific inTxHashToXmsg in the store from its index
func (k Keeper) SetInTxHashToXmsg(ctx sdk.Context, inTxHashToXmsg types.InTxHashToXmsg) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InTxHashToXmsgKeyPrefix))
	b := k.cdc.MustMarshal(&inTxHashToXmsg)
	store.Set(types.InTxHashToXmsgKey(
		inTxHashToXmsg.InTxHash,
	), b)
}

// GetInTxHashToXmsg returns a inTxHashToXmsg from its index
func (k Keeper) GetInTxHashToXmsg(
	ctx sdk.Context,
	inTxHash string,

) (val types.InTxHashToXmsg, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InTxHashToXmsgKeyPrefix))

	b := store.Get(types.InTxHashToXmsgKey(
		inTxHash,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveInTxHashToXmsg removes a inTxHashToXmsg from the store
func (k Keeper) RemoveInTxHashToXmsg(
	ctx sdk.Context,
	inTxHash string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InTxHashToXmsgKeyPrefix))
	store.Delete(types.InTxHashToXmsgKey(
		inTxHash,
	))
}

// GetAllInTxHashToXmsg returns all inTxHashToXmsg
func (k Keeper) GetAllInTxHashToXmsg(ctx sdk.Context) (list []types.InTxHashToXmsg) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.InTxHashToXmsgKeyPrefix))
	iterator := store.Iterator(nil, nil)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.InTxHashToXmsg
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
