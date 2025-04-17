package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestKeeper_GetChainParamsForChain(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetChainParamsForChain(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if chain params not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetChainParamsForChain(wctx, &types.QueryGetChainParamsForChainRequest{
			ChainId: 987,
		})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should return if chain params found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		list := types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     chains.PellPrivnetChain().Id,
					IsSupported: false,
				},
			},
		}
		k.SetChainParamsList(ctx, list)

		res, err := k.GetChainParamsForChain(wctx, &types.QueryGetChainParamsForChainRequest{
			ChainId: chains.PellPrivnetChain().Id,
		})
		require.NoError(t, err)
		require.EqualValues(t, list.ChainParams[0].String(), res.ChainParams.String())
	})
}

func TestKeeper_GetChainParams(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetChainParams(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if chain params not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetChainParams(wctx, &types.QueryGetChainParamsRequest{})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should return if chain params found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		list := types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     chains.PellPrivnetChain().Id,
					IsSupported: false,
				},
			},
		}
		k.SetChainParamsList(ctx, list)

		res, err := k.GetChainParams(wctx, &types.QueryGetChainParamsRequest{})
		require.NoError(t, err)
		require.EqualValues(t, list.String(), res.ChainParams.String())
	})
}
