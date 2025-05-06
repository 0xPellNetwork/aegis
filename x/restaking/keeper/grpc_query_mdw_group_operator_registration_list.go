package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// QueryDVSGroupOperatorRegistrationList returns the group operator registration list
func (k Keeper) QueryDVSGroupOperatorRegistrationList(ctx context.Context, req *types.QueryDVSGroupOperatorRegistrationListRequest) (*types.QueryDVSGroupOperatorRegistrationListResponse, error) {
	list, found := k.GetGroupOperatorRegistrationList(sdk.UnwrapSDKContext(ctx), common.HexToAddress(req.RegistryRouterAddress))
	if !found {
		return nil, errors.New("supported group data found")
	}

	return &types.QueryDVSGroupOperatorRegistrationListResponse{
		OperatorRegisteredInfos: list.OperatorRegisteredInfos,
	}, nil
}
