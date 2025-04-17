package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	emissionsmocks "github.com/pell-chain/pellcore/testutil/keeper/mocks/emissions"
	"github.com/pell-chain/pellcore/x/emissions/keeper"
	"github.com/pell-chain/pellcore/x/emissions/types"
)

type EmissionMockOptions struct {
	UseBankMock       bool
	UseStakingMock    bool
	UseObserverMock   bool
	UseAccountMock    bool
	UseParamStoreMock bool
}

func EmissionsKeeper(t testing.TB) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	return EmissionKeeperWithMockOptions(t, EmissionMockOptions{})
}

func EmissionKeeperWithMockOptions(
	t testing.TB,
	mockOptions EmissionMockOptions,
) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	SetConfig(false)
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Initialize local store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cdc := NewCodec()
	amino := codec.NewLegacyAmino()
	// Create regular keepers
	sdkKeepers := NewSDKKeepers(cdc, amino, db, stateStore)

	authorityKeeper := initAuthorityKeeper(cdc, db, stateStore)

	// Create pell keepers
	observerKeeperTmp := initRelayerKeeper(
		cdc,
		db,
		stateStore,
		sdkKeepers.StakingKeeper,
		sdkKeepers.SlashingKeeper,
		sdkKeepers.ParamsKeeper,
		authorityKeeper,
		initLightclientKeeper(cdc, db, stateStore, authorityKeeper),
	)

	pellKeepers := PellKeepers{
		ObserverKeeper: observerKeeperTmp,
	}
	var observerKeeper types.RelayerKeeper = observerKeeperTmp

	// Create the fungible keeper
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := NewContext(stateStore)

	// Initialize modules genesis
	sdkKeepers.InitGenesis(ctx)
	pellKeepers.InitGenesis(ctx)

	// Add a proposer to the context
	ctx = sdkKeepers.InitBlockProposer(t, ctx)

	// Initialize mocks for mocked keepers
	var authKeeper types.AccountKeeper = sdkKeepers.AuthKeeper
	var bankKeeper types.BankKeeper = sdkKeepers.BankKeeper
	var stakingKeeper types.StakingKeeper = sdkKeepers.StakingKeeper
	if mockOptions.UseAccountMock {
		authKeeper = emissionsmocks.NewEmissionAccountKeeper(t)
	}
	if mockOptions.UseBankMock {
		bankKeeper = emissionsmocks.NewEmissionBankKeeper(t)
	}
	if mockOptions.UseStakingMock {
		stakingKeeper = emissionsmocks.NewEmissionStakingKeeper(t)
	}
	if mockOptions.UseObserverMock {
		observerKeeper = emissionsmocks.NewEmissionRelayerKeeper(t)
	}

	var paramStore types.ParamStore
	if mockOptions.UseParamStoreMock {
		mock := emissionsmocks.NewEmissionParamStore(t)
		// mock this method for the keeper constructor
		mock.On("HasKeyTable").Maybe().Return(true)
		paramStore = mock
	} else {
		paramStore = sdkKeepers.ParamsKeeper.Subspace(types.ModuleName)
	}

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramStore,
		authtypes.FeeCollectorName,
		bankKeeper,
		stakingKeeper,
		observerKeeper,
		authKeeper,
	)

	if !mockOptions.UseParamStoreMock {
		k.SetParams(ctx, types.DefaultParams())
	}

	return k, ctx, sdkKeepers, pellKeepers
}

func GetEmissionsBankMock(t testing.TB, keeper *keeper.Keeper) *emissionsmocks.EmissionBankKeeper {
	cbk, ok := keeper.GetBankKeeper().(*emissionsmocks.EmissionBankKeeper)
	require.True(t, ok)
	return cbk
}

func GetEmissionsParamStoreMock(t testing.TB, keeper *keeper.Keeper) *emissionsmocks.EmissionParamStore {
	m, ok := keeper.GetParamStore().(*emissionsmocks.EmissionParamStore)
	require.True(t, ok)
	return m
}

func GetEmissionsStakingMock(t testing.TB, keeper *keeper.Keeper) *emissionsmocks.EmissionStakingKeeper {
	cbk, ok := keeper.GetStakingKeeper().(*emissionsmocks.EmissionStakingKeeper)
	require.True(t, ok)
	return cbk
}
