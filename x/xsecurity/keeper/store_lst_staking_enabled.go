package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// SetLSTStakingEnabled sets the LST staking enabled status
func (k Keeper) SetLSTStakingEnabled(ctx sdk.Context, enabled bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.LSTStakingEnabledKey)

	data := &types.LSTStakingEnabled{Enabled: enabled}
	store.Set(types.LSTStakingEnabledKey, k.cdc.MustMarshal(data))
}

// GetLSTStakingEnabled gets the LST staking enabled status
func (k Keeper) GetLSTStakingEnabled(ctx sdk.Context) (*types.LSTStakingEnabled, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.LSTStakingEnabledKey)

	b := store.Get(types.LSTStakingEnabledKey)
	if len(b) == 0 {
		return nil, false
	}

	data := new(types.LSTStakingEnabled)
	k.cdc.MustUnmarshal(b, data)

	return data, true
}
