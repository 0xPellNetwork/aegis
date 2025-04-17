package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// AddPools adds pools to the group
func (k Keeper) AddPools(goctx context.Context, msg *types.MsgAddPools) (*types.MsgAddPoolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgAddPoolsResponse{}, authoritytypes.ErrUnauthorized
	}

	groupInfo, exist := k.GetGroupInfo(ctx)
	if !exist {
		return nil, types.ErrDataEmpty
	}

	address, exist := k.GetLSTRegistryRouterAddress(ctx)
	if !exist {
		return nil, types.ErrDataEmpty
	}

	// call the pevm keeper to add pools
	if _, _, err := k.pevmKeeper.CallStakelRegistryRouterToAddPools(ctx, common.HexToAddress(address.StakeRegistryRouterAddress), groupInfo.GroupNumber, msg.Pools); err != nil {
		return nil, err
	}

	// Update the group info with the new pools
	groupInfo.PoolParams = append(groupInfo.PoolParams, msg.Pools...)
	k.SetGroupInfo(ctx, groupInfo)

	return &types.MsgAddPoolsResponse{}, nil
}

// RemovePools removes pools from the group
func (k Keeper) RemovePools(goctx context.Context, msg *types.MsgRemovePools) (*types.MsgRemovePoolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgRemovePoolsResponse{}, authoritytypes.ErrUnauthorized
	}

	groupInfo, exist := k.GetGroupInfo(ctx)
	if !exist {
		return nil, types.ErrDataEmpty
	}

	address, exist := k.GetLSTRegistryRouterAddress(ctx)
	if !exist {
		return nil, types.ErrDataEmpty
	}

	var indicesToRemove []uint
	for _, pool := range msg.Pools {
		for i, poolInfo := range groupInfo.PoolParams {
			if poolInfo.Pool == pool.Pool {
				indicesToRemove = append(indicesToRemove, uint(i))
			}
		}
	}

	// call the pevm keeper to remove pools
	if _, _, err := k.pevmKeeper.CallStakelRegistryRouterToRemovePools(ctx, common.HexToAddress(address.StakeRegistryRouterAddress), groupInfo.GroupNumber, indicesToRemove); err != nil {
		return nil, err
	}

	// Update the group groupInfo by removing the specified pools
	for _, index := range indicesToRemove {
		if index < uint(len(groupInfo.PoolParams)) {
			groupInfo.PoolParams = append(groupInfo.PoolParams[:index], groupInfo.PoolParams[index+1:]...)
		}
	}

	k.SetGroupInfo(ctx, groupInfo)

	return &types.MsgRemovePoolsResponse{}, nil
}
