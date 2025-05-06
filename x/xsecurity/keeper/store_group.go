package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// SetGroupInfo stores the LSTGroupInfo
func (k Keeper) SetGroupInfo(ctx sdk.Context, groupInfo *types.LSTGroupInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTGroupInfoKey))

	store.Set([]byte(types.LSTGroupInfoKey), k.cdc.MustMarshal(groupInfo))
}

// GetGroupInfo returns the LSTGroupInfo
func (k Keeper) GetGroupInfo(ctx sdk.Context) (*types.LSTGroupInfo, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTGroupInfoKey))

	bz := store.Get([]byte(types.LSTGroupInfoKey))
	if len(bz) == 0 {
		return nil, false
	}

	data := new(types.LSTGroupInfo)
	k.cdc.MustUnmarshal(bz, data)

	return data, true
}
