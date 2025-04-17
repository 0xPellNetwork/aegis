package main_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/app"
	pellcored "github.com/pell-chain/pellcore/cmd/pellcored"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	emissionstypes "github.com/pell-chain/pellcore/x/emissions/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func setConfig(t *testing.T) {
	defer func(t *testing.T) {
		if r := recover(); r != nil {
			t.Log("config is already sealed", r)
		}
	}(t)
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.Seal()
}

func Test_ModifyCrossChainState(t *testing.T) {
	setConfig(t)
	t.Run("successfully modify cross chain state to reduce data", func(t *testing.T) {
		cdc := keepertest.NewCodec()
		appState := sample.AppState(t)
		importData := GetImportData(t, cdc, 100)
		err := pellcored.ModifyCrosschainState(appState, importData, cdc)
		require.NoError(t, err)

		modifiedCrosschainAppState := xmsgtypes.GetGenesisStateFromAppState(cdc, appState)
		t.Log("modifiedCrosschainAppState", modifiedCrosschainAppState)
		// TODO: because now not export this data
		//require.Len(t, modifiedCrosschainAppState.Xmsgs, 10)
		//require.Len(t, modifiedCrosschainAppState.InTxHashToXmsgList, 10)
		//require.Len(t, modifiedCrosschainAppState.FinalizedInbounds, 10)
	})

	t.Run("successfully modify cross chain state without changing data when not needed", func(t *testing.T) {
		cdc := keepertest.NewCodec()
		appState := sample.AppState(t)
		importData := GetImportData(t, cdc, 8)
		err := pellcored.ModifyCrosschainState(appState, importData, cdc)
		require.NoError(t, err)

		modifiedCrosschainAppState := xmsgtypes.GetGenesisStateFromAppState(cdc, appState)
		t.Log("modifiedCrosschainAppState", modifiedCrosschainAppState)
		// TODO: because now not export this data
		//require.Len(t, modifiedCrosschainAppState.Xmsgs, 8)
		//require.Len(t, modifiedCrosschainAppState.InTxHashToXmsgList, 8)
		//require.Len(t, modifiedCrosschainAppState.FinalizedInbounds, 8)
	})
}

func Test_ModifyObserverState(t *testing.T) {
	setConfig(t)
	t.Run("successfully modify observer state to reduce data", func(t *testing.T) {
		cdc := keepertest.NewCodec()
		appState := sample.AppState(t)
		importData := GetImportData(t, cdc, 100)
		err := pellcored.ModifyObserverState(appState, importData, cdc)
		require.NoError(t, err)

		modifiedObserverAppState := observertypes.GetGenesisStateFromAppState(cdc, appState)
		require.Len(t, modifiedObserverAppState.Ballots, pellcored.MaxItemsForList)
		require.Len(t, modifiedObserverAppState.NonceToXmsg, pellcored.MaxItemsForList)
	})

	t.Run("successfully modify observer state without changing data when not needed", func(t *testing.T) {
		cdc := keepertest.NewCodec()
		appState := sample.AppState(t)
		importData := GetImportData(t, cdc, 8)
		err := pellcored.ModifyObserverState(appState, importData, cdc)
		require.NoError(t, err)

		modifiedObserverAppState := observertypes.GetGenesisStateFromAppState(cdc, appState)
		require.Len(t, modifiedObserverAppState.Ballots, 8)
		require.Len(t, modifiedObserverAppState.NonceToXmsg, 8)
	})

}

func Test_ImportDataIntoFile(t *testing.T) {
	setConfig(t)
	cdc := keepertest.NewCodec()
	appGenesis := sample.AppGenesis(t)
	importAppGenesis := ImportAppGenesis(t, cdc, 100)

	err := pellcored.ImportDataIntoFile(appGenesis, importAppGenesis, cdc, true)
	require.NoError(t, err)

	types.GenesisDocFromJSON(appGenesis.AppState)

	var appState map[string]json.RawMessage
	err = json.Unmarshal(appGenesis.AppState, &appState)
	require.NoError(t, err)

	// Crosschain module is in Modify list
	crosschainStateAfterImport := xmsgtypes.GetGenesisStateFromAppState(cdc, appState)
	t.Log("crosschainStateAfterImport", crosschainStateAfterImport)
	// TODO: because now not export this data
	//require.Len(t, crosschainStateAfterImport.Xmsgs, pellcored.MaxItemsForList)
	//require.Len(t, crosschainStateAfterImport.InTxHashToXmsgList, pellcored.MaxItemsForList)
	//require.Len(t, crosschainStateAfterImport.FinalizedInbounds, pellcored.MaxItemsForList)

	// Bank module is in Skip list
	var bankStateAfterImport banktypes.GenesisState
	if appState[banktypes.ModuleName] != nil {
		err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankStateAfterImport)
		if err != nil {
			panic(fmt.Sprintf("Failed to get genesis state from app state: %s", err.Error()))
		}
	}
	// 4 balances were present in the original genesis state
	// TODO: because now not export this data
	//require.Len(t, bankStateAfterImport.Balances, 11)

	// Emissions module is in Copy list
	var emissionStateAfterImport emissionstypes.GenesisState
	if appState[emissionstypes.ModuleName] != nil {
		err := cdc.UnmarshalJSON(appState[emissionstypes.ModuleName], &emissionStateAfterImport)
		if err != nil {
			panic(fmt.Sprintf("Failed to get genesis state from app state: %s", err.Error()))
		}
	}
	// TODO: because now not export this data
	//require.Len(t, emissionStateAfterImport.WithdrawableEmissions, 100)
}

