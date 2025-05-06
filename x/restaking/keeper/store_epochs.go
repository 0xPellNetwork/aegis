package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

var EPOCH_NUMBER_KEY = []byte{0}

const DEFAULT_BLOCKS_PER_EPOCH = 10

// SetEpochNumber sets the epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyPrefixEpochNumber)))

	store.Set(EPOCH_NUMBER_KEY, k.cdc.MustMarshal(&types.EpochNumber{EpochNumber: epochNumber}))
}

// GetEpochNumber gets the epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyPrefixEpochNumber)))

	b := store.Get(EPOCH_NUMBER_KEY)
	if b == nil {
		return 0
	}

	var epochNumber types.EpochNumber
	k.cdc.MustUnmarshal(b, &epochNumber)

	return epochNumber.EpochNumber
}

// GetBlocksPerEpoch get blocks per epoch
func (k Keeper) GetBlocksPerEpoch(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.KeyBlocksPerEpoch))
	if bz == nil {
		return DEFAULT_BLOCKS_PER_EPOCH
	}
	return sdk.BigEndianToUint64(bz)
}

// SetBlocksPerEpoch set blocks per epoch
func (k Keeper) SetBlocksPerEpoch(ctx sdk.Context, blocks uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.KeyBlocksPerEpoch), sdk.Uint64ToBigEndian(blocks))
}
