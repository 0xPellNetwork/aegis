package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	// MaxPendingXmsgs is the maximum number of pending xmsgs that can be queried
	MaxPendingXmsgs = 700

	// MaxLookbackNonce is the maximum number of nonces to look back to find missed pending xmsgs
	MaxLookbackNonce = 1000
)

func (k Keeper) XmsgAll(c context.Context, req *types.QueryAllXmsgRequest) (*types.QueryXmsgAllResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	var sends []*types.Xmsg
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	sendStore := prefix.NewStore(store, types.KeyPrefix(types.SendKey))

	pageRes, err := query.Paginate(sendStore, req.Pagination, func(_ []byte, value []byte) error {
		var send types.Xmsg
		if err := k.cdc.Unmarshal(value, &send); err != nil {
			return err
		}
		sends = append(sends, &send)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryXmsgAllResponse{Xmsgs: sends, Pagination: pageRes}, nil
}

func (k Keeper) Xmsg(c context.Context, req *types.QueryGetXmsgRequest) (*types.QueryXmsgResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetXmsg(ctx, req.Index)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}

	return &types.QueryXmsgResponse{Xmsg: &val}, nil
}

func (k Keeper) XmsgByNonce(c context.Context, req *types.QueryGetXmsgByNonceRequest) (*types.QueryXmsgByNonceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	tss, found := k.GetRelayerKeeper().GetTSS(ctx)
	if !found {
		return nil, status.Error(codes.Internal, "tss not found")
	}
	// #nosec G701 always in range
	xmsg, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, req.ChainId, int64(req.Nonce))
	if err != nil {
		return nil, err
	}

	return &types.QueryXmsgByNonceResponse{Xmsg: xmsg}, nil
}

// ListPendingXmsg returns a list of pending xmsgs and the total number of pending xmsgs
// a limit for the number of xmsgs to return can be specified or the default is MaxPendingXmsgs
func (k Keeper) ListPendingXmsg(c context.Context, req *types.QueryListPendingXmsgRequest) (*types.QueryListPendingXmsgResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// use default MaxPendingXmsgs if not specified or too high
	limit := req.Limit
	if limit == 0 || limit > MaxPendingXmsgs {
		limit = MaxPendingXmsgs
	}
	ctx := sdk.UnwrapSDKContext(c)

	// query the nonces that are pending
	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, observertypes.ErrTssNotFound
	}
	pendingNonces, found := k.GetRelayerKeeper().GetPendingNonces(ctx, tss.TssPubkey, req.ChainId)
	if !found {
		return nil, status.Error(codes.Internal, "pending nonces not found")
	}

	xmsgs := make([]*types.Xmsg, 0)
	maxXmsgsReached := func() bool {
		// #nosec G701 len always positive
		return uint32(len(xmsgs)) >= limit
	}

	totalPending := uint64(0)

	// now query the previous nonces up to 1000 prior to find any pending xmsg that we might have missed
	// need this logic because a confirmation of higher nonce will automatically update the p.NonceLow
	// therefore might mask some lower nonce xmsg that is still pending.
	startNonce := pendingNonces.NonceLow - MaxLookbackNonce
	if startNonce < 0 {
		startNonce = 0
	}
	for i := startNonce; i < pendingNonces.NonceLow; i++ {
		xmsg, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, req.ChainId, i)
		if err != nil {
			return nil, err
		}

		// only take a `limit` number of pending xmsgs as result but still count the total pending xmsgs
		if IsPending(xmsg) {
			totalPending++
			if !maxXmsgsReached() {
				xmsgs = append(xmsgs, xmsg)
			}
		}
	}

	// add the pending nonces to the total pending
	// #nosec G701 always in range
	totalPending += uint64(pendingNonces.NonceHigh - pendingNonces.NonceLow)

	// now query the pending nonces that we know are pending
	for i := pendingNonces.NonceLow; i < pendingNonces.NonceHigh && !maxXmsgsReached(); i++ {
		xmsg, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, req.ChainId, i)
		if err != nil {
			return nil, err
		}
		xmsgs = append(xmsgs, xmsg)
	}

	return &types.QueryListPendingXmsgResponse{
		Xmsg:         xmsgs,
		TotalPending: totalPending,
	}, nil
}

// getXmsgByChainIDAndNonce returns the xmsg by chainID and nonce
func getXmsgByChainIDAndNonce(k Keeper, ctx sdk.Context, tssPubkey string, chainID int64, nonce int64) (*types.Xmsg, error) {
	nonceToXmsg, found := k.GetRelayerKeeper().GetNonceToXmsg(ctx, tssPubkey, chainID, nonce)
	if !found {
		return nil, status.Error(codes.Internal, fmt.Sprintf("nonceToXmsg not found: chainid %d, nonce %d", chainID, nonce))
	}
	xmsg, found := k.GetXmsg(ctx, nonceToXmsg.XmsgIndex)
	if !found {
		return nil, status.Error(codes.Internal, fmt.Sprintf("xmsg not found: index %s", nonceToXmsg.XmsgIndex))
	}
	return &xmsg, nil
}
