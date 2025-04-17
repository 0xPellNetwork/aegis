package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
)

func TestLastBlockHeightGet(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNLastBlockHeight(k, ctx, 10)
	for _, item := range items {
		rst, found := k.GetLastBlockHeight(ctx, item.Index)
		require.True(t, found)
		require.Equal(t, item, rst)
	}
}

func TestLastBlockHeightRemove(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNLastBlockHeight(k, ctx, 10)
	for _, item := range items {
		k.RemoveLastBlockHeight(ctx, item.Index)
		_, found := k.GetLastBlockHeight(ctx, item.Index)
		require.False(t, found)
	}
}

func TestLastBlockHeightGetAll(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNLastBlockHeight(k, ctx, 10)
	require.Equal(t, items, k.GetAllLastBlockHeight(ctx))
}
