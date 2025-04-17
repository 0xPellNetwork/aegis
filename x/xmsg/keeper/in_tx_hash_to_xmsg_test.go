package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/nullify"
)

func TestInTxHashToXmsgGet(t *testing.T) {
	keeper, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNInTxHashToXmsg(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetInTxHashToXmsg(ctx,
			item.InTxHash,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestInTxHashToXmsgRemove(t *testing.T) {
	keeper, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNInTxHashToXmsg(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveInTxHashToXmsg(ctx,
			item.InTxHash,
		)
		_, found := keeper.GetInTxHashToXmsg(ctx,
			item.InTxHash,
		)
		require.False(t, found)
	}
}

func TestInTxHashToXmsgGetAll(t *testing.T) {
	keeper, ctx, _, _ := keepertest.XmsgKeeper(t)
	items := createNInTxHashToXmsg(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllInTxHashToXmsg(ctx)),
	)
}
