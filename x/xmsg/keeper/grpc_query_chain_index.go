package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func (k Keeper) ChainIndex(c context.Context, req *types.QueryChainIndexRequest) (*types.QueryChainIndexResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chainIndex, exist := k.GetChainIndex(ctx, uint64(req.ChainId))
	if exist {
		return &types.QueryChainIndexResponse{
			ChainId:    chainIndex.ChainId,
			CurrHeight: chainIndex.CurrHeight,
		}, nil
	}

	if chain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, req.ChainId); chain != nil {
		return &types.QueryChainIndexResponse{
			ChainId:    uint64(req.ChainId),
			CurrHeight: 0,
		}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "chain index not exist")
}
