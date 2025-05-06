package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_RateLimiterFlags(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.RateLimiterFlags(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if rate limiter flags not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.RateLimiterFlags(wctx, &types.QueryRateLimiterFlagsRequest{})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should return if rate limiter flags found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		flags := sample.RateLimiterFlags_pell()
		k.SetRateLimiterFlags(ctx, flags)

		res, err := k.RateLimiterFlags(wctx, &types.QueryRateLimiterFlagsRequest{})

		require.NoError(t, err)
		require.Equal(t, &types.QueryRateLimiterFlagsResponse{
			RateLimiterFlags: flags,
		}, res)
	})
}
