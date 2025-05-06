package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	"github.com/0xPellNetwork/aegis/x/xsecurity/keeper"
	types "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// createAddPoolsMsg creates a standard test message for adding pools
func createAddPoolsMsg(signer string) *types.MsgAddPools {
	return &types.MsgAddPools{
		Signer: signer,
		Pools: []*restakingtypes.PoolParams{
			{
				ChainId:    2,
				Pool:       "pool2",
				Multiplier: 2,
			},
			{
				ChainId:    3,
				Pool:       "pool3",
				Multiplier: 3,
			},
		},
	}
}

// createRemovePoolsMsg creates a standard test message for removing pools
func createRemovePoolsMsg(signer string) *types.MsgRemovePools {
	return &types.MsgRemovePools{
		Signer: signer,
		Pools: []*restakingtypes.PoolParams{
			{
				Pool: "pool1",
			},
		},
	}
}

// setupGroupWithPools sets up a group with initial pools for testing
func setupGroupWithPools(mocks *keeper.Keeper, ctx sdk.Context) {
	// Set up group info with initial pools
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber: 1,
		PoolParams: []*restakingtypes.PoolParams{
			{
				ChainId:    1,
				Pool:       "pool1",
				Multiplier: 1,
			},
		},
	})

	// Set up registry router address
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})
}

// TestAddPoolsSuccess tests the successful addition of pools to a group
func TestAddPoolsSuccess(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message
	msg := createAddPoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock successful PEVM call
	pevmk.On("CallStakelRegistryRouterToAddPools",
		mock.Anything,
		mock.Anything, // stake registry router address
		uint64(1),     // group number
		msg.Pools,     // pools to add
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		true,
		nil,
	).Once()

	// Execute
	_, err := mocks.AddPools(ctx, msg)

	// Verify results
	require.NoError(t, err)

	// Verify the group info was updated correctly
	groupInfo, exists := mocks.GetGroupInfo(ctx)
	require.True(t, exists)
	require.Len(t, groupInfo.PoolParams, 3) // Original pool + 2 new pools
	require.Equal(t, "pool1", groupInfo.PoolParams[0].Pool)
	require.Equal(t, "pool2", groupInfo.PoolParams[1].Pool)
	require.Equal(t, "pool3", groupInfo.PoolParams[2].Pool)

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestAddPoolsUnauthorized tests the unauthorized access scenario for adding pools
func TestAddPoolsUnauthorized(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message with unauthorized signer
	msg := createAddPoolsMsg(unauthSigner)

	// Setup auth mock for unauthorized case
	ak.On("IsAuthorized",
		mock.Anything,
		unauthSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(false).Once()

	// Execute
	_, err := mocks.AddPools(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestAddPoolsGroupInfoNotExists tests the scenario where group info doesn't exist
func TestAddPoolsGroupInfoNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Create message
	msg := createAddPoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute without setting up group info
	_, err := mocks.AddPools(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, types.ErrDataEmpty)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestAddPoolsRegistryRouterNotExists tests the scenario where registry router doesn't exist
func TestAddPoolsRegistryRouterNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Setup group info without registry router
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber: 1,
		PoolParams: []*restakingtypes.PoolParams{
			{
				ChainId:    1,
				Pool:       "pool1",
				Multiplier: 1,
			},
		},
	})

	// Create message
	msg := createAddPoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute without registry router
	_, err := mocks.AddPools(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, types.ErrDataEmpty)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestAddPoolsPEVMCallFailure tests the handling of PEVM call failures
func TestAddPoolsPEVMCallFailure(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message
	msg := createAddPoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock PEVM call failure
	pevmk.On("CallStakelRegistryRouterToAddPools",
		mock.Anything,
		mock.Anything, // stake registry router address
		uint64(1),     // group number
		msg.Pools,     // pools to add
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		false,
		fmt.Errorf("PEVM call failed"),
	).Once()

	// Execute
	_, err := mocks.AddPools(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Contains(t, err.Error(), "PEVM call failed")

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestRemovePoolsSuccess tests the successful removal of pools from a group
func TestRemovePoolsSuccess(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message
	msg := createRemovePoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock successful PEVM call
	pevmk.On("CallStakelRegistryRouterToRemovePools",
		mock.Anything,
		mock.Anything,   // stake registry router address
		uint64(1),       // group number
		[]uint{uint(0)}, // indices to remove (pool1 is at index 0)
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		true,
		nil,
	).Once()

	// Execute
	_, err := mocks.RemovePools(ctx, msg)

	// Verify results
	require.NoError(t, err)

	// Verify the group info was updated correctly
	groupInfo, exists := mocks.GetGroupInfo(ctx)
	require.True(t, exists)
	require.Len(t, groupInfo.PoolParams, 0) // All pools should be removed

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestRemovePoolsUnauthorized tests the unauthorized access scenario for removing pools
func TestRemovePoolsUnauthorized(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message with unauthorized signer
	msg := createRemovePoolsMsg(unauthSigner)

	// Setup auth mock for unauthorized case
	ak.On("IsAuthorized",
		mock.Anything,
		unauthSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(false).Once()

	// Execute
	_, err := mocks.RemovePools(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestRemovePoolsGroupInfoNotExists tests the scenario where group info doesn't exist
func TestRemovePoolsGroupInfoNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Create message
	msg := createRemovePoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute without setting up group info
	_, err := mocks.RemovePools(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, types.ErrDataEmpty)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestRemovePoolsRegistryRouterNotExists tests the scenario where registry router doesn't exist
func TestRemovePoolsRegistryRouterNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Setup group info without registry router
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber: 1,
		PoolParams: []*restakingtypes.PoolParams{
			{
				ChainId:    1,
				Pool:       "pool1",
				Multiplier: 1,
			},
		},
	})

	// Create message
	msg := createRemovePoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute without registry router
	_, err := mocks.RemovePools(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, types.ErrDataEmpty)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestRemovePoolsPEVMCallFailure tests the handling of PEVM call failures
func TestRemovePoolsPEVMCallFailure(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message
	msg := createRemovePoolsMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock PEVM call failure
	pevmk.On("CallStakelRegistryRouterToRemovePools",
		mock.Anything,
		mock.Anything,   // stake registry router address
		uint64(1),       // group number
		[]uint{uint(0)}, // indices to remove (pool1 is at index 0)
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		false,
		fmt.Errorf("PEVM call failed"),
	).Once()

	// Execute
	_, err := mocks.RemovePools(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Contains(t, err.Error(), "PEVM call failed")

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestRemovePoolsNonExistentPool tests handling when trying to remove a pool that doesn't exist
func TestRemovePoolsNonExistentPool(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	setupGroupWithPools(mocks, ctx)

	// Create message with non-existent pool
	msg := &types.MsgRemovePools{
		Signer: authSigner,
		Pools: []*restakingtypes.PoolParams{
			{
				Pool: "non_existent_pool",
			},
		},
	}

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock PEVM call - with nil instead of empty slice
	pevmk.On("CallStakelRegistryRouterToRemovePools",
		mock.Anything,
		mock.Anything, // stake registry router address
		uint64(1),     // group number
		mock.Anything, // using mock.Anything to match any slice (either nil or []uint{})
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		true,
		nil,
	).Once()

	// Execute
	_, err := mocks.RemovePools(ctx, msg)

	// Verify results
	require.NoError(t, err)

	// Group info should remain unchanged
	groupInfo, exists := mocks.GetGroupInfo(ctx)
	require.True(t, exists)
	require.Len(t, groupInfo.PoolParams, 1) // Original pool should still be there
	require.Equal(t, "pool1", groupInfo.PoolParams[0].Pool)

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}
