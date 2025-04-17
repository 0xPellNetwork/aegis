package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// GasRechargeOperationIndex queries the gas token increment index
func (k Keeper) GasRechargeOperationIndex(c context.Context, req *types.QueryGasRechargeOperationIndexRequest) (*types.QueryGasRechargeOperationIndexResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	index := k.GetGasRechargeOperationIndex(ctx, req.ChainId)
	if chain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, req.ChainId); chain != nil {
		return &types.QueryGasRechargeOperationIndexResponse{
			ChainId:   uint64(req.ChainId),
			CurrIndex: index,
		}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "gas token increment index not exist")
}
