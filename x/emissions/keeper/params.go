package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/emissions/types"
)

// GetParamsIfExists get all parameters as types.Params if they exist
// non existent parameters will return zero values
func (k Keeper) GetParamsIfExists(ctx sdk.Context) (params types.Params) {
	k.paramStore.GetParamSetIfExists(ctx, &params)
	return
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
