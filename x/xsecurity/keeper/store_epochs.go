package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

var EPOCH_NUMBER_KEY = []byte{0}

const DEFAULT_BLOCKS_PER_EPOCH = 10

// SetEpochNumber sets the epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpochNumber)

	store.Set(EPOCH_NUMBER_KEY, k.cdc.MustMarshal(&types.EpochNumber{EpochNumber: epochNumber}))
}

// GetEpochNumber gets the epoch number
// Returns the current epoch number and a boolean indicating if it exists in the store
func (k Keeper) GetEpochNumber(ctx sdk.Context) (epochNumber uint64, exists bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpochNumber)

	b := store.Get(EPOCH_NUMBER_KEY)
	if len(b) == 0 {
		return 0, false
	}

	var epochNumberObj types.EpochNumber
	k.cdc.MustUnmarshal(b, &epochNumberObj)

	return epochNumberObj.EpochNumber, true
}

// GetBlocksPerEpoch get blocks per epoch
// Returns the number of blocks per epoch and a boolean indicating if it exists in the store
func (k Keeper) GetBlocksPerEpoch(ctx sdk.Context) (blocksPerEpoch uint64, exist bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.KeyBlocksPerEpoch))
	if len(bz) == 0 {
		return DEFAULT_BLOCKS_PER_EPOCH, false
	}

	return sdk.BigEndianToUint64(bz), true
}

// SetBlocksPerEpoch set blocks per epoch
func (k Keeper) SetBlocksPerEpoch(ctx sdk.Context, blocks uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(types.KeyBlocksPerEpoch), sdk.Uint64ToBigEndian(blocks))
}
