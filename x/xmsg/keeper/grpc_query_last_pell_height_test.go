package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_LastPellHeight(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		res, err := k.LastPellHeight(ctx, nil)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if height less than zero", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		ctx = ctx.WithBlockHeight(-1)
		res, err := k.LastPellHeight(ctx, &types.QueryLastPellHeightRequest{})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should return height if gte 0", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		ctx = ctx.WithBlockHeight(0)
		res, err := k.LastPellHeight(ctx, &types.QueryLastPellHeightRequest{})
		require.NoError(t, err)
		require.Equal(t, int64(0), res.Height)

		ctx = ctx.WithBlockHeight(5)
		res, err = k.LastPellHeight(ctx, &types.QueryLastPellHeightRequest{})
		require.NoError(t, err)
		require.Equal(t, int64(5), res.Height)
	})
}
