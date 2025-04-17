package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"cosmossdk.io/math"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/app"
	emissionstypes "github.com/pell-chain/pellcore/x/emissions/types"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const MaxItemsForList = 10

// Copy represents a set of modules for which, the entire state is copied without any modifications
var Copy = map[string]bool{
	slashingtypes.ModuleName:  false,
	crisistypes.ModuleName:    false,
	feemarkettypes.ModuleName: true,
	paramstypes.ModuleName:    false,
	upgradetypes.ModuleName:   false,
	evidencetypes.ModuleName:  false,
	vestingtypes.ModuleName:   false,
	emissionstypes.ModuleName: false,
	restakingtypes.ModuleName: true,
}

// Skip represents a set of modules for which, the entire state is skipped and nothing gets imported
var Skip = map[string]bool{
	// Skipping evm this is done to reduce the size of the genesis file evm module uses the majority of the space due to smart contract data
	evmtypes.ModuleName: true,
	// Skipping staking as new validators would be created for the new chain
	stakingtypes.ModuleName: true,
	// Skipping genutil as new gentxs would be created
	genutiltypes.ModuleName: true,
	// Skipping auth as new accounts would be created for the new chain. This also needs to be done as we are skipping evm module
	authtypes.ModuleName: true,
	// Skipping bank module as it is not used when starting a new chain this is done to make sure the total supply invariant is maintained.
	// This would need modification but might be possible to add in non evm based modules in the future
	banktypes.ModuleName: true,
	// Skipping distribution module as it is not used when starting a new chain , rewards are based on validators and delegators , and so rewards from a different chain do not hold any value
	distributiontypes.ModuleName: true,
	// Skipping group module as it is not used when starting a new chain, new groups should be created based on the validator operator keys
	group.ModuleName: true,
	// Skipping authz as it is not used when starting a new chain, new grants should be created based on the validator hotkeys abd operator keys
	authz.ModuleName: true,
	// Skipping fungible module as new fungible tokens would be created and system contract would be deployed
	pevmtypes.ModuleName: true,
	// Skipping gov types as new parameters are set for the new chain
	govtypes.ModuleName: true,
}

// Modify represents a set of modules for which, the state is modified before importing. Each Module should have a corresponding Modify function
var Modify = map[string]bool{
	xmsgtypes.ModuleName:     true,
	observertypes.ModuleName: true,
}

func CmdParseGenesisFile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse-genesis-file [import-genesis-file] [optional-genesis-file]",
		Short: "Parse the provided genesis file and import the required data into the optionally provided genesis file",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec
			modifyEnabled, err := cmd.Flags().GetBool("modify")
			if err != nil {
				return err
			}

			if clientCtx.HomeDir == "" {
				clientCtx.HomeDir = app.DefaultNodeHome
			}

			nodeHome := clientCtx.HomeDir
			if nodeHome == "" {
				nodeHome = app.DefaultNodeHome
			}

			genesisFilePath := filepath.Join(nodeHome, "config", "genesis.json")
			if len(args) == 2 {
				genesisFilePath = args[1]
			}

			_, genesis, err := genutiltypes.GenesisStateFromGenFile(genesisFilePath)
			if err != nil {
				return err
			}

			_, importData, err := genutiltypes.GenesisStateFromGenFile(args[0])
			if err != nil {
				return err
			}

			err = ImportDataIntoFile(genesis, importData, cdc, modifyEnabled)
			if err != nil {
				return err
			}

			err = genutil.ExportGenesisFile(genesis, genesisFilePath)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.PersistentFlags().Bool("modify", false, "modify the genesis file before importing")
	return cmd
}

func ImportDataIntoFile(
	gen *genutiltypes.AppGenesis,
	importFile *genutiltypes.AppGenesis,
	cdc codec.Codec,
	modifyEnabled bool,
) error {
	appState, err := genutiltypes.GenesisStateFromAppGenesis(gen)
	if err != nil {
		return err
	}

	importAppState, err := genutiltypes.GenesisStateFromAppGenesis(importFile)
	if err != nil {
		return err
	}

	moduleList := app.InitGenesisModuleList()
	for _, m := range moduleList {
		if Skip[m] {
			continue
		}
		if Copy[m] {
			appState[m] = importAppState[m]
		}
		if Modify[m] && modifyEnabled {
			switch m {
			case xmsgtypes.ModuleName:
				err := ModifyCrosschainState(appState, importAppState, cdc)
				if err != nil {
					return err
				}
			case observertypes.ModuleName:
				err := ModifyObserverState(appState, importAppState, cdc)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("modify function for %s not found", m)
			}
		}
	}

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}
	gen.AppState = appStateJSON

	return nil
}

// ModifyCrosschainState modifies the crosschain state before importing
// It truncates the crosschain transactions, inbound transactions and finalized inbounds to MaxItemsForList
func ModifyCrosschainState(appState map[string]json.RawMessage, importAppState map[string]json.RawMessage, cdc codec.Codec) error {
	importedCrossChainGenState := xmsgtypes.GetGenesisStateFromAppState(cdc, appState)
	// The genesis state has been modified between the two versions, so we add only the required fields and leave out the rest
	// v16 adds the rate_limiter_flags and removes params from the genesis state
	importedCrossChainGenState.Xmsgs = importedCrossChainGenState.Xmsgs[:math.Min(MaxItemsForList, len(importedCrossChainGenState.Xmsgs))]
	importedCrossChainGenState.InTxHashToXmsgList = importedCrossChainGenState.InTxHashToXmsgList[:math.Min(MaxItemsForList, len(importedCrossChainGenState.InTxHashToXmsgList))]
	importedCrossChainGenState.FinalizedInbounds = importedCrossChainGenState.FinalizedInbounds[:math.Min(MaxItemsForList, len(importedCrossChainGenState.FinalizedInbounds))]
	importedCrossChainStateBz, err := json.Marshal(importedCrossChainGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal zetacrosschain genesis state: %w", err)
	}
	appState[xmsgtypes.ModuleName] = importedCrossChainStateBz
	return nil
}

// ModifyObserverState modifies the observer state before importing
// It truncates the ballots and nonce to xmsg list to MaxItemsForList
func ModifyObserverState(appState map[string]json.RawMessage, importAppState map[string]json.RawMessage, cdc codec.Codec) error {
	importedObserverGenState := observertypes.GetGenesisStateFromAppState(cdc, importAppState)
	importedObserverGenState.Ballots = importedObserverGenState.Ballots[:math.Min(MaxItemsForList, len(importedObserverGenState.Ballots))]
	importedObserverGenState.NonceToXmsg = importedObserverGenState.NonceToXmsg[:math.Min(MaxItemsForList, len(importedObserverGenState.NonceToXmsg))]

	currentGenState := observertypes.GetGenesisStateFromAppState(cdc, appState)
	currentGenState.Ballots = importedObserverGenState.Ballots
	currentGenState.NonceToXmsg = importedObserverGenState.NonceToXmsg

	currentGenStateBz, err := cdc.MarshalJSON(&currentGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal observer genesis state: %w", err)
	}

	appState[observertypes.ModuleName] = currentGenStateBz
	return nil
}
