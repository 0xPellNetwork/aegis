package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (k Keeper) BlameByIdentifier(goCtx context.Context, request *types.QueryBlameByIdentifierRequest) (*types.QueryBlameByIdentifierResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	blame, found := k.GetBlame(ctx, request.BlameIdentifier)
	if !found {
		return nil, status.Error(codes.NotFound, "blame info not found")
	}

	return &types.QueryBlameByIdentifierResponse{
		BlameInfo: &blame,
	}, nil
}

func (k Keeper) GetAllBlameRecords(goCtx context.Context, request *types.QueryAllBlameRecordsRequest) (*types.QueryGetAllBlameRecordsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	blameRecords, pageRes, err := k.GetAllBlamePaginated(ctx, request.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetAllBlameRecordsResponse{
		BlameInfo:  blameRecords,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) BlamesByChainAndNonce(goCtx context.Context, request *types.QueryBlameByChainAndNonceRequest) (*types.QueryBlamesByChainAndNonceResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	blameRecords, found := k.GetBlamesByChainAndNonce(ctx, request.ChainId, request.Nonce)
	if !found {
		return nil, status.Error(codes.NotFound, "blame info not found")
	}
	return &types.QueryBlamesByChainAndNonceResponse{
		BlameInfo: blameRecords,
	}, nil
}
