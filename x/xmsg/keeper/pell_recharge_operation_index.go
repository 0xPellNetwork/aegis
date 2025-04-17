package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func (k Keeper) SetPellRechargeOperationIndex(ctx sdk.Context, chainID int64, voteIndex uint64) {
	indexBytes := make([]byte, 8) // uint64 size
	binary.BigEndian.PutUint64(indexBytes, voteIndex)

	key := fmt.Sprint(chainID)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FinalizedAddPellTokenKeyPrefix))

	store.Set(types.KeyPrefix(key), indexBytes)
}

func (k Keeper) IsPellRechargeOperationIndexFinalized(ctx sdk.Context, chainID int64, voteIndex uint64) bool {
	key := fmt.Sprint(chainID)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FinalizedAddPellTokenKeyPrefix))

	indexBytes := store.Get(types.KeyPrefix(key))
	if indexBytes == nil || len(indexBytes) < 8 {
		return false
	}

	index := binary.BigEndian.Uint64(indexBytes)
	return voteIndex <= index
}

func (k Keeper) GetPellRechargeOperationIndex(ctx sdk.Context, chainID int64) uint64 {
	key := fmt.Sprint(chainID)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FinalizedAddPellTokenKeyPrefix))

	indexBytes := store.Get(types.KeyPrefix(key))
	if indexBytes == nil || len(indexBytes) < 8 {
		return 0
	}

	return binary.BigEndian.Uint64(indexBytes)
}
