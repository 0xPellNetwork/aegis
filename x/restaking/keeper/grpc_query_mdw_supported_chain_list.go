package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// QueryDVSSupportedChainList returns the supported chain list
func (k Keeper) QueryDVSSupportedChainList(ctx context.Context, req *types.QueryDVSSupportedChainListRequest) (*types.QueryDVSSupportedChainListResponse, error) {
	list, found := k.GetDVSSupportedChainList(sdk.UnwrapSDKContext(ctx), common.HexToAddress(req.RegistryRouterAddress))
	if !found {
		return nil, errors.New("supported chain not found")
	}

	return &types.QueryDVSSupportedChainListResponse{
		DvsInfos: list,
	}, nil
}
