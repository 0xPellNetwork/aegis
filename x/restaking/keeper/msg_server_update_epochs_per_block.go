package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
)

func (k Keeper) UpdateBlocksPerEpoch(goCtx context.Context, msg *restakingtypes.MsgUpdateBlocksPerEpoch) (*restakingtypes.MsgUpdateBlocksPerEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, "Update can only be executed by the correct policy account")
	}

	k.SetBlocksPerEpoch(ctx, msg.BlocksPerEpoch)

	return &restakingtypes.MsgUpdateBlocksPerEpochResponse{}, nil
}
