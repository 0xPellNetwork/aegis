package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

// RemoveChainParams removes chain parameters for a specific chain.
func (k msgServer) RemoveChainParams(goCtx context.Context, msg *types.MsgRemoveChainParams) (*types.MsgRemoveChainParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgRemoveChainParamsResponse{}, authoritytypes.ErrUnauthorized
	}

	// find current core params list or initialize a new one
	chainParamsList, found := k.GetChainParamsList(ctx)
	if !found {
		return &types.MsgRemoveChainParamsResponse{}, types.ErrChainParamsNotFound
	}

	tmpCtx, commit := ctx.CacheContext()

	// remove the core param from the list
	newChainParamsList := types.ChainParamsList{}
	found = false
	for _, cp := range chainParamsList.ChainParams {
		if cp.ChainId != msg.ChainId {
			newChainParamsList.ChainParams = append(newChainParamsList.ChainParams, cp)
		} else {
			found = true
		}
	}
	if !found {
		return &types.MsgRemoveChainParamsResponse{}, types.ErrChainParamsNotFound
	}

	k.SetChainParamsList(tmpCtx, newChainParamsList)

	commit()
	return &types.MsgRemoveChainParamsResponse{}, nil
}
