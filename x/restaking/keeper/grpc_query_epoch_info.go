package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// GetEpochInfo returns the epoch info
func (k Keeper) GetEpochInfo(ctx context.Context, req *types.QueryEpochInfoRequest) (*types.QueryGetEpochInfoResponse, error) {
	return &types.QueryGetEpochInfoResponse{BlockNumber: k.GetBlocksPerEpoch(sdk.UnwrapSDKContext(ctx))}, nil
}
