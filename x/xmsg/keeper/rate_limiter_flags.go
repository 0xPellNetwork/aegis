package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// SetRateLimiterFlags set the rate limiter flags in the store
func (k Keeper) SetRateLimiterFlags(ctx sdk.Context, rateLimiterFlags types.RateLimiterFlags) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RateLimiterFlagsKey))
	b := k.cdc.MustMarshal(&rateLimiterFlags)
	store.Set([]byte{0}, b)
}

// GetRateLimiterFlags returns the rate limiter flags
func (k Keeper) GetRateLimiterFlags(ctx sdk.Context) (val types.RateLimiterFlags, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RateLimiterFlagsKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetRateLimiterAssetRateList returns a list of all foreign asset rate
func (k Keeper) GetRateLimiterAssetRateList(
	ctx sdk.Context,
) (flags types.RateLimiterFlags, found bool) {
	flags, found = k.GetRateLimiterFlags(ctx)
	if !found {
		return flags, false
	}

	return flags, true
}
