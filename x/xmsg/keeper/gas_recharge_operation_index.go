package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func (k Keeper) SetGasRechargeOperationIndex(ctx sdk.Context, chainID int64, voteIndex uint64) {
	indexBytes := make([]byte, 8) // uint64 size
	binary.BigEndian.PutUint64(indexBytes, voteIndex)

	key := fmt.Sprint(chainID)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FinalizedAddGasTokenKeyPrefix))

	store.Set(types.KeyPrefix(key), indexBytes)
}

func (k Keeper) IsGasRechargeOperationIndexFinalized(ctx sdk.Context, chainID int64, voteIndex uint64) bool {
	key := fmt.Sprint(chainID)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FinalizedAddGasTokenKeyPrefix))

	indexBytes := store.Get(types.KeyPrefix(key))
	if indexBytes == nil || len(indexBytes) < 8 {
		return false
	}

	index := binary.BigEndian.Uint64(indexBytes)
	return voteIndex <= index
}

func (k Keeper) GetGasRechargeOperationIndex(ctx sdk.Context, chainID int64) uint64 {
	key := fmt.Sprint(chainID)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FinalizedAddGasTokenKeyPrefix))

	indexBytes := store.Get(types.KeyPrefix(key))
	if indexBytes == nil || len(indexBytes) < 8 {
		return 0
	}

	return binary.BigEndian.Uint64(indexBytes)
}
