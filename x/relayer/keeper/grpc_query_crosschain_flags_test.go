package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestKeeper_CrosschainFlags(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.CrosschainFlags(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if crosschain flags not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.CrosschainFlags(wctx, &types.QueryGetCrosschainFlagsRequest{})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should return if crosschain flags found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		flags := types.CrosschainFlags{
			IsInboundEnabled: false,
		}
		k.SetCrosschainFlags(ctx, flags)

		res, err := k.CrosschainFlags(wctx, &types.QueryGetCrosschainFlagsRequest{})

		require.NoError(t, err)
		require.Equal(t, &types.QueryCrosschainFlagsResponse{
			CrosschainFlags: flags,
		}, res)
	})
}
