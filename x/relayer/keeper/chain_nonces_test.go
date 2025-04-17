package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
)

func TestKeeper_GetChainNonces(t *testing.T) {
	t.Run("Get chain nonces", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		chainNoncesList := sample.ChainNoncesList_pell(t, 10)
		for _, n := range chainNoncesList {
			k.SetChainNonces(ctx, n)
		}
		for _, n := range chainNoncesList {
			rst, found := k.GetChainNonces(ctx, n.Index)
			require.True(t, found)
			require.Equal(t, n, rst)
		}
	})
	t.Run("Get chain nonces not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		chainNoncesList := sample.ChainNoncesList_pell(t, 10)
		for _, n := range chainNoncesList {
			k.SetChainNonces(ctx, n)
		}
		_, found := k.GetChainNonces(ctx, "not_found")
		require.False(t, found)
	})
	t.Run("Get all chain nonces", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		chainNoncesList := sample.ChainNoncesList_pell(t, 10)
		for _, n := range chainNoncesList {
			k.SetChainNonces(ctx, n)
		}
		rst := k.GetAllChainNonces(ctx)
		require.Equal(t, chainNoncesList, rst)
	})
}
