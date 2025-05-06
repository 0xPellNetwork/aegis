package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// SetGroupParam sets the group parameters
func (k Keeper) SetGroupParam(goctx context.Context, msg *types.MsgSetGroupParam) (*types.MsgSetGroupParamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgSetGroupParamResponse{}, authoritytypes.ErrUnauthorized
	}

	groupInfo, exist := k.GetGroupInfo(ctx)
	if !exist {
		return nil, types.ErrDataEmpty
	}

	address, exist := k.GetLSTRegistryRouterAddress(ctx)
	if !exist {
		return nil, types.ErrDataEmpty
	}

	// call the pevm keeper to set group parameters
	if _, _, err := k.pevmKeeper.CallRegistryRouterToSetOperatorSetParams(ctx, common.HexToAddress(address.RegistryRouterAddress), groupInfo.GroupNumber, msg.OperatorSetParams); err != nil {
		return nil, err
	}

	// Update the group info with the new parameters
	groupInfo.OperatorSetParam = msg.OperatorSetParams

	k.SetGroupInfo(ctx, groupInfo)

	return &types.MsgSetGroupParamResponse{}, nil
}
