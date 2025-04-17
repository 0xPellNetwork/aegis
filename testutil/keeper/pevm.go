package keeper

import (
	"math/big"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pevmmocks "github.com/pell-chain/pellcore/testutil/keeper/mocks/pevm"
	"github.com/pell-chain/pellcore/testutil/sample"
	pevmmodule "github.com/pell-chain/pellcore/x/pevm"
	"github.com/pell-chain/pellcore/x/pevm/keeper"
	"github.com/pell-chain/pellcore/x/pevm/types"
)

type PevmMockOptions struct {
	UseBankMock      bool
	UseAccountMock   bool
	UseObserverMock  bool
	UseEVMMock       bool
	UseAuthorityMock bool
}

var (
	PevmMocksAll = PevmMockOptions{
		UseBankMock:      true,
		UseAccountMock:   true,
		UseObserverMock:  true,
		UseEVMMock:       true,
		UseAuthorityMock: true,
	}
	PevmNoMocks = PevmMockOptions{}
)

func initPevmKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	authKeeper types.AccountKeeper,
	bankKeepr types.BankKeeper,
	evmKeeper types.EVMKeeper,
	observerKeeper types.RelayerKeeper,
	authorityKeeper types.AuthorityKeeper,
) *keeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ss.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, db)

	return keeper.NewKeeper(
		cdc,
		storeKey,
		memKey,
		authKeeper,
		evmKeeper,
		bankKeepr,
		observerKeeper,
		authorityKeeper,
	)
}

// PevmKeeperWithMocks initializes a pevm keeper for testing purposes with option to mock specific keepers
func PevmKeeperWithMocks(t testing.TB, mockOptions PevmMockOptions) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Initialize local store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cdc := NewCodec()
	amino := codec.NewLegacyAmino()
	// Create regular keepers
	sdkKeepers := NewSDKKeepers(cdc, amino, db, stateStore)

	// Create authority keeper
	authorityKeeperTmp := initAuthorityKeeper(
		cdc,
		db,
		stateStore,
	)

	// Create lightclient keeper
	lightclientKeeperTmp := initLightclientKeeper(
		cdc,
		db,
		stateStore,
		authorityKeeperTmp,
	)

	// Create observer keeper
	observerKeeperTmp := initRelayerKeeper(
		cdc,
		db,
		stateStore,
		sdkKeepers.StakingKeeper,
		sdkKeepers.SlashingKeeper,
		sdkKeepers.ParamsKeeper,
		authorityKeeperTmp,
		lightclientKeeperTmp,
	)
	pellKeepers := PellKeepers{
		ObserverKeeper:    observerKeeperTmp,
		AuthorityKeeper:   &authorityKeeperTmp,
		LightclientKeeper: &lightclientKeeperTmp,
	}
	var observerKeeper types.RelayerKeeper = observerKeeperTmp
	var authorityKeeper types.AuthorityKeeper = authorityKeeperTmp

	// Create the pevm keeper
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
	var evmKeeper types.EVMKeeper = sdkKeepers.EvmKeeper

	if mockOptions.UseAccountMock {
		authKeeper = pevmmocks.NewPevmAccountKeeper(t)
	}
	if mockOptions.UseBankMock {
		bankKeeper = pevmmocks.NewPevmBankKeeper(t)
	}
	if mockOptions.UseObserverMock {
		observerKeeper = pevmmocks.NewPevmRelayerKeeper(t)
		fok, ok := observerKeeper.(*pevmmocks.PevmRelayerKeeper)
		if ok {
			fok.On("SetPevmKeeper", mock.Anything).Maybe()
		}
	}
	if mockOptions.UseEVMMock {
		evmKeeper = pevmmocks.NewPevmEVMKeeper(t)
	}
	if mockOptions.UseAuthorityMock {
		authorityKeeper = pevmmocks.NewPevmAuthorityKeeper(t)
	}

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		authKeeper,
		evmKeeper,
		bankKeeper,
		observerKeeper,
		authorityKeeper,
	)

	pevmmodule.InitGenesis(ctx, *k, *types.DefaultGenesis())

	return k, ctx, sdkKeepers, pellKeepers
}

