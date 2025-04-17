package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestKeeper_GetSupportedChainFromChainID(t *testing.T) {
	t.Run("return nil if chain not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		// no core params list
		require.Nil(t, k.GetSupportedChainFromChainID(ctx, getValidEthChainIDWithIndex(t, 0)))

		// core params list but chain not in list
		setSupportedChain(ctx, *k, getValidEthChainIDWithIndex(t, 0))
		require.Nil(t, k.GetSupportedChainFromChainID(ctx, getValidEthChainIDWithIndex(t, 1)))

		// chain params list but chain not supported
		chainParams := sample.ChainParams_pell(getValidEthChainIDWithIndex(t, 0))
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{chainParams},
		})
		require.Nil(t, k.GetSupportedChainFromChainID(ctx, getValidEthChainIDWithIndex(t, 0)))
	})

	t.Run("return chain if chain found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		chainID := getValidEthChainIDWithIndex(t, 0)
		setSupportedChain(ctx, *k, getValidEthChainIDWithIndex(t, 1), chainID)
		chain := k.GetSupportedChainFromChainID(ctx, chainID)
		require.NotNil(t, chain)
		require.EqualValues(t, chainID, chain.Id)
	})
}

func TestKeeper_GetChainParamsByChainID(t *testing.T) {
	t.Run("return false if chain params not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		_, found := k.GetChainParamsByChainID(ctx, getValidEthChainIDWithIndex(t, 0))
		require.False(t, found)
	})

	t.Run("return true if found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		chainParams := sample.ChainParams_pell(getValidEthChainIDWithIndex(t, 0))
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{chainParams},
		})
		res, found := k.GetChainParamsByChainID(ctx, getValidEthChainIDWithIndex(t, 0))
		require.True(t, found)
		require.Equal(t, chainParams, res)
	})

	t.Run("return false if chain not found in params", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		chainParams := sample.ChainParams_pell(getValidEthChainIDWithIndex(t, 0))
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{chainParams},
		})
		_, found := k.GetChainParamsByChainID(ctx, getValidEthChainIDWithIndex(t, 1))
		require.False(t, found)
	})
}

func TestKeeper_GetSupportedChains(t *testing.T) {
	t.Run("return empty list if no core params list", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		require.Empty(t, k.GetSupportedChains(ctx))
	})

	t.Run("return list containing supported chains", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		externalChainList := chains.ExternalChainList()
		require.Greater(t, len(externalChainList), 5)
		supported1 := externalChainList[0]
		supported2 := externalChainList[1]
		unsupported := externalChainList[2]
		supported3 := externalChainList[3]
		supported4 := externalChainList[4]

		var chainParamsList []*types.ChainParams
		chainParamsList = append(chainParamsList, sample.ChainParamsSupported_pell(supported1.Id))
		chainParamsList = append(chainParamsList, sample.ChainParamsSupported_pell(supported2.Id))
		chainParamsList = append(chainParamsList, sample.ChainParams_pell(unsupported.Id))
		chainParamsList = append(chainParamsList, sample.ChainParamsSupported_pell(supported3.Id))
		chainParamsList = append(chainParamsList, sample.ChainParamsSupported_pell(supported4.Id))

		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: chainParamsList,
		})

		supportedChains := k.GetSupportedChains(ctx)

		require.Len(t, supportedChains, 4)
		require.EqualValues(t, supported1.Id, supportedChains[0].Id)
		require.EqualValues(t, supported2.Id, supportedChains[1].Id)
		require.EqualValues(t, supported3.Id, supportedChains[2].Id)
		require.EqualValues(t, supported4.Id, supportedChains[3].Id)
	})
}
