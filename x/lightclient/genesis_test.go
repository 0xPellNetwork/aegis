package lightclient_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/proofs"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/nullify"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/lightclient"
	"github.com/pell-chain/pellcore/x/lightclient/types"
)

func TestGenesis(t *testing.T) {
	t.Run("can import and export genesis", func(t *testing.T) {
		genesisState := types.GenesisState{
			VerificationFlags: types.VerificationFlags{
				EthTypeChainEnabled: false,
				BtcTypeChainEnabled: true,
			},
			BlockHeaders: []proofs.BlockHeader{
				sample.BlockHeader(sample.Hash().Bytes()),
				sample.BlockHeader(sample.Hash().Bytes()),
				sample.BlockHeader(sample.Hash().Bytes()),
			},
			ChainStates: []types.ChainState{
				sample.ChainState(chains.EthChain().Id),
				sample.ChainState(chains.BscMainnetChain().Id),
			},
		}

		// Init and export
		k, ctx, _, _ := keepertest.LightclientKeeper(t)
		lightclient.InitGenesis(ctx, *k, genesisState)
		got := lightclient.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		// Compare genesis after init and export
		nullify.Fill(&genesisState)
		nullify.Fill(got)
		require.Equal(t, genesisState, *got)
	})

	t.Run("can export genesis with empty state", func(t *testing.T) {
		// Export genesis with empty state
		k, ctx, _, _ := keepertest.LightclientKeeper(t)
		got := lightclient.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		// Compare genesis after export
		expected := types.GenesisState{
			VerificationFlags: types.DefaultVerificationFlags(),
			BlockHeaders:      []proofs.BlockHeader{},
			ChainStates:       []types.ChainState{},
		}
		nullify.Fill(got)
		nullify.Fill(expected)
		require.Equal(t, expected, *got)
	})
}
