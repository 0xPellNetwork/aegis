package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestKeeper_ShowObserverCount(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.ShowObserverCount(wctx, nil)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.ShowObserverCount(wctx, &types.QueryShowObserverCountRequest{})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should return if found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		count := 1
		loc := &types.LastRelayerCount{
			Count: uint64(count),
		}
		k.SetLastObserverCount(ctx, loc)

		res, err := k.ShowObserverCount(wctx, &types.QueryShowObserverCountRequest{})
		require.NoError(t, err)
		require.Equal(t, loc, res.LastObserverCount)
	})
}

func TestKeeper_ObserverSet(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.ObserverSet(wctx, nil)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.ObserverSet(wctx, &types.QueryObserverSet{})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should return if found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		os := sample.ObserverSet_pell(10)
		k.SetObserverSet(ctx, os)

		res, err := k.ObserverSet(wctx, &types.QueryObserverSet{})
		require.NoError(t, err)
		require.Equal(t, os.RelayerList, res.Observers)
	})
}
