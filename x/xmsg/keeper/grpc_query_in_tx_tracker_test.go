package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_InTxTrackerAllByChain(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	k.SetInTxTracker(ctx, types.InTxTracker{
		ChainId: 1,
		TxHash:  sample.Hash().Hex(),
	})
	k.SetInTxTracker(ctx, types.InTxTracker{
		ChainId: 2,
		TxHash:  sample.Hash().Hex(),
	})

	res, err := k.InTxTrackerAllByChain(ctx, &types.QueryAllInTxTrackerByChainRequest{
		ChainId: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(res.InTxTrackers))
}

func TestKeeper_InTxTrackerAll(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	k.SetInTxTracker(ctx, types.InTxTracker{
		ChainId: 1,
		TxHash:  sample.Hash().Hex(),
	})
	k.SetInTxTracker(ctx, types.InTxTracker{
		ChainId: 2,
		TxHash:  sample.Hash().Hex(),
	})

	res, err := k.InTxTrackerAll(ctx, &types.QueryAllInTxTrackersRequest{})
	require.NoError(t, err)
	require.Equal(t, 2, len(res.InTxTrackers))
}
