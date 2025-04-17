package keeper

import (
	"context"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// QueryDVSGroupSyncStatus returns the group data sync status
func (k Keeper) QueryDVSGroupSyncStatus(ctx context.Context, req *types.QueryDVSGroupSyncStatusRequest) (*types.QueryDVSGroupSyncStatusResponse, error) {
	list, found := k.GetGroupSyncList(sdk.UnwrapSDKContext(ctx), req.TxHash)
	if !found {
		return nil, cosmoserrors.Wrapf(types.ErrInvalidData, "quorum sync not found")
	}

	ret, err := k.xmsgKeeper.XmsgAll(ctx, &xmsgtypes.QueryAllXmsgRequest{})
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrInvalidData, "failed to query xmsg: %s", err)
	}

	xmsgs := make([]*xmsgtypes.Xmsg, 0)
	for _, xmsg := range ret.Xmsgs {
		for _, xmsgIndex := range list.XmsgIndex {
			if xmsg.Index == xmsgIndex {
				xmsgs = append(xmsgs, xmsg)
			}
		}
	}

	return &types.QueryDVSGroupSyncStatusResponse{Xmsg: xmsgs}, nil
}
