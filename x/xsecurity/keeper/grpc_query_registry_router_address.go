package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// QueryRegistryRouterAddress returns the registry router address and stake registry router address
func (k Keeper) QueryRegistryRouterAddress(goCtx context.Context, req *types.QueryRegistryRouterAddressRequest) (*types.QueryRegistryRouterAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, exist := k.GetLSTRegistryRouterAddress(ctx)
	if !exist {
		return nil, errors.New("data not found")
	}

	return &types.QueryRegistryRouterAddressResponse{
		RegistryRouterAddress:      address.RegistryRouterAddress,
		StakeRegistryRouterAddress: address.StakeRegistryRouterAddress,
	}, nil
}
