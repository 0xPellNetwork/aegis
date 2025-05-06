package keeper

import (
	"errors"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// AddRegistryRouterAddress adds a registry router address to the quorum data
func (k *Keeper) AddRegistryRouterAddress(ctx sdk.Context, addrs []ethcommon.Address) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	k.SetStakeRegistryRouterAddress(ctx, addrs)

	// Get existing routers
	var routerList types.RegistryRouterList
	existingData := store.Get(types.RegistryRouterKey())
	if existingData != nil {
		k.cdc.MustUnmarshal(existingData, &routerList)
	}

	// Append new addresses
	for _, fromAddr := range addrs {
		// Check if address already exists
		exists := false
		for _, addr := range routerList.Addresses {
			if addr == fromAddr.Hex() {
				exists = true
				break
			}
		}
		if !exists {
			routerList.Addresses = append(routerList.Addresses, fromAddr.Hex())
		}
	}

	// Marshal and store
	bz := k.cdc.MustMarshal(&routerList)
	store.Set(types.RegistryRouterKey(), bz)

	return nil
}

// GetAllRegistryRouterAddresses returns the list of registry routers
func (k *Keeper) GetAllRegistryRouterAddresses(ctx sdk.Context) ([]ethcommon.Address, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	var routerList types.RegistryRouterList
	existingData := store.Get(types.RegistryRouterKey())
	if existingData != nil {
		k.cdc.MustUnmarshal(existingData, &routerList)
	}

	addresses := make([]ethcommon.Address, len(routerList.Addresses))
	for i, addr := range routerList.Addresses {
		addresses[i] = ethcommon.HexToAddress(addr)
	}
	return addresses, nil
}

func (k *Keeper) SetStakeRegistryRouterAddress(ctx sdk.Context, addrs []ethcommon.Address) error {
	if len(addrs) != 2 {
		return nil
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Use the second address as the key and the first address as the value
	key := types.StakeRegistryRouterKey(addrs[1])
	value := addrs[0].Hex()

	// Store the value
	store.Set(key, []byte(value))

	return nil
}

// GetStakeRegistryRouterAddress retrieves the stake registry router address from the store
func (k *Keeper) GetStakeRegistryRouterAddress(ctx sdk.Context, key ethcommon.Address) (ethcommon.Address, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the value using the key
	value := store.Get(types.StakeRegistryRouterKey(key))
	if value == nil {
		return ethcommon.Address{}, errors.New("stake registry router address not found")
	}

	// Convert the value to ethcommon.Address
	address := ethcommon.HexToAddress(string(value))
	return address, nil
}
