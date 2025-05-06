package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// Chain nonces queries

func (k Keeper) ChainNoncesAll(c context.Context, req *types.QueryAllChainNoncesRequest) (*types.QueryChainNoncesAllResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var chainNoncess []types.ChainNonces
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	chainNoncesStore := prefix.NewStore(store, types.KeyPrefix(types.ChainNoncesKey))

	pageRes, err := query.Paginate(chainNoncesStore, req.Pagination, func(_ []byte, value []byte) error {
		var chainNonces types.ChainNonces
		if err := k.cdc.Unmarshal(value, &chainNonces); err != nil {
			return err
		}

		chainNoncess = append(chainNoncess, chainNonces)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryChainNoncesAllResponse{ChainNonces: chainNoncess, Pagination: pageRes}, nil
}

func (k Keeper) ChainNonces(c context.Context, req *types.QueryGetChainNoncesRequest) (*types.QueryChainNoncesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetChainNonces(ctx, req.Index)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}

	return &types.QueryChainNoncesResponse{ChainNonces: val}, nil
}

// Pending nonces queries

func (k Keeper) PendingNoncesAll(c context.Context, req *types.QueryAllPendingNoncesRequest) (*types.QueryPendingNoncesAllResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	list, pageRes, err := k.GetAllPendingNoncesPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryPendingNoncesAllResponse{
		PendingNonces: list,
		Pagination:    pageRes,
	}, nil
}

func (k Keeper) PendingNoncesByChain(c context.Context, req *types.QueryPendingNoncesByChainRequest) (*types.QueryPendingNoncesByChainResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	tss, found := k.GetTSS(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "tss not found")
	}
	list, found := k.GetPendingNonces(ctx, tss.TssPubkey, req.ChainId)
	if !found {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("pending nonces not found for chain id : %d", req.ChainId))
	}

	return &types.QueryPendingNoncesByChainResponse{
		PendingNonces: list,
	}, nil
}
