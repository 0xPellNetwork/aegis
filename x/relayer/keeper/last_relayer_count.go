package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) SetLastObserverCount(ctx sdk.Context, lbc *types.LastRelayerCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LastBlockObserverCountKey))
	b := k.cdc.MustMarshal(lbc)
	store.Set([]byte{0}, b)
}

func (k Keeper) GetLastObserverCount(ctx sdk.Context) (val types.LastRelayerCount, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LastBlockObserverCountKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
