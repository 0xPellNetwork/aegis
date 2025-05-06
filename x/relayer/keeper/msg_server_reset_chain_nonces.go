package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// ResetChainNonces handles resetting chain nonces
func (k msgServer) ResetChainNonces(goCtx context.Context, msg *types.MsgResetChainNonces) (*types.MsgResetChainNoncesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgResetChainNoncesResponse{}, authoritytypes.ErrUnauthorized
	}

	tss, found := k.GetTSS(ctx)
	if !found {
		return nil, types.ErrTssNotFound
	}

	chain, exist := chains.GetChainByChainId(msg.ChainId)
	if !exist {
		return nil, types.ErrSupportedChains
	}

	// set chain nonces
	chainNonce := types.ChainNonces{
		Index:   chain.ChainName(),
		ChainId: chain.Id,
		// #nosec G701 always positive
		Nonce: uint64(msg.ChainNonceHigh),
		// #nosec G701 always positive
		FinalizedHeight: uint64(ctx.BlockHeight()),
	}
	k.SetChainNonces(ctx, chainNonce)

	// set pending nonces
	p := types.PendingNonces{
		NonceLow:  msg.ChainNonceLow,
		NonceHigh: msg.ChainNonceHigh,
		ChainId:   chain.Id,
		Tss:       tss.TssPubkey,
	}
	k.SetPendingNonces(ctx, p)

	return &types.MsgResetChainNoncesResponse{}, nil
}
