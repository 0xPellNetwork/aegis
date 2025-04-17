package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// set epoch operator shares sync txs
func (k Keeper) SetEpochOperatorSharesSyncTxs(ctx sdk.Context, chainId uint64, epochNumber uint64, xmsgIndexes []string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyEpochOperatorSharesSyncTxs)))

	b := k.cdc.MustMarshal(&types.EpochOperatorSharesSyncTxs{
		PendingXmsgIndexes: xmsgIndexes,
	})

	store.Set(types.KeyPrefix(fmt.Sprint(chainId, epochNumber)), b)
}

// get epoch operator shares sync txs
func (k Keeper) GetEpochOperatorSharesSyncTxs(ctx sdk.Context, chainId uint64, epochNumber uint64) (*types.EpochOperatorSharesSyncTxs, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyEpochOperatorSharesSyncTxs)))

	b := store.Get(types.KeyPrefix(fmt.Sprint(chainId, epochNumber)))
	if b == nil {
		return nil, false
	}

	var syncEpochTx types.EpochOperatorSharesSyncTxs
	k.cdc.MustUnmarshal(b, &syncEpochTx)

	return &syncEpochTx, true
}

// delete epoch operator shares sync txs
func (k Keeper) DeleteEpochOperatorSharesSyncTxs(ctx sdk.Context, chainId uint64, epochNumber uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyEpochOperatorSharesSyncTxs)))
	store.Delete(types.KeyPrefix(fmt.Sprint(chainId, epochNumber)))
}
