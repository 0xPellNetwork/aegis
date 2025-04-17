package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestKeeper_Keygen(t *testing.T) {
	t.Run("should error if keygen not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.Keygen(wctx, &types.QueryGetKeygenRequest{})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should return if keygen found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)
		keygen := types.Keygen{
			BlockNumber: 10,
		}
		k.SetKeygen(ctx, keygen)

		res, err := k.Keygen(wctx, &types.QueryGetKeygenRequest{})
		require.NoError(t, err)
		require.Equal(t, &types.QueryKeygenResponse{
			Keygen: &keygen,
		}, res)
	})
}
