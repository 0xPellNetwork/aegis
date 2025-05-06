package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
)

// Keeper Tests

func TestGasPriceGet(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNGasPrice(k, ctx, 10)
	for _, item := range items {
		rst, found := k.GetGasPrice(ctx, item.ChainId)
		require.True(t, found)
		require.Equal(t, item, rst)
	}
}

func TestGasPriceRemove(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNGasPrice(k, ctx, 10)
	for _, item := range items {
		k.RemoveGasPrice(ctx, item.Index)
		_, found := k.GetGasPrice(ctx, item.ChainId)
		require.False(t, found)
	}
}

func TestGasPriceGetAll(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNGasPrice(k, ctx, 10)
	require.Equal(t, items, k.GetAllGasPrice(ctx))
}
