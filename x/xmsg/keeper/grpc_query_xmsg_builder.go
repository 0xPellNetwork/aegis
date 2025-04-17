package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// ListXmsgBuilders returns a list of authorized Xmsg builder addresses.
// It retrieves the list of builders from the keeper and returns them in the response.
// This function is used to query the current set of authorized Xmsg builders in the system.
func (k Keeper) ListAllowedXmsgSenders(ctx context.Context, req *types.QueryListAllowedXmsgSendersRequest) (*types.QueryListAllowedXmsgSendersResponse, error) {
	builders, _ := k.GetAllowedXmsgSenders(sdk.UnwrapSDKContext(ctx))
	return &types.QueryListAllowedXmsgSendersResponse{Builders: builders}, nil
}
