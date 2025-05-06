package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// AddAllowedXmsgSender adds a list of allowed xmsg sender to the xmsg module.
func (k Keeper) AddAllowedXmsgSender(goCtx context.Context, msg *types.MsgAddAllowedXmsgSender) (*types.MsgAddAllowedXmsgSenderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// admin only
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, "AddXmsgBuilders can only be executed by the correct policy account")
	}

	k.SaveAllowedXmsgSenders(ctx, msg.Builders)

	return &types.MsgAddAllowedXmsgSenderResponse{}, nil
}

// RemoveAllowedXmsgSender removes a list of allowed xmsg sender from the xmsg module.
func (k Keeper) RemoveAllowedXmsgSender(goCtx context.Context, msg *types.MsgRemoveAllowedXmsgSender) (*types.MsgRemoveAllowedXmsgSenderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// admin only
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, "AddXmsgBuilders can only be executed by the correct policy account")
	}

	k.DeleteAllowedXmsgSenders(ctx, msg.Builders)

	return &types.MsgRemoveAllowedXmsgSenderResponse{}, nil
}
