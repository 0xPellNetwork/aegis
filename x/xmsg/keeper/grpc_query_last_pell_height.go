package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func (k Keeper) LastPellHeight(goCtx context.Context, req *types.QueryLastPellHeightRequest) (*types.QueryLastPellHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	height := ctx.BlockHeight()
	if height < 0 {
		return nil, status.Error(codes.OutOfRange, "height out of range")
	}
	return &types.QueryLastPellHeightResponse{
		Height: height,
	}, nil
}
