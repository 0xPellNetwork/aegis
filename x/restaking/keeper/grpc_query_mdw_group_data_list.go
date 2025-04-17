package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// QueryDVSGroupDataList returns the group data list
func (k Keeper) QueryDVSGroupDataList(ctx context.Context, req *types.QueryDVSGroupDataListRequest) (*types.QueryDVSGroupDataListResponse, error) {
	list, found := k.GetGroupDataList(sdk.UnwrapSDKContext(ctx), common.HexToAddress(req.RegistryRouterAddress))
	if !found {
		return nil, errors.New("supported group data found")
	}

	return &types.QueryDVSGroupDataListResponse{
		Groups: list.Groups,
	}, nil
}
