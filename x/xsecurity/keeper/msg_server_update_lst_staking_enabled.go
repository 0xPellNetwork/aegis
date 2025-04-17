package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// UpdateLSTStakingEnabled updates the LST staking enabled status
func (k Keeper) UpdateLSTStakingEnabled(goctx context.Context, msg *types.MsgUpdateLSTStakingEnabled) (*types.MsgUpdateLSTStakingEnabledResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgUpdateLSTStakingEnabledResponse{}, authoritytypes.ErrUnauthorized
	}

	k.SetLSTStakingEnabled(ctx, msg.Enabled)

	return &types.MsgUpdateLSTStakingEnabledResponse{}, nil
}
