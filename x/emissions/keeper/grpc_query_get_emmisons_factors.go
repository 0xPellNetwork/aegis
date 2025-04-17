package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/emissions/types"
)

func (k Keeper) GetEmissionsFactors(goCtx context.Context, _ *types.QueryGetEmissionsFactorsRequest) (*types.QueryGetEmissionsFactorsResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)
	reservesFactor, bondFactor, durationFactor := k.GetBlockRewardComponents(ctx)
	return &types.QueryGetEmissionsFactorsResponse{
		ReservesFactor: reservesFactor.String(),
		BondFactor:     bondFactor.String(),
		DurationFactor: durationFactor.String(),
	}, nil
}
