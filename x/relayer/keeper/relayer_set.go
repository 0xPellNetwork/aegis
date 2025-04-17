package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) SetObserverSet(ctx context.Context, om types.RelayerSet) {
	store := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.KeyPrefix(types.RelayerSetKey))
	b := k.cdc.MustMarshal(&om)
	store.Set([]byte{0}, b)
}

func (k Keeper) GetObserverSet(ctx context.Context) (val types.RelayerSet, found bool) {
	store := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.KeyPrefix(types.RelayerSetKey))
	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) IsAddressPartOfObserverSet(ctx sdk.Context, address string) bool {
	observerSet, found := k.GetObserverSet(ctx)
	if !found {
		return false
	}
	for _, addr := range observerSet.RelayerList {
		if addr == address {
			return true
		}
	}
	return false

}

func (k Keeper) AddObserverToSet(ctx sdk.Context, address string) {
	observerSet, found := k.GetObserverSet(ctx)
	if !found {
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{address},
		})
		return
	}
	for _, addr := range observerSet.RelayerList {
		if addr == address {
			return
		}
	}
	observerSet.RelayerList = append(observerSet.RelayerList, address)
	k.SetObserverSet(ctx, observerSet)
}

func (k Keeper) RemoveObserverFromSet(ctx context.Context, address string) {
	observerSet, found := k.GetObserverSet(ctx)
	if !found {
		return
	}
	for i, addr := range observerSet.RelayerList {
		if addr == address {
			observerSet.RelayerList = append(observerSet.RelayerList[:i], observerSet.RelayerList[i+1:]...)
			k.SetObserverSet(ctx, observerSet)
			return
		}
	}
}

func (k Keeper) UpdateObserverAddress(ctx context.Context, oldObserverAddress, newObserverAddress string) error {
	observerSet, found := k.GetObserverSet(ctx)
	if !found {
		return types.ErrObserverSetNotFound
	}
	for i, addr := range observerSet.RelayerList {
		if addr == oldObserverAddress {
			observerSet.RelayerList[i] = newObserverAddress
			k.SetObserverSet(ctx, observerSet)
			return nil
		}
	}
	return types.ErrUpdateObserver
}
