package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

// UpdateVerificationFlags updates the light client verification flags.
// This disables/enables blocks verification of the light client for the specified chain.
// Emergency group can disable flags, it requires operational group if at least one flag is being enabled
func (k msgServer) UpdateVerificationFlags(goCtx context.Context, msg *types.MsgUpdateVerificationFlags) (
	*types.MsgUpdateVerificationFlagsResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, msg.GetRequireGroup()) {
		return &types.MsgUpdateVerificationFlagsResponse{}, authoritytypes.ErrUnauthorized
	}

	k.SetVerificationFlags(ctx, msg.VerificationFlags)

	return &types.MsgUpdateVerificationFlagsResponse{}, nil
}
