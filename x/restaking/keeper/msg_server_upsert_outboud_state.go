package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
)

// UpsertOutboundState upserts the outbound state
func (k Keeper) UpsertOutboundState(goctx context.Context, msg *restakingtypes.MsgUpsertOutboundState) (*restakingtypes.MsgUpsertOutboundStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, "Update can only be executed by the correct policy account")
	}

	if _, found := k.relayerKeeper.GetChainParamsByChainID(ctx, int64(msg.OutboundState.ChainId)); !found {
		return nil, errorsmod.Wrap(relayertypes.ErrChainParamsNotFound, "chain id not found in chain params")
	}

	if err := k.SetOutboundState(ctx, msg.OutboundState); err != nil {
		return nil, err
	}

	return &restakingtypes.MsgUpsertOutboundStateResponse{}, nil
}
