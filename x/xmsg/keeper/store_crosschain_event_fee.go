package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// StoreCrosschainEventFee stores a CrosschainEventFee object in the KVStore using chainId as the key
func (k Keeper) StoreCrosschainEventFee(ctx sdk.Context, crosschainEventFee types.CrosschainFeeParam) {
	store := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.KeyPrefix(types.KeyPrefixCrosschainFeeParam))
	b := k.cdc.MustMarshal(&crosschainEventFee)

	store.Set([]byte(fmt.Sprint(crosschainEventFee.ChainId)), b)
}

// GetCrosschainEventFee retrieves a CrosschainEventFee object from the KVStore by chainId
func (k Keeper) GetCrosschainEventFee(ctx sdk.Context, chainId int64) (types.CrosschainFeeParam, bool) {
	store := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.KeyPrefix(types.KeyPrefixCrosschainFeeParam))
	b := store.Get([]byte(fmt.Sprint(chainId)))
	if b == nil {
		return types.CrosschainFeeParam{}, false
	}

	var crosschainEventFee types.CrosschainFeeParam
	k.cdc.MustUnmarshal(b, &crosschainEventFee)
	return crosschainEventFee, true
}

// GetAllCrosschainEventFees retrieves all CrosschainEventFee objects stored in the KVStore
func (k Keeper) GetAllCrosschainEventFees(ctx sdk.Context) ([]types.CrosschainFeeParam, error) {
	store := prefix.NewStore(sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey), types.KeyPrefix(types.KeyPrefixCrosschainFeeParam))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var crosschainEventFees []types.CrosschainFeeParam
	for ; iterator.Valid(); iterator.Next() {
		var crosschainEventFee types.CrosschainFeeParam
		k.cdc.MustUnmarshal(iterator.Value(), &crosschainEventFee)
		crosschainEventFees = append(crosschainEventFees, crosschainEventFee)
	}

	return crosschainEventFees, nil
}
