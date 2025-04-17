package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// SetOutboundState stores the outbound state of cross-chain messages in the state store.
// This function serializes the OutboundState object and stores it in the state database.
// The storage uses chainId as the key and organizes data in a prefix store format to
// ensure isolation between different types of data.
func (k Keeper) SetOutboundState(ctx sdk.Context, outboundState *types.EpochOutboundState) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyPrefixOutboundState)))

	store.Set(types.KeyPrefix(fmt.Sprint(outboundState.ChainId)), k.cdc.MustMarshal(outboundState))

	return nil
}

// GetOutboundState retrieves the outbound state of cross-chain messages for a specific chain ID.
// This function retrieves and deserializes the OutboundState object from the state database.
// If no state is found for the given chainId, it returns nil and false.
func (k Keeper) GetOutboundState(ctx sdk.Context, chainID uint64) (*types.EpochOutboundState, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyPrefixOutboundState)))

	bz := store.Get(types.KeyPrefix(fmt.Sprint(chainID)))
	if bz == nil {
		return nil, false
	}

	var outboundState types.EpochOutboundState
	k.cdc.MustUnmarshal(bz, &outboundState)

	return &outboundState, true
}

// GetAllOutboundStates retrieves all outbound states from the state store.
func (k Keeper) GetAllOutboundStates(ctx sdk.Context) ([]*types.EpochOutboundState, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(fmt.Sprintf(types.KeyPrefixOutboundState)))

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	outboundStates := []*types.EpochOutboundState{}
	for ; iter.Valid(); iter.Next() {
		var outboundState types.EpochOutboundState
		k.cdc.MustUnmarshal(iter.Value(), &outboundState)
		outboundStates = append(outboundStates, &outboundState)
	}

	return outboundStates, nil
}
