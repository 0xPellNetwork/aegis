package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pell-chain/pellcore/x/pevm/types"
)

func (k Keeper) SystemContract(c context.Context, req *types.QueryGetSystemContractRequest) (*types.SystemContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetSystemContract(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.SystemContractResponse{SystemContract: val}, nil
}
