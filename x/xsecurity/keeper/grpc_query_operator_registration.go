package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// QueryOperatorRegistrationList return the operator registration list
func (k Keeper) QueryOperatorRegistrationList(goCtx context.Context, req *types.QueryOperatorRegistrationListRequest) (*types.QueryOperatorRegistrationListResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	data, exist := k.GetOperatorRegistrationList(ctx)
	if !exist {
		return nil, errors.New("data not found")
	}

	return &types.QueryOperatorRegistrationListResponse{OperatorRegistrations: data.OperatorRegistrations}, nil
}
