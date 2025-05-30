package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/authority/types"
)

// Policies queries policies
func (k Keeper) Policies(c context.Context, req *types.QueryGetPoliciesRequest) (*types.QueryPoliciesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// fetch policies
	policies, found := k.GetPolicies(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "policies not found")
	}

	return &types.QueryPoliciesResponse{Policies: policies}, nil
}
