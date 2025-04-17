package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// RemoveFromOutTxTracker removes a record from the outbound transaction tracker by chain ID and nonce.
//
// Authorized: admin policy group 1.
func (k msgServer) RemoveFromOutTxTracker(goCtx context.Context, msg *types.MsgRemoveFromOutTxTracker) (*types.MsgRemoveFromOutTxTrackerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_EMERGENCY) {
		return &types.MsgRemoveFromOutTxTrackerResponse{}, authoritytypes.ErrUnauthorized
	}

	k.RemoveOutTxTracker(ctx, msg.ChainId, msg.Nonce)
	return &types.MsgRemoveFromOutTxTrackerResponse{}, nil
}
