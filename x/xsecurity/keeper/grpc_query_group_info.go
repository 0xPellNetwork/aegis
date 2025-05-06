package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// QueryGroupInfo return the group info
func (k Keeper) QueryGroupInfo(goCtx context.Context, req *types.QueryGroupInfoRequest) (*types.QueryGroupInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupInfo, exist := k.GetGroupInfo(ctx)
	if !exist {
		return nil, errors.New("data not found")
	}

	return &types.QueryGroupInfoResponse{
		GroupNumber:        groupInfo.GroupNumber,
		OperatorSetParam:   groupInfo.OperatorSetParam,
		MinimumStake:       groupInfo.MinimumStake.String(),
		PoolParams:         groupInfo.PoolParams,
		GroupEjectionParam: groupInfo.GroupEjectionParam,
	}, nil
}
