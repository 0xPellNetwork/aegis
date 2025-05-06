package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func (k Keeper) InTxTrackerAllByChain(goCtx context.Context, request *types.QueryAllInTxTrackerByChainRequest) (*types.QueryInTxTrackerAllByChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var inTxTrackers []types.InTxTracker
	inTxTrackers, pageRes, err := k.GetAllInTxTrackerForChainPaginated(ctx, request.ChainId, request.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryInTxTrackerAllByChainResponse{InTxTrackers: inTxTrackers, Pagination: pageRes}, nil
}

func (k Keeper) InTxTrackerAll(goCtx context.Context, req *types.QueryAllInTxTrackersRequest) (*types.QueryInTxTrackerAllResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var inTxTrackers []types.InTxTracker
	inTxTrackers, pageRes, err := k.GetAllInTxTrackerPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryInTxTrackerAllResponse{InTxTrackers: inTxTrackers, Pagination: pageRes}, nil
}