// PevmKeeperAllMocks initializes a pevm keeper for testing purposes with all keeper mocked
func PevmKeeperAllMocks(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx, _, _ := PevmKeeperWithMocks(t, PevmMocksAll)
	return k, ctx
}

// PevmKeeper initializes a pevm keeper for testing purposes
func PevmKeeper(t testing.TB) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	k, ctx, sdkk, zk := PevmKeeperWithMocks(t, PevmNoMocks)
	return k, ctx, sdkk, zk
}

// GetPevmAuthorityMock returns a new pevm authority keeper mock
func GetPevmAuthorityMock(t testing.TB, keeper *keeper.Keeper) *pevmmocks.PevmAuthorityKeeper {
	cok, ok := keeper.GetAuthorityKeeper().(*pevmmocks.PevmAuthorityKeeper)
	require.True(t, ok)
	return cok
}

func GetPevmAccountMock(t testing.TB, keeper *keeper.Keeper) *pevmmocks.PevmAccountKeeper {
	fak, ok := keeper.GetAuthKeeper().(*pevmmocks.PevmAccountKeeper)
	require.True(t, ok)
	return fak
}

func GetPevmBankMock(t testing.TB, keeper *keeper.Keeper) *pevmmocks.PevmBankKeeper {
	fbk, ok := keeper.GetBankKeeper().(*pevmmocks.PevmBankKeeper)
	require.True(t, ok)
	return fbk
}

func GetPevmObserverMock(t testing.TB, keeper *keeper.Keeper) *pevmmocks.PevmRelayerKeeper {
	fok, ok := keeper.GetRelayerKeeper().(*pevmmocks.PevmRelayerKeeper)
	require.True(t, ok)
	return fok
}

func GetPevmEVMMock(t testing.TB, keeper *keeper.Keeper) *PevmMockEVMKeeper {
	fek, ok := keeper.GetEVMKeeper().(*pevmmocks.PevmEVMKeeper)
	require.True(t, ok)
	return &PevmMockEVMKeeper{
		PevmEVMKeeper: fek,
	}
}

type PevmMockEVMKeeper struct {
	*pevmmocks.PevmEVMKeeper
}

func (m *PevmMockEVMKeeper) SetupMockEVMKeeperForSystemContractDeployment() {
	gasRes := &evmtypes.EstimateGasResponse{Gas: 1000}
	m.On("WithChainID", mock.Anything).Maybe().Return(mock.Anything)
	m.On("PellChainID").Maybe().Return(big.NewInt(1))
	m.On(
		"EstimateGas",
		mock.Anything,
		mock.Anything,
	).Return(gasRes, nil)
	m.MockEVMSuccessCallTimes(5)
	m.On(
		"GetAccount",
		mock.Anything,
		mock.Anything,
	).Return(&statedb.Account{
		Nonce: 1,
	})
	m.On(
		"GetCode",
		mock.Anything,
		mock.Anything,
	).Return([]byte{1, 2, 3})
}

func (m *PevmMockEVMKeeper) MockEVMSuccessCallOnce() {
	m.MockEVMSuccessCallOnceWithReturn(&evmtypes.MsgEthereumTxResponse{})
}

func (m *PevmMockEVMKeeper) MockEVMSuccessCallTimes(times int) {
	m.MockEVMSuccessCallTimesWithReturn(&evmtypes.MsgEthereumTxResponse{}, times)
}

func (m *PevmMockEVMKeeper) MockEVMSuccessCallOnceWithReturn(ret *evmtypes.MsgEthereumTxResponse) {
	m.MockEVMSuccessCallTimesWithReturn(ret, 1)
}

func (m *PevmMockEVMKeeper) MockEVMSuccessCallTimesWithReturn(ret *evmtypes.MsgEthereumTxResponse, times int) {
	if ret == nil {
		ret = &evmtypes.MsgEthereumTxResponse{}
	}
	m.On(
		"ApplyMessage",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(ret, nil).Times(times)
}

func (m *PevmMockEVMKeeper) MockEVMFailCallOnce() {
	m.On(
		"ApplyMessage",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&evmtypes.MsgEthereumTxResponse{}, sample.ErrSample).Once()
}
