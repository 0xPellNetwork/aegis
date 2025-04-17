package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	restakingmocks "github.com/0xPellNetwork/aegis/testutil/keeper/mocks/restaking"
	"github.com/0xPellNetwork/aegis/x/restaking/keeper"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

type RestakingMockOptions struct {
	UseBankMock      bool
	UseAccountMock   bool
	UseAuthorityMock bool
	UseEvmMock       bool
	UseRelayerMock   bool
	UsePevmMock      bool
}

var (
	RestakingMocksAll = RestakingMockOptions{
		UseBankMock:      true,
		UseAccountMock:   true,
		UseAuthorityMock: true,
		UseEvmMock:       true,
		UseRelayerMock:   true,
		UsePevmMock:      true,
	}
	RestakingNoMocks = RestakingMockOptions{}
)

// **RestakingKeeperWithMocks initializes a `restaking` keeper for testing with optional mocked dependencies.**
func RestakingKeeperWithMocks(
	t testing.TB,
	mockOptions RestakingMockOptions,
) (*keeper.Keeper, sdk.Context) {
	SetConfig(false)
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

	// **Initialize default keepers**
	var (
		accountKeeper   types.AccountKeeper
		evmKeeper       types.EVMKeeper
		bankKeeper      types.BankKeeper
		relayerKeeper   types.RelayerKeeper
		authorityKeeper types.AuthorityKeeper
		pevmKeeper      types.PevmKeeper
	)

	// **Apply mock options**
	if mockOptions.UseAccountMock {
		accountKeeper = restakingmocks.NewRestakingAccountKeeper(t)
	}
	if mockOptions.UseEvmMock {
		evmKeeper = restakingmocks.NewRestakingEVMKeeper(t)
	}
	if mockOptions.UseBankMock {
		bankKeeper = restakingmocks.NewRestakingBankKeeper(t)
	}
	if mockOptions.UseRelayerMock {
		relayerKeeper = restakingmocks.NewRestakingRelayerKeeper(t)
	}
	if mockOptions.UseAuthorityMock {
		authorityKeeper = restakingmocks.NewRestakingAuthorityKeeper(t)
	}
	if mockOptions.UsePevmMock {
		pevmKeeper = restakingmocks.NewRestakingPevmKeeper(t)
	}

	// **Handle additional mock behavior if needed**
	if mockOptions.UseRelayerMock {
		relayerKeeperMock, ok := relayerKeeper.(*restakingmocks.RestakingRelayerKeeper)
		require.True(t, ok, "relayerKeeper should be a *restakingmocks.RestakingRelayerKeeper")
		relayerKeeperMock.On("SetPevmKeeper", mock.Anything).Return()
	}

	// **Create the keeper instance**
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		accountKeeper,
		evmKeeper,
		bankKeeper,
		relayerKeeper,
		authorityKeeper,
		pevmKeeper,
		nil,
	)

	return k, ctx
}

// **RestakingKeeperAllMocks initializes a `restaking` keeper for testing with all mocks.**
func RestakingKeeperAllMocks(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx := RestakingKeeperWithMocks(t, RestakingMocksAll)
	return k, ctx
}

// **RestakingKeeper initializes a `restaking` keeper for testing without mocks.**
func RestakingKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx := RestakingKeeperWithMocks(t, RestakingNoMocks)
	return k, ctx
}

// **Mock GetAuthorityKeeper**
func GetRestakingAuthorityMock(t testing.TB, keeper *keeper.Keeper) *restakingmocks.RestakingAuthorityKeeper {
	ak, ok := keeper.GetAuthorityKeeper().(*restakingmocks.RestakingAuthorityKeeper)
	require.True(t, ok)
	return ak
}

// **Mock GetBankKeeper**
func GetRestakingBankMock(t testing.TB, keeper *keeper.Keeper) *restakingmocks.RestakingBankKeeper {
	bk, ok := keeper.GetBankKeeper().(*restakingmocks.RestakingBankKeeper)
	require.True(t, ok)
	return bk
}

// **Mock GetEVMKeeper**
func GetRestakingEvmMock(t testing.TB, keeper *keeper.Keeper) *restakingmocks.RestakingEVMKeeper {
	evm, ok := keeper.GetEVMKeeper().(*restakingmocks.RestakingEVMKeeper)
	require.True(t, ok)
	return evm
}

// **Mock GetRelayerKeeper**
func GetRestakingRelayerMock(t testing.TB, keeper *keeper.Keeper) *restakingmocks.RestakingRelayerKeeper {
	rel, ok := keeper.GetRelayerKeeper().(*restakingmocks.RestakingRelayerKeeper)
	require.True(t, ok)
	return rel
}
