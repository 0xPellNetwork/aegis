package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/emissions/types"
)

func TestKeeper_ShowAvailableEmissions(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionsKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.ShowAvailableEmissions(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should return 0 if emissions not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionsKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		req := &types.QueryShowAvailableEmissionsRequest{
			Address: sample.AccAddress(),
		}
		res, err := k.ShowAvailableEmissions(wctx, req)
		require.NoError(t, err)
		expectedRes := &types.QueryShowAvailableEmissionsResponse{
			Amount: sdk.NewCoin(config.BaseDenom, math.ZeroInt()).String(),
		}
		require.Equal(t, expectedRes, res)
	})

	t.Run("should return emissions if found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionsKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		emissions := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, emissions)

		req := &types.QueryShowAvailableEmissionsRequest{
			Address: emissions.Address,
		}
		res, err := k.ShowAvailableEmissions(wctx, req)
		require.NoError(t, err)
		expectedRes := &types.QueryShowAvailableEmissionsResponse{
			Amount: sdk.NewCoin(config.BaseDenom, emissions.Amount).String(),
		}
		require.Equal(t, expectedRes, res)
	})
}
