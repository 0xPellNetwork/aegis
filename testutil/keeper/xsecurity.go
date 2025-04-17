package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	xsecuritymocks "github.com/pell-chain/pellcore/testutil/keeper/mocks/xsecurity"
	"github.com/pell-chain/pellcore/x/xsecurity/keeper"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

type XSecurityMockOptions struct {
	UseStakingMock      bool
	UseSlashingMock     bool
	UseStakingHooksMock bool
	UseAuthorityMock    bool
	UseRelayerMock      bool
	UsePevmMock         bool
	UseRestakingMock    bool
}

var (
	XSecurityMocksAll = XSecurityMockOptions{
		UseStakingMock:      true,
		UseSlashingMock:     true,
		UseStakingHooksMock: true,
		UseAuthorityMock:    true,
		UseRelayerMock:      true,
		UsePevmMock:         true,
		UseRestakingMock:    true,
	}

	XSecurityNoMocks = XSecurityMockOptions{}
)

// XSecurityKeeperWithMocks initializes a `xsecurity` keeper for testing with optional mocked dependencies.
func XSecurityKeeperWithMocks(
	t testing.TB,
	mockOptions XSecurityMockOptions,
) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Initialize local store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cdc := NewCodec()

	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := NewContext(stateStore)

	// Initialize default keepers
	var (
		stakingKeeper   types.StakingKeeper
		slashingKeeper  types.SlashingKeeper
		authorityKeeper types.AuthorityKeeper
		relayerKeeper   types.RelayerKeeper
		pevmKeeper      types.PevmKeeper
		restakingKeeper types.RestakingKeeper
	)

	// Apply mock options
	if mockOptions.UseStakingMock {
		stakingKeeper = xsecuritymocks.NewXSecurityStakingKeeper(t)
	}
	if mockOptions.UseSlashingMock {
		slashingKeeper = xsecuritymocks.NewXSecuritySlashingKeeper(t)
	}
	if mockOptions.UseAuthorityMock {
		authorityKeeper = xsecuritymocks.NewXSecurityAuthorityKeeper(t)
	}
	if mockOptions.UseRelayerMock {
		relayerKeeper = xsecuritymocks.NewXSecurityRelayerKeeper(t)
	}
	if mockOptions.UsePevmMock {
		pevmKeeper = xsecuritymocks.NewXSecurityPevmKeeper(t)
	}
	if mockOptions.UseRestakingMock {
		restakingKeeper = xsecuritymocks.NewXSecurityRestakingKeeper(t)
	}

	// Create the keeper instance
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		stakingKeeper,
		slashingKeeper,
		pevmKeeper,
		relayerKeeper,
		restakingKeeper,
		authorityKeeper,
	)

	return k, ctx
}

// XSecurityKeeperAllMocks initializes a `xsecurity` keeper for testing with all mocks.
func XSecurityKeeperAllMocks(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx := XSecurityKeeperWithMocks(t, XSecurityMocksAll)
	return k, ctx
}

// XSecurityKeeper initializes a `xsecurity` keeper for testing without mocks.
func XSecurityKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx := XSecurityKeeperWithMocks(t, XSecurityNoMocks)
	return k, ctx
}

// GetXSecuritySlashingMock returns the mocked slashing keeper
func GetXSecuritySlashingMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecuritySlashingKeeper {
	sk, ok := keeper.GetSlashingKeeper().(*xsecuritymocks.XSecuritySlashingKeeper)
	require.True(t, ok)
	return sk
}

// GetXSecurityAuthorityMock returns the mocked authority keeper
func GetXSecurityAuthorityMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecurityAuthorityKeeper {
	ak, ok := keeper.GetAuthorityKeeper().(*xsecuritymocks.XSecurityAuthorityKeeper)
	require.True(t, ok)
	return ak
}

// GetXSecurityRelayerMock returns the mocked relayer keeper
func GetXSecurityRelayerMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecurityRelayerKeeper {
	rk, ok := keeper.GetRelayerKeeper().(*xsecuritymocks.XSecurityRelayerKeeper)
	require.True(t, ok)
	return rk
}

// GetXSecurityPevmMock returns the mocked pevm keeper
func GetXSecurityPevmMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecurityPevmKeeper {
	pk, ok := keeper.GetPevmKeeper().(*xsecuritymocks.XSecurityPevmKeeper)
	require.True(t, ok)
	return pk
}

// GetXSecurityRestakingMock returns the mocked restaking keeper
func GetXSecurityRestakingMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecurityRestakingKeeper {
	rk, ok := keeper.GetRestakingKeeper().(*xsecuritymocks.XSecurityRestakingKeeper)
	require.True(t, ok)
	return rk
}

// GetXSecurityStakingMock returns the mocked staking keeper
func GetXSecurityStakingMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecurityStakingKeeper {
	sk, ok := keeper.GetStakingKeeper().(*xsecuritymocks.XSecurityStakingKeeper)
	require.True(t, ok)
	return sk
}

// GetXSecurityPEVMKeeperMock returns the mocked pevm keeper
func GetXSecurityPEVMKeeperMock(t testing.TB, keeper *keeper.Keeper) *xsecuritymocks.XSecurityPevmKeeper {
	pk, ok := keeper.GetPevmKeeper().(*xsecuritymocks.XSecurityPevmKeeper)
	require.True(t, ok)
	return pk
}
