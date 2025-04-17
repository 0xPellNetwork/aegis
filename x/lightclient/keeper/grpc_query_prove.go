package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

// Prove checks two things:
// 1. the block header is available
// 2. the proof is valid
func (k Keeper) Prove(c context.Context, req *types.QueryProveRequest) (*types.QueryProveResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	blockHash, err := chains.StringToHash(req.ChainId, req.BlockHash)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	res, found := k.GetBlockHeader(ctx, blockHash)
	if !found {
		return nil, status.Error(codes.NotFound, "block header not found")
	}

	proven := false

	txBytes, err := req.Proof.Verify(res.Header, int(req.TxIndex))
	if err != nil && !proofs.IsErrorInvalidProof(err) {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err == nil {
		if chains.IsEVMChain(req.ChainId) {
			var txx ethtypes.Transaction
			err = txx.UnmarshalBinary(txBytes)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed to unmarshal evm transaction: %s", err))
			}
			if txx.Hash().Hex() != req.TxHash {
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("tx hash mismatch: %s != %s", txx.Hash().Hex(), req.TxHash))
			}
			proven = true
		} else {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid chain id (%d)", req.ChainId))
		}
	}

	return &types.QueryProveResponse{
		Valid: proven,
	}, nil
}
