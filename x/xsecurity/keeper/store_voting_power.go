package keeper

import (
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// SetLSTVotingPowerRatio sets the LST voting power ratio
// This is used to calculate the voting power of LST tokens
// The ratio is a uint64 value that represents the percentage of voting power
// such as ratio=20 means 20% of the voting power
func (k Keeper) SetLSTVotingPowerRatio(ctx sdk.Context, numerator, denominator math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.LSTVotingPowerRatioKey)

	data := &types.LSTVotingPowerRatio{
		Numerator:   numerator,
		Denominator: denominator,
	}
	store.Set(types.LSTVotingPowerRatioKey, k.cdc.MustMarshal(data))
}

// GetLSTVotingPowerRatio gets the LST voting power ratio
func (k Keeper) GetLSTVotingPowerRatio(ctx sdk.Context) (*types.LSTVotingPowerRatio, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.LSTVotingPowerRatioKey)

	b := store.Get(types.LSTVotingPowerRatioKey)
	if len(b) == 0 {
		return nil, false
	}

	var data types.LSTVotingPowerRatio
	k.cdc.MustUnmarshal(b, &data)

	return &data, true
}

// SetLastNativeVotingPower sets the last native voting power value in the store
func (k Keeper) SetLastNativeVotingPower(ctx sdk.Context, votingPower int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.LastNativeVotingPowerKey)

	store.Set(types.LastNativeVotingPowerKey, sdk.Uint64ToBigEndian(uint64(votingPower)))
}

// GetLastNativeVotingPower retrieves the last native voting power value from the store
// Returns the voting power value and a boolean indicating if the value exists
func (k Keeper) GetLastNativeVotingPower(ctx sdk.Context) (votingPower int64, exists bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.LastNativeVotingPowerKey)

	bz := store.Get(types.LastNativeVotingPowerKey)
	if len(bz) == 0 {
		return 0, false
	}

	// never overflow
	return int64(sdk.BigEndianToUint64(bz)), true
}
