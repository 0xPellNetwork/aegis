package keeper

import (
	"context"

	"github.com/cockroachdb/errors/grpc/status"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func (k Keeper) BlockProof(c context.Context, req *types.QueryBlockProofRequest) (*types.QueryBlockProofResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	blockProof, exist := k.GetBlockProof(ctx, uint64(req.ChainId), req.Height)
	if exist {
		return (*types.QueryBlockProofResponse)(&blockProof), nil
	}

	return nil, status.Error(codes.InvalidArgument, "block proof not exist")
}
