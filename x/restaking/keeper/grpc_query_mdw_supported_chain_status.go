package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// QueryDVSSupportedChainStatus returns the supported chain state by chain id
func (k Keeper) QueryDVSSupportedChainStatus(ctx context.Context, req *types.QueryDVSSupportedChainStatusRequest) (*types.QueryDVSSupportedChainStatusResponse, error) {
	status, err := k.GetDVSSupportedChainStatus(sdk.UnwrapSDKContext(ctx), common.HexToAddress(req.RegistryRouterAddress), req.ChainId)
	if err != nil {
		return nil, errors.New("supported chain state not found")
	}

	return &types.QueryDVSSupportedChainStatusResponse{OutboundState: status}, nil
}
