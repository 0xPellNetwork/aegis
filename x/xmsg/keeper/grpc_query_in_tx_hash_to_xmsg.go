package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func (k Keeper) InTxHashToXmsgAll(c context.Context, req *types.QueryAllInTxHashToXmsgRequest) (*types.QueryInTxHashToXmsgAllResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var inTxHashToXmsgs []types.InTxHashToXmsg
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	inTxHashToXmsgStore := prefix.NewStore(store, types.KeyPrefix(types.InTxHashToXmsgKeyPrefix))

	pageRes, err := query.Paginate(inTxHashToXmsgStore, req.Pagination, func(_ []byte, value []byte) error {
		var inTxHashToXmsg types.InTxHashToXmsg
		if err := k.cdc.Unmarshal(value, &inTxHashToXmsg); err != nil {
			return err
		}

		inTxHashToXmsgs = append(inTxHashToXmsgs, inTxHashToXmsg)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryInTxHashToXmsgAllResponse{InTxHashToXmsg: inTxHashToXmsgs, Pagination: pageRes}, nil
}

func (k Keeper) InTxHashToXmsg(c context.Context, req *types.QueryGetInTxHashToXmsgRequest) (*types.QueryInTxHashToXmsgResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetInTxHashToXmsg(
		ctx,
		req.InTxHash,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryInTxHashToXmsgResponse{InTxHashToXmsg: val}, nil
}

// InTxHashToXmsgData queries the data of all xmsgs indexed by a in tx hash
func (k Keeper) InTxHashToXmsgData(
	c context.Context,
	req *types.QueryInTxHashToXmsgDataRequest,
) (*types.QueryInTxHashToXmsgDataResponse, error) {
	inTxHashToXmsgRes, err := k.InTxHashToXmsg(c, &types.QueryGetInTxHashToXmsgRequest{InTxHash: req.InTxHash})
	if err != nil {
		return nil, err
	}

	xmsgs := make([]types.Xmsg, len(inTxHashToXmsgRes.InTxHashToXmsg.XmsgIndices))
	ctx := sdk.UnwrapSDKContext(c)
	for i, xmsgIndex := range inTxHashToXmsgRes.InTxHashToXmsg.XmsgIndices {
		xmsg, found := k.GetXmsg(ctx, xmsgIndex)
		if !found {
			// This is an internal error because the xmsg should always exist from the index
			return nil, status.Errorf(codes.Internal, "xmsg indexed %s doesn't exist", xmsgIndex)
		}

		xmsgs[i] = xmsg
	}

	return &types.QueryInTxHashToXmsgDataResponse{Xmsgs: xmsgs}, nil
}
