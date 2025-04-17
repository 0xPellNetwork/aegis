package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_GetRateLimiterFlags(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)

	// not found
	_, found := k.GetRateLimiterFlags(ctx)
	require.False(t, found)

	flags := sample.RateLimiterFlags_pell()

	k.SetRateLimiterFlags(ctx, flags)
	r, found := k.GetRateLimiterFlags(ctx)
	require.True(t, found)
	require.Equal(t, flags, r)
}

func TestKeeper_GetRateLimiterRates(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)

	flags := types.RateLimiterFlags{
		Rate: math.NewUint(100),
	}

	// set flags
	k.SetRateLimiterFlags(ctx, flags)
	r, found := k.GetRateLimiterFlags(ctx)
	require.True(t, found)
	require.Equal(t, flags, r)
}
