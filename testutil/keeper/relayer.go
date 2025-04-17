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
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	relayermocks "github.com/0xPellNetwork/aegis/testutil/keeper/mocks/relayer"
	"github.com/0xPellNetwork/aegis/x/relayer/keeper"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// RelayerMockOptions represents options for instantiating an relayer keeper with mocks
type RelayerMockOptions struct {
	UseStakingMock     bool
	UseSlashingMock    bool
	UseAuthorityMock   bool
	UseLightclientMock bool
	UsePevmMock        bool
}

var (
	RelayerMocksAll = RelayerMockOptions{
		UseStakingMock:     true,
		UseSlashingMock:    true,
		UseAuthorityMock:   true,
		UseLightclientMock: true,
		UsePevmMock:        true,
	}
	RelayerNoMocks = RelayerMockOptions{}
)

func initRelayerKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	stakingKeeper stakingkeeper.Keeper,
	slashingKeeper slashingkeeper.Keeper,
	paramKeeper paramskeeper.Keeper,
	authorityKeeper types.AuthorityKeeper,
	lightclientKeeper types.LightclientKeeper,
) *keeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ss.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, db)

	return keeper.NewKeeper(
		cdc,
		storeKey,
		memKey,
		paramKeeper.Subspace(types.ModuleName),
		stakingKeeper,
		slashingKeeper,
		authorityKeeper,
		lightclientKeeper,
	)
}

// RelayerKeeperWithMocks instantiates an relayer keeper for testing purposes with the option to mock specific keepers
func RelayerKeeperWithMocks(t testing.TB, mockOptions RelayerMockOptions) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Initialize local store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cdc := NewCodec()
	amino := codec.NewLegacyAmino()
	authorityKeeperTmp := initAuthorityKeeper(cdc, db, stateStore)
	lightclientKeeperTmp := initLightclientKeeper(cdc, db, stateStore, authorityKeeperTmp)

	// Create regular keepers
	sdkKeepers := NewSDKKeepers(cdc, amino, db, stateStore)

	// Create the relayer keeper
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := NewContext(stateStore)

	// Initialize modules genesis
	sdkKeepers.InitGenesis(ctx)

	// Add a proposer to the context
	ctx = sdkKeepers.InitBlockProposer(t, ctx)

	// Initialize mocks for mocked keepers
	var stakingKeeper types.StakingKeeper = sdkKeepers.StakingKeeper
	var slashingKeeper types.SlashingKeeper = sdkKeepers.SlashingKeeper
	var authorityKeeper types.AuthorityKeeper = authorityKeeperTmp
	var lightclientKeeper types.LightclientKeeper = lightclientKeeperTmp
	if mockOptions.UseStakingMock {
		stakingKeeper = relayermocks.NewRelayerStakingKeeper(t)
	}
	if mockOptions.UseSlashingMock {
		slashingKeeper = relayermocks.NewRelayerSlashingKeeper(t)
	}
	if mockOptions.UseAuthorityMock {
		authorityKeeper = relayermocks.NewRelayerAuthorityKeeper(t)
	}
	if mockOptions.UseLightclientMock {
		lightclientKeeper = relayermocks.NewRelayerLightclientKeeper(t)
	}

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		sdkKeepers.ParamsKeeper.Subspace(types.ModuleName),
		stakingKeeper,
		slashingKeeper,
		authorityKeeper,
		lightclientKeeper,
	)

	k.SetParams(ctx, types.DefaultParams())

	pevmKeeperTmp := initPevmKeeper(cdc, db, stateStore,
		sdkKeepers.AuthKeeper, sdkKeepers.BankKeeper, sdkKeepers.EvmKeeper, k, authorityKeeper)
	var pevmKeeper types.PevmKeeper = pevmKeeperTmp
	if mockOptions.UsePevmMock {
		pevmKeeper = relayermocks.NewRelayerPevmKeeper(t)
	}
	k.SetPevmKeeper(pevmKeeper)

	return k, ctx, sdkKeepers, PellKeepers{
		AuthorityKeeper: &authorityKeeperTmp,
	}
}

// RelayerKeeper instantiates an relayer keeper for testing purposes
func RelayerKeeper(t testing.TB) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	return RelayerKeeperWithMocks(t, RelayerNoMocks)
}

// GetRelayerLightclientMock returns a new relayer lightclient keeper mock
func GetRelayerLightclientMock(t testing.TB, keeper *keeper.Keeper) *relayermocks.RelayerLightclientKeeper {
	cok, ok := keeper.GetLightclientKeeper().(*relayermocks.RelayerLightclientKeeper)
	require.True(t, ok)
	return cok
}

// GetRelayerAuthorityMock returns a new relayer authority keeper mock
func GetRelayerAuthorityMock(t testing.TB, keeper *keeper.Keeper) *relayermocks.RelayerAuthorityKeeper {
	cok, ok := keeper.GetAuthorityKeeper().(*relayermocks.RelayerAuthorityKeeper)
	require.True(t, ok)
	return cok
}

// GetRelayerStakingMock returns a new relayer staking keeper mock
func GetRelayerStakingMock(t testing.TB, keeper *keeper.Keeper) *RelayerMockStakingKeeper {
	k, ok := keeper.GetStakingKeeper().(*relayermocks.RelayerStakingKeeper)
	require.True(t, ok)
	return &RelayerMockStakingKeeper{
		RelayerStakingKeeper: k,
	}
}

// GetRelayerPevmMock returns a new relayer pevm keeper mock
func GetObserverPevmMock(t testing.TB, keeper *keeper.Keeper) *relayermocks.RelayerPevmKeeper {
	cok, ok := keeper.GetPevmKeeper().(*relayermocks.RelayerPevmKeeper)
	require.True(t, ok)
	return cok
}

// RelayerMockStakingKeeper is a wrapper of the relayer staking keeper mock that add methods to mock the GetValidator method
type RelayerMockStakingKeeper struct {
	*relayermocks.RelayerStakingKeeper
}

func (m *RelayerMockStakingKeeper) MockGetValidator(validator stakingtypes.Validator) {
	m.On("GetValidator", mock.Anything, mock.Anything).Return(validator, nil)
}

// GetRelayerSlashingMock returns a new relayer slashing keeper mock
func GetRelayerSlashingMock(t testing.TB, keeper *keeper.Keeper) *RelayerMockSlashingKeeper {
	k, ok := keeper.GetSlashingKeeper().(*relayermocks.RelayerSlashingKeeper)
	require.True(t, ok)
	return &RelayerMockSlashingKeeper{
		RelayerSlashingKeeper: k,
	}
}

// RelayerMockSlashingKeeper is a wrapper of the relayer slashing keeper mock that add methods to mock the IsTombstoned method
type RelayerMockSlashingKeeper struct {
	*relayermocks.RelayerSlashingKeeper
}

func (m *RelayerMockSlashingKeeper) MockIsTombstoned(isTombstoned bool) {
	m.On("IsTombstoned", mock.Anything, mock.Anything).Return(isTombstoned)
}
