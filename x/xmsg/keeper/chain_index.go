package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// store chain index
func (k Keeper) SetChainIndex(ctx sdk.Context, chainId, height uint64) {
	key := types.KeyPrefix(types.ChainIndexKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), key)
	b := k.cdc.MustMarshal(&types.ChainIndex{
		ChainId:    chainId,
		CurrHeight: height,
	})

	store.Set(types.KeyPrefix(fmt.Sprint(chainId)), b)
}

// get chain index from local store
// if not exist. it will try to sync from observer chainParam
func (k Keeper) GetChainIndex(ctx sdk.Context, chainId uint64) (val types.ChainIndex, exist bool) {
	key := types.KeyPrefix(types.ChainIndexKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), key)

	b := store.Get(types.KeyPrefix(fmt.Sprint(chainId)))
	if b == nil {
		return types.ChainIndex{ChainId: chainId}, false
	}

	k.cdc.MustUnmarshal(b, &val)

	return val, true
}

// store block proof
func (k Keeper) SetBlockProof(ctx sdk.Context, blockProof *types.BlockProof) {
	key := types.KeyPrefix(fmt.Sprint(types.BlockProofKey, blockProof.ChainId))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), key)
	b := k.cdc.MustMarshal(blockProof)

	store.Set(types.KeyPrefix(fmt.Sprint(blockProof.BlockHeight)), b)
}

// get block proof from chainId
func (k Keeper) GetBlockProof(ctx sdk.Context, chainId, height uint64) (val types.BlockProof, exist bool) {
	key := types.KeyPrefix(fmt.Sprint(types.BlockProofKey, chainId))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), key)

	b := store.Get(types.KeyPrefix(fmt.Sprint(height)))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// delete block proof
func (k Keeper) DeleteBlockProof(ctx sdk.Context, chainId, height uint64) {
	key := types.KeyPrefix(fmt.Sprint(types.BlockProofKey, chainId))
	store := prefix.NewStore(ctx.KVStore(k.storeKey), key)

	store.Delete(types.KeyPrefix(fmt.Sprint(height)))
}
