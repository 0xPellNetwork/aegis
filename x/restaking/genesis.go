package restaking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/x/restaking/keeper"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// InitGenesis initializes the pevm module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// import all operator shares
	for _, share := range genState.OperatorShare {
		k.SetOperatorShares(ctx, share.ChainId, share.Operator, share.Strategy, share.Shares)
	}

	// import all registry router addresses
	for _, registryRouter := range genState.RegistryRouterData {
		registryRouterAddr := common.HexToAddress(registryRouter.RegistryRouterSet.RegistryRouterAddress)

		if err := k.AddRegistryRouterAddress(ctx, []common.Address{
			registryRouterAddr,
			common.HexToAddress(registryRouter.RegistryRouterSet.StakeRegistryRouterAddress),
		}); err != nil {
			continue
		}

		// import all supported chain info by registry router
		for _, dvs := range registryRouter.DvsInfoList.DvsInfos {
			if err := k.AddDVSSupportedChain(ctx, registryRouterAddr, dvs); err != nil {
				continue
			}
		}

		// import all group data by registry router
		for _, group := range registryRouter.GroupList.Groups {
			if err := k.AddGroupData(ctx, registryRouterAddr, group); err != nil {
				continue
			}
		}

		// import all group operator registration by registry router
		for _, registration := range registryRouter.GroupOperatorRegistrationList.OperatorRegisteredInfos {
			if err := k.AddGroupOperatorRegistration(ctx, registryRouterAddr, registration); err != nil {
				continue
			}
		}
	}
}

// ExportGenesis returns the pevm module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// export all operator shares
	shares := k.GetAllShares(ctx)
	operatorShares := make([]types.OperatorShares, len(shares))
	for i, share := range shares {
		operatorShares[i] = *share
	}

	// export all registry router addresses
	registryRouterSetResponse, err := k.QueryDVSRegistryRouterList(ctx, nil)
	if err != nil {
		return nil
	}

	registryRouterDataList := make([]types.RegistryRouterData, 0)

	for _, registryRouter := range registryRouterSetResponse.RegistryRouterSet {
		registryRouterAddr := common.HexToAddress(registryRouter.RegistryRouterAddress)

		// export all supported chain info by registry router
		dvsInfoList, exist := k.GetDVSSupportedChainList(ctx, registryRouterAddr)
		if !exist {
			continue
		}

		// export all group data by registry router
		groupDataList, exist := k.GetGroupDataList(ctx, registryRouterAddr)
		if !exist {
			continue
		}

		// export all group operator registration by registry router
		registrationList, exist := k.GetGroupOperatorRegistrationList(ctx, registryRouterAddr)
		if !exist {
			continue
		}

		registryRouterDataList = append(registryRouterDataList, types.RegistryRouterData{
			RegistryRouterSet:             *registryRouter,
			DvsInfoList:                   *dvsInfoList,
			GroupList:                     *groupDataList,
			GroupOperatorRegistrationList: *registrationList,
		})
	}

	return &types.GenesisState{
		OperatorShare:      operatorShares,
		RegistryRouterData: registryRouterDataList,
	}
}
