package observer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/nullify"
	"github.com/pell-chain/pellcore/testutil/sample"
	observer "github.com/pell-chain/pellcore/x/relayer"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestGenesis(t *testing.T) {
	t.Run("genState fields defined", func(t *testing.T) {
		params := types.DefaultParams()
		tss := sample.Tss_pell()
		genesisState := types.GenesisState{
			Params:    &params,
			Tss:       &tss,
			BlameList: sample.BlameRecordsList_pell(t, 10),
			Ballots: []*types.Ballot{
				sample.Ballot_pell(t, "0"),
				sample.Ballot_pell(t, "1"),
				sample.Ballot_pell(t, "2"),
			},
			Observers: sample.ObserverSet_pell(3),
			NodeAccountList: []*types.NodeAccount{
				sample.NodeAccount_pell(),
				sample.NodeAccount_pell(),
				sample.NodeAccount_pell(),
			},
			CrosschainFlags:   types.DefaultCrosschainFlags(),
			Keygen:            sample.Keygen_pell(t),
			ChainParamsList:   sample.ChainParamsList_pell(),
			LastObserverCount: sample.LastObserverCount_pell(10),
			TssFundMigrators:  []types.TssFundMigratorInfo{sample.TssFundsMigrator_pell(1), sample.TssFundsMigrator_pell(2)},
			ChainNonces: []types.ChainNonces{
				sample.ChainNonces_pell(t, "0"),
				sample.ChainNonces_pell(t, "1"),
				sample.ChainNonces_pell(t, "2"),
			},
			PendingNonces: sample.PendingNoncesList_pell(t, "sample", 20),
			NonceToXmsg:   sample.NonceToXmsgList_pell(t, "sample", 20),
			TssHistory:    []types.TSS{sample.Tss_pell()},
		}

		// Init and export
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		observer.InitGenesis(ctx, *k, genesisState)
		got := observer.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		// Compare genesis after init and export
		nullify.Fill(&genesisState)
		nullify.Fill(got)
		require.Equal(t, genesisState, *got)
	})

	t.Run("genState fields not defined", func(t *testing.T) {
		genesisState := types.GenesisState{}

		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		observer.InitGenesis(ctx, *k, genesisState)
		got := observer.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		defaultParams := types.DefaultParams()
		goerliChainParams := types.GetDefaultGoerliLocalnetChainParams()
		goerliChainParams.IsSupported = true
		bscTestChainParams := types.GetDefaultBscTestnetChainParams()
		bscTestChainParams.IsSupported = true
		pellPrivnetChainParams := types.GetDefaultPellPrivnetChainParams()
		pellPrivnetChainParams.IsSupported = true
		localnetChainParams := types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				goerliChainParams,
				bscTestChainParams,
				pellPrivnetChainParams,
			},
		}
		expectedGenesisState := types.GenesisState{
			Params:            &defaultParams,
			CrosschainFlags:   types.DefaultCrosschainFlags(),
			ChainParamsList:   localnetChainParams,
			Tss:               &types.TSS{},
			Keygen:            &types.Keygen{},
			LastObserverCount: &types.LastRelayerCount{},
			NodeAccountList:   []*types.NodeAccount{},
		}

		// TODO: ensure the default chain params
		t.Log("expectedGenesisState", expectedGenesisState, "got", got)
		//require.Equal(t, expectedGenesisState, *got)
	})

	t.Run("genState fields not defined except tss", func(t *testing.T) {
		tss := sample.Tss_pell()
		genesisState := types.GenesisState{
			Tss: &tss,
		}

		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		observer.InitGenesis(ctx, *k, genesisState)
		got := observer.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		defaultParams := types.DefaultParams()
		goerliChainParams := types.GetDefaultGoerliLocalnetChainParams()
		goerliChainParams.IsSupported = true
		bscTestChainParams := types.GetDefaultBscTestnetChainParams()
		bscTestChainParams.IsSupported = true
		pellPrivnetChainParams := types.GetDefaultPellPrivnetChainParams()
		pellPrivnetChainParams.IsSupported = true

		localnetChainParams := types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				goerliChainParams,
				bscTestChainParams,
				pellPrivnetChainParams,
			},
		}
		pendingNonces, err := k.GetAllPendingNonces(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, pendingNonces)
		expectedGenesisState := types.GenesisState{
			Params:            &defaultParams,
			CrosschainFlags:   types.DefaultCrosschainFlags(),
			ChainParamsList:   localnetChainParams,
			Tss:               &tss,
			Keygen:            &types.Keygen{},
			LastObserverCount: &types.LastRelayerCount{},
			NodeAccountList:   []*types.NodeAccount{},
			PendingNonces:     pendingNonces,
		}

		// TODO: ensure the default chain params
		t.Log("expectedGenesisState", expectedGenesisState, "got", got)
		//require.Equal(t, expectedGenesisState, *got)
	})

	t.Run("export without init", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		got := observer.ExportGenesis(ctx, *k)
		require.NotNil(t, got)

		params := k.GetParamsIfExists(ctx)
		expectedGenesisState := types.GenesisState{
			Params:            &params,
			CrosschainFlags:   types.DefaultCrosschainFlags(),
			ChainParamsList:   types.ChainParamsList{},
			Tss:               &types.TSS{},
			Keygen:            &types.Keygen{},
			LastObserverCount: &types.LastRelayerCount{},
			NodeAccountList:   []*types.NodeAccount{},
			Ballots:           k.GetAllBallots(ctx),
			TssHistory:        k.GetAllTSS(ctx),
			TssFundMigrators:  k.GetAllTssFundMigrators(ctx),
			BlameList:         k.GetAllBlame(ctx),
			ChainNonces:       k.GetAllChainNonces(ctx),
			NonceToXmsg:       k.GetAllNonceToXmsg(ctx),
		}

		require.Equal(t, expectedGenesisState, *got)
	})
}
