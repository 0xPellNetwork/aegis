package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) CrosschainFlags(c context.Context, req *types.QueryGetCrosschainFlagsRequest) (*types.QueryCrosschainFlagsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetCrosschainFlags(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryCrosschainFlagsResponse{CrosschainFlags: val}, nil
}
