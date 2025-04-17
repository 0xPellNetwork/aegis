package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// PellRechargeOperationIndex queries the pell token increment index
func (k Keeper) PellRechargeOperationIndex(c context.Context, req *types.QueryPellRechargeOperationIndexRequest) (*types.QueryPellRechargeOperationIndexResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	index := k.GetPellRechargeOperationIndex(ctx, req.ChainId)
	if chain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, req.ChainId); chain != nil {
		return &types.QueryPellRechargeOperationIndexResponse{
			ChainId:   uint64(req.ChainId),
			CurrIndex: index,
		}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "pell token increment index not exist")
}
