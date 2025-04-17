package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// UpdateRateLimiterFlags updates the rate limiter flags.
// Authorized: admin policy operational.
func (k msgServer) UpdateRateLimiterFlags(goCtx context.Context, msg *types.MsgUpdateRateLimiterFlags) (*types.MsgUpdateRateLimiterFlagsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, fmt.Sprintf("Creator %s", msg.Signer))
	}

	k.SetRateLimiterFlags(ctx, msg.RateLimiterFlags)

	return &types.MsgUpdateRateLimiterFlagsResponse{}, nil
}
