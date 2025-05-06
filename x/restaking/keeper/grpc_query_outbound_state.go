package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// GetOutboundStateByChainID returns the outbound state by chain id
func (k Keeper) GetOutboundStateByChainID(ctx context.Context, req *types.QueryOutboundStateByChainIDRequest) (*types.QueryGetOutboundStateByChainIDResponse, error) {
	outboundState, exist := k.GetOutboundState(sdk.UnwrapSDKContext(ctx), req.ChainId)
	if !exist {
		return nil, errors.New("outbound state not found")
	}

	return &types.QueryGetOutboundStateByChainIDResponse{OutboundState: outboundState}, nil
}
