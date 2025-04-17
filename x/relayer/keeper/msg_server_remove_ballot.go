package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

// TODO: remove this after the next upgrade
func (k msgServer) DeleteBallot(ctx context.Context, msg *types.MsgDeleteBallot) (*types.MsgDeleteBallotResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(sdkCtx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return &types.MsgDeleteBallotResponse{}, authoritytypes.ErrUnauthorized
	}

	k.Keeper.DeleteBallot(sdkCtx, msg.BallotIndex)

	return &types.MsgDeleteBallotResponse{}, nil
}
