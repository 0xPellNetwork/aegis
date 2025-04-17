package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// GetEpochChangedOperatorSharesSnapshot returns the changed operator shares snapshot
func (k Keeper) GetEpochChangedOperatorSharesSnapshot(ctx context.Context, req *types.QueryEpochChangedOperatorSharesSnapshotRequest) (*types.QueryGetEpochChangedOperatorSharesSnapshotResponse, error) {
	snapshot, exist := k.GetChangedOperatorSharesSnapshot(sdk.UnwrapSDKContext(ctx), req.EpochNumber)
	if !exist {
		return nil, errors.New("snapshot not found")
	}

	operatorShares := make([]types.OperatorShares, len(snapshot.OperatorShares))
	for i, os := range snapshot.OperatorShares {
		operatorShares[i] = *os
	}

	return &types.QueryGetEpochChangedOperatorSharesSnapshotResponse{ChangedOperatorSharesSnapshot: operatorShares}, nil
}
