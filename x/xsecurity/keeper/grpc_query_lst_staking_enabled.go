package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// QueryLSTStakingEnabled is a gRPC query handler for the LST staking enabled status.
func (k Keeper) QueryLSTStakingEnabled(goCtx context.Context, req *types.QueryLSTStakingEnabledRequest) (*types.QueryLSTStakingEnabledResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	info, exist := k.GetLSTStakingEnabled(ctx)
	if !exist {
		return nil, errors.New("data not found")
	}

	return &types.QueryLSTStakingEnabledResponse{
		LstStakingEnabled: info.Enabled,
	}, nil
}
