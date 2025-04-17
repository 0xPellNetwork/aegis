package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// SetLSTRegistryRouterAddress stores (or overwrites) the unique LSTRegistryRouterAddress
func (k Keeper) SetLSTRegistryRouterAddress(ctx sdk.Context, addr *types.LSTRegistryRouterAddress) {
	// Construct a store with a prefix
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTRegistryRouterAddressKey))

	// Serialize the struct to bytes
	bz := k.cdc.MustMarshal(addr)

	// Store with a fixed key
	store.Set([]byte(types.LSTRegistryRouterAddressKey), bz)
}

// GetLSTRegistryRouterAddress reads the stored unique LSTRegistryRouterAddress
// If it does not exist, the second return value is false
func (k Keeper) GetLSTRegistryRouterAddress(ctx sdk.Context) (*types.LSTRegistryRouterAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.LSTRegistryRouterAddressKey))

	bz := store.Get([]byte(types.LSTRegistryRouterAddressKey))
	if len(bz) == 0 {
		// Return false if it does not exist
		return nil, false
	}

	var addr types.LSTRegistryRouterAddress
	// Deserialize
	k.cdc.MustUnmarshal(bz, &addr)

	return &addr, true
}
