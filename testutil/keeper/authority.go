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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/authority/keeper"
	"github.com/0xPellNetwork/aegis/x/authority/types"
)

var (
	AuthorityGovAddress = sample.Bech32AccAddress()
)

func initAuthorityKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
) keeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ss.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, db)

	return keeper.NewKeeper(
		cdc,
		storeKey,
		memKey,
		AuthorityGovAddress,
	)
}

// AuthorityKeeper instantiates an authority keeper for testing purposes
func AuthorityKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Initialize local store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cdc := NewCodec()
	amino := codec.NewLegacyAmino()
	// Create regular keepers
	sdkKeepers := NewSDKKeepers(cdc, amino, db, stateStore)

	// Create the observer keeper
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := NewContext(stateStore)

	// Initialize modules genesis
	sdkKeepers.InitGenesis(ctx)

	// Add a proposer to the context
	ctx = sdkKeepers.InitBlockProposer(t, ctx)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		AuthorityGovAddress,
	)

	return &k, ctx
}

// MockIsAuthorized mocks the IsAuthorized method of an authority keeper mock
func MockIsAuthorized(m *mock.Mock, address string, policyType types.PolicyType, isAuthorized bool) {
	m.On("IsAuthorized", mock.Anything, address, policyType).Return(isAuthorized).Once()
}

func SetAdminPolices(ctx sdk.Context, ak *keeper.Keeper) string {
	admin := sample.AccAddress()
	ak.SetPolicies(ctx, types.Policies{Items: []*types.Policy{
		{
			Address:    admin,
			PolicyType: types.PolicyType_GROUP_ADMIN,
		},
	}})
	return admin
}
