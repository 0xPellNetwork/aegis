package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// UpsertCrosschainFeeParams upserts crosschain fee params
func (k Keeper) UpsertCrosschainFeeParams(goctx context.Context, msg *types.MsgUpsertCrosschainFeeParams) (*types.MsgUpsertCrosschainFeeParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgUpsertCrosschainFeeParamsResponse{}, authoritytypes.ErrUnauthorized
	}

	for _, crosschainFeeParam := range msg.CrosschainFeeParams {
		k.StoreCrosschainEventFee(ctx, *crosschainFeeParam)
	}

	return &types.MsgUpsertCrosschainFeeParamsResponse{}, nil
}
