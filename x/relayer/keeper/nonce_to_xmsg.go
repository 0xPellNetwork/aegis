package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// NonceToXmsg methods
// The object stores the mapping from nonce to cross chain tx

func (k Keeper) RemoveNonceToXmsg(ctx sdk.Context, nonceToXmsg types.NonceToXmsg) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.NonceToXmsgKeyPrefix))
	store.Delete(types.KeyPrefix(fmt.Sprintf("%s-%d-%d", nonceToXmsg.Tss, nonceToXmsg.ChainId, nonceToXmsg.Nonce)))
}

func (k Keeper) SetNonceToXmsg(ctx sdk.Context, nonceToXmsg types.NonceToXmsg) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.NonceToXmsgKeyPrefix))
	b := k.cdc.MustMarshal(&nonceToXmsg)
	store.Set(types.KeyPrefix(fmt.Sprintf("%s-%d-%d", nonceToXmsg.Tss, nonceToXmsg.ChainId, nonceToXmsg.Nonce)), b)
}

func (k Keeper) GetNonceToXmsg(ctx sdk.Context, tss string, chainID int64, nonce int64) (val types.NonceToXmsg, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.NonceToXmsgKeyPrefix))

	b := store.Get(types.KeyPrefix(fmt.Sprintf("%s-%d-%d", tss, chainID, nonce)))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetAllNonceToXmsg(ctx sdk.Context) (list []types.NonceToXmsg) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.NonceToXmsgKeyPrefix))
	iterator := store.Iterator(nil, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var val types.NonceToXmsg
		err := k.cdc.Unmarshal(iterator.Value(), &val)
		if err == nil {
			list = append(list, val)
		}

	}

	return

}
