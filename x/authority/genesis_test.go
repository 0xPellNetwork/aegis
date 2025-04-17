package authority_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/nullify"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/authority"
	"github.com/pell-chain/pellcore/x/authority/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Policies: sample.Policies(),
	}

	// Init
	k, ctx := keepertest.AuthorityKeeper(t)
	authority.InitGenesis(ctx, *k, genesisState)

	// Check policy is set
	policies, found := k.GetPolicies(ctx)
	require.True(t, found)
	require.Equal(t, genesisState.Policies, policies)

	// Export
	got := authority.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// Compare genesis after init and export
	nullify.Fill(&genesisState)
	nullify.Fill(got)
	require.Equal(t, genesisState, *got)
}
