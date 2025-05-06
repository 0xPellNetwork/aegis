package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	// AbortMessage is the message to abort a stuck Xmsg
	AbortMessage = "Xmsg aborted with admin cmd"
)

// AbortStuckXmsg aborts a stuck Xmsg
// Authorized: admin policy group 2
func (k msgServer) AbortStuckXmsg(
	goCtx context.Context,
	msg *types.MsgAbortStuckXmsg,
) (*types.MsgAbortStuckXmsgResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if authorized
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, authoritytypes.ErrUnauthorized
	}

	// check if the xmsg exists
	xmsg, found := k.GetXmsg(ctx, msg.XmsgIndex)
	if !found {
		return nil, types.ErrCannotFindXmsg
	}

	// check if the xmsg is pending
	isPending := xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_OUTBOUND ||
		xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_INBOUND ||
		xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_REVERT
	if !isPending {
		return nil, types.ErrStatusNotPending
	}

	xmsg.XmsgStatus = &types.Status{
		Status:        types.XmsgStatus_ABORTED,
		StatusMessage: AbortMessage,
	}

	k.SetXmsg(ctx, xmsg)

	return &types.MsgAbortStuckXmsgResponse{}, nil
}
