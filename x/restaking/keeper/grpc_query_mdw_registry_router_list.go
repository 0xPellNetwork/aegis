package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// QueryDVSRegistryRouterList queries all registry router addresses and their stake addresses.
func (k Keeper) QueryDVSRegistryRouterList(ctx context.Context, req *types.QueryDVSRegistryRouterListRequest) (*types.QueryDVSRegistryRouterListResponse, error) {
	// Unwrap the SDK context
	sdkContext := sdk.UnwrapSDKContext(ctx)

	// Retrieve all registry router addresses
	registryRouterAddresses, err := k.GetAllRegistryRouterAddresses(sdkContext)
	if err != nil {
		return nil, errors.New("registry router address not found")
	}

	// Preallocate the slice to optimize memory usage
	list := make([]*types.RegistryRouterSet, 0, len(registryRouterAddresses))

	// Build the response slice
	for _, address := range registryRouterAddresses {

		// Attempt to get stake registry router address
		registryRouterAddr, err := k.GetStakeRegistryRouterAddress(sdkContext, address)
		if err != nil {
			continue
		}

		// Construct the router set struct
		routerSet := &types.RegistryRouterSet{
			RegistryRouterAddress:      registryRouterAddr.String(),
			StakeRegistryRouterAddress: address.String(),
		}

		list = append(list, routerSet)
	}

	// Return the final response
	return &types.QueryDVSRegistryRouterListResponse{
		RegistryRouterSet: list,
	}, nil
}
