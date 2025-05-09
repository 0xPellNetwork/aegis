package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// UpdateKeygen updates the block height of the keygen and sets the status to
// "pending keygen".
//
// Authorized: admin policy group 1.
func (k msgServer) UpdateKeygen(goCtx context.Context, msg *types.MsgUpdateKeygen) (*types.MsgUpdateKeygenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_EMERGENCY) {
		return &types.MsgUpdateKeygenResponse{}, authoritytypes.ErrUnauthorized
	}

	keygen, found := k.GetKeygen(ctx)
	if !found {
		return nil, types.ErrKeygenNotFound
	}
	if msg.Block <= (ctx.BlockHeight() + 10) {
		return nil, types.ErrKeygenBlockTooLow
	}

	nodeAccountList := k.GetAllNodeAccount(ctx)
	granteePubKeys := make([]string, len(nodeAccountList))
	for i, nodeAccount := range nodeAccountList {
		granteePubKeys[i] = nodeAccount.GranteePubkey.Secp256k1.String()
	}

	// update keygen
	keygen.GranteePubkeys = granteePubKeys
	keygen.BlockNumber = msg.Block
	keygen.Status = types.KeygenStatus_PENDING
	k.SetKeygen(ctx, keygen)

	EmitEventKeyGenBlockUpdated(ctx, &keygen)

	return &types.MsgUpdateKeygenResponse{}, nil
}
