package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// QueryOperatorWeightedShares return the operator weighted shares
func (k Keeper) QueryOperatorWeightedShares(goCtx context.Context, req *types.QueryOperatorWeightedSharesRequest) (*types.QueryOperatorWeightedSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	data, exist := k.GetOperatorWeightedShareList(ctx)
	if !exist {
		return nil, errors.New("data not found")
	}

	return &types.QueryOperatorWeightedSharesResponse{OperatorWeightedShares: data.OperatorWeightedShares}, nil
}
