package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

// Keeper Tests
func createTestKeygen(keeper *keeper.Keeper, ctx sdk.Context) types.Keygen {
	item := types.Keygen{
		BlockNumber: 10,
	}
	keeper.SetKeygen(ctx, item)
	return item
}

func TestKeygenGet(t *testing.T) {
	k, ctx, _, _ := keepertest.RelayerKeeper(t)
	item := createTestKeygen(k, ctx)
	rst, found := k.GetKeygen(ctx)
	require.True(t, found)
	require.Equal(t, item, rst)
}

func TestKeygenRemove(t *testing.T) {
	k, ctx, _, _ := keepertest.RelayerKeeper(t)
	createTestKeygen(k, ctx)
	k.RemoveKeygen(ctx)
	_, found := k.GetKeygen(ctx)
	require.False(t, found)
}
