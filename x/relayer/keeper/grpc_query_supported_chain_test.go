package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
)

func TestKeeper_SupportedChains(t *testing.T) {
	t.Run("should return supported chains", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.SupportedChains(wctx, nil)
		require.NoError(t, err)
		require.Equal(t, []*chains.Chain{}, res.Chains)
	})
}
