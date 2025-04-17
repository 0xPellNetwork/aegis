package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// CrosschainFeeParams returns all crosschain fee params
func (k Keeper) CrosschainFeeParams(ctx context.Context, req *types.QueryCrosschainFeeParamsRequest) (*types.QueryCrosschainFeeParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	crosschainEventFees, err := k.GetAllCrosschainEventFees(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get all crosschain event fees: %v", err)
	}

	res := make([]*types.CrosschainFeeParam, len(crosschainEventFees))
	for i, fee := range crosschainEventFees {
		res[i] = &fee
	}

	return &types.QueryCrosschainFeeParamsResponse{
		CrosschainFeeParams: res,
	}, nil
}

// CrosschainFeeParamByChainId returns the crosschain fee param for a given chain id
func (k Keeper) CrosschainFeeParamByChainId(ctx context.Context, req *types.QueryCrosschainFeeParamByChainIdRequest) (*types.QueryCrosschainFeeParamByChainIdResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	crosschainEventFee, found := k.GetCrosschainEventFee(sdkCtx, req.ChainId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "crosschain event fee not found for chain id: %d", req.ChainId)
	}

	return &types.QueryCrosschainFeeParamByChainIdResponse{
		CrosschainFeeParam: &crosschainEventFee,
	}, nil
}
