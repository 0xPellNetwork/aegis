package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) Keygen(c context.Context, _ *types.QueryGetKeygenRequest) (*types.QueryKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	val, found := k.GetKeygen(ctx)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}
	return &types.QueryKeygenResponse{Keygen: &val}, nil
}
