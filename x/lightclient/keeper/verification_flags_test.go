package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

func TestKeeper_GetVerificationFlags(t *testing.T) {
	t.Run("can get and set verification flags", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		vf, found := k.GetVerificationFlags(ctx)
		require.False(t, found)

		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: false,
			BtcTypeChainEnabled: true,
		})
		vf, found = k.GetVerificationFlags(ctx)
		require.True(t, found)
		require.Equal(t, types.VerificationFlags{
			EthTypeChainEnabled: false,
			BtcTypeChainEnabled: true,
		}, vf)
	})
}

func TestKeeper_CheckVerificationFlagsEnabled(t *testing.T) {
	t.Run("can check verification flags with ethereum enabled", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)
		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: true,
			BtcTypeChainEnabled: false,
		})

		err := k.CheckVerificationFlagsEnabled(ctx, chains.EthChain().Id)
		require.NoError(t, err)

		err = k.CheckVerificationFlagsEnabled(ctx, 1000)
		require.Error(t, err)
		require.ErrorContains(t, err, "doesn't support block header verification")
	})

	t.Run("can check verification flags with bitcoin enabled", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)
		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: false,
			BtcTypeChainEnabled: true,
		})

		err := k.CheckVerificationFlagsEnabled(ctx, chains.EthChain().Id)
		require.Error(t, err)
		require.ErrorContains(t, err, "proof verification not enabled for evm")

		err = k.CheckVerificationFlagsEnabled(ctx, 1000)
		require.Error(t, err)
		require.ErrorContains(t, err, "doesn't support block header verification")
	})
}
