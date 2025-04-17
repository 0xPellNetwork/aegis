package pevm_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/nullify"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/pevm"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		SystemContract: sample.SystemContract_pell(),
	}

	// Init and export
	k, ctx, _, _ := keepertest.PevmKeeper(t)
	pevm.InitGenesis(ctx, *k, genesisState)
	got := pevm.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// Compare genesis after init and export
	nullify.Fill(&genesisState)
	nullify.Fill(got)
	require.Equal(t, genesisState, *got)
}
