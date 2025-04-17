package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// UpdateVotingPowerRatio updates the voting power ratio for LST tokens
func (k Keeper) UpdateVotingPowerRatio(goctx context.Context, msg *types.MsgUpdateVotingPowerRatio) (*types.MsgUpdateVotingPowerRatioResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	if msg.Denominator.IsZero() {
		return &types.MsgUpdateVotingPowerRatioResponse{}, types.ErrInvalidDenominator
	}

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgUpdateVotingPowerRatioResponse{}, authoritytypes.ErrUnauthorized
	}

	k.SetLSTVotingPowerRatio(ctx, msg.Numerator, msg.Denominator)

	return &types.MsgUpdateVotingPowerRatioResponse{}, nil
}
