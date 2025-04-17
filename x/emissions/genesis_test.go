package emissions_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/nullify"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/emissions"
	"github.com/pell-chain/pellcore/x/emissions/types"
)

func TestGenesis(t *testing.T) {
	params := types.DefaultParams()
	params.ObserverSlashAmount = sdkmath.Int{}

	genesisState := types.GenesisState{
		Params: params,
		WithdrawableEmissions: []types.WithdrawableEmissions{
			sample.WithdrawableEmissions(t),
			sample.WithdrawableEmissions(t),
			sample.WithdrawableEmissions(t),
		},
	}

	// Init and export
	k, ctx, _, _ := keepertest.EmissionsKeeper(t)
	emissions.InitGenesis(ctx, *k, genesisState)
	got := emissions.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// Compare genesis after init and export
	nullify.Fill(&genesisState)
	nullify.Fill(got)
	require.Equal(t, genesisState, *got)
}