func ImportAppGenesis(t *testing.T, cdc *codec.ProtoCodec, n int) *genutiltypes.AppGenesis {
	importAppGenesis := sample.AppGenesis(t)
	importStateJson, err := json.Marshal(GetImportData(t, cdc, n))
	require.NoError(t, err)
	importAppGenesis.AppState = importStateJson
	return importAppGenesis
}

func GetImportData(t *testing.T, cdc *codec.ProtoCodec, n int) map[string]json.RawMessage {
	importData := sample.AppState(t)

	// Add crosschain data to genesis state
	importedCrossChainGenState := xmsgtypes.GetGenesisStateFromAppState(cdc, importData)
	xmsgList := make([]*xmsgtypes.Xmsg, n)
	intxHashToXmsgList := make([]xmsgtypes.InTxHashToXmsg, n)
	finalLizedInbounds := make([]string, n)
	for i := 0; i < n; i++ {
		xmsgList[i] = sample.Xmsg_pell(t, fmt.Sprintf("crosschain-%d", i))
		intxHashToXmsgList[i] = sample.InTxHashToXmsg_pell(t, fmt.Sprintf("intxHashToXmsgList-%d", i))
		finalLizedInbounds[i] = fmt.Sprintf("finalLizedInbounds-%d", i)
	}

	importedCrossChainGenState.Xmsgs = xmsgList
	importedCrossChainGenState.InTxHashToXmsgList = intxHashToXmsgList
	importedCrossChainGenState.FinalizedInbounds = finalLizedInbounds
	importedCrossChainStateBz, err := cdc.MarshalJSON(&importedCrossChainGenState)
	require.NoError(t, err)
	importData[xmsgtypes.ModuleName] = importedCrossChainStateBz

	// Add observer data to genesis state
	importedObserverGenState := observertypes.GetGenesisStateFromAppState(cdc, importData)
	ballots := make([]*observertypes.Ballot, n)
	nonceToXmsg := make([]observertypes.NonceToXmsg, n)
	for i := 0; i < n; i++ {
		ballots[i] = sample.Ballot_pell(t, fmt.Sprintf("ballots-%d", i))
		nonceToXmsg[i] = sample.NonceToXmsg_pell(t, fmt.Sprintf("nonceToXmsg-%d", i))
	}
	importedObserverGenState.Ballots = ballots
	importedObserverGenState.NonceToXmsg = nonceToXmsg
	importedObserverStateBz, err := cdc.MarshalJSON(&importedObserverGenState)
	require.NoError(t, err)
	importData[observertypes.ModuleName] = importedObserverStateBz

	// Add emission data to genesis state
	var importedEmissionGenesis emissionstypes.GenesisState
	if importData[emissionstypes.ModuleName] != nil {
		err := cdc.UnmarshalJSON(importData[emissionstypes.ModuleName], &importedEmissionGenesis)
		if err != nil {
			panic(fmt.Sprintf("Failed to get genesis state from app state: %s", err.Error()))
		}
	}
	withdrawableEmissions := make([]emissionstypes.WithdrawableEmissions, n)
	for i := 0; i < n; i++ {
		withdrawableEmissions[i] = sample.WithdrawableEmissions(t)
	}
	importedEmissionGenesis.WithdrawableEmissions = withdrawableEmissions
	importedEmissionGenesisBz, err := cdc.MarshalJSON(&importedEmissionGenesis)
	require.NoError(t, err)
	importData[emissionstypes.ModuleName] = importedEmissionGenesisBz

	// Add bank data to genesis state
	var importedBankGenesis banktypes.GenesisState
	if importData[banktypes.ModuleName] != nil {
		err := cdc.UnmarshalJSON(importData[banktypes.ModuleName], &importedBankGenesis)
		if err != nil {
			panic(fmt.Sprintf("Failed to get genesis state from app state: %s", err.Error()))
		}
	}
	balances := make([]banktypes.Balance, n)
	supply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()))
	for i := 0; i < n; i++ {
		balances[i] = banktypes.Balance{
			Address: sample.AccAddress(),
			Coins:   sample.Coins(),
		}
		supply = supply.Add(balances[i].Coins...)
	}
	importedBankGenesis.Balances = balances
	importedBankGenesis.Supply = supply
	importedBankGenesisBz, err := cdc.MarshalJSON(&importedBankGenesis)
	require.NoError(t, err)
	importData[banktypes.ModuleName] = importedBankGenesisBz

	return importData
}
