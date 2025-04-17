package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

func TestKeeper_SystemContract(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		res, err := k.SystemContract(ctx, nil)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if system contract not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		res, err := k.SystemContract(ctx, &types.QueryGetSystemContractRequest{})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should return system contract if found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		sc := types.SystemContract{
			SystemContract:                   sample.EthAddress().Hex(),
			Connector:                        sample.EthAddress().Hex(),
			DelegationManagerProxy:           sample.EthAddress().Hex(),
			DelegationManagerInteractorProxy: sample.EthAddress().Hex(),
			StrategyManagerProxy:             sample.EthAddress().Hex(),
			SlasherProxy:                     sample.EthAddress().Hex(),
			DvsDirectoryProxy:                sample.EthAddress().Hex(),
			RegistryRouter:                   sample.EthAddress().Hex(),
			RegistryRouterFactory:            sample.EthAddress().Hex(),
		}
		k.SetSystemContract(ctx, sc)
		res, err := k.SystemContract(ctx, &types.QueryGetSystemContractRequest{})
		require.NoError(t, err)
		require.Equal(t, &types.SystemContractResponse{
			SystemContract: sc,
		}, res)
	})
}
