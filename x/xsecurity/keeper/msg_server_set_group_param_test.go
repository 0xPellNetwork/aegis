package keeper_test

import (
	"fmt"
	"testing"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	types "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// createSetGroupParamMsg creates a standard test message for setting group parameters
func createSetGroupParamMsg(signer string) *types.MsgSetGroupParam {
	return &types.MsgSetGroupParam{
		Signer: signer,
		OperatorSetParams: &restakingtypes.OperatorSetParam{
			MaxOperatorCount:        20,
			KickBipsOfOperatorStake: 200,
			KickBipsOfTotalStake:    200,
		},
	}
}

// TestSetGroupParamSuccess tests the successful update of group parameters
func TestSetGroupParamSuccess(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	// Set up registry router address
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})

	// Set up existing group info with initial parameters
	initialParam := &restakingtypes.OperatorSetParam{
		MaxOperatorCount:        10,
		KickBipsOfOperatorStake: 100,
		KickBipsOfTotalStake:    100,
	}
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber:      1,
		OperatorSetParam: initialParam,
	})

	// Create message
	msg := createSetGroupParamMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock successful PEVM call
	pevmk.On("CallRegistryRouterToSetOperatorSetParams",
		mock.Anything,
		mock.Anything, // registry router address
		uint64(1),     // group number
		msg.OperatorSetParams,
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		true,
		nil,
	).Once()

	// Execute
	_, err := mocks.SetGroupParam(ctx, msg)

	// Verify results
	require.NoError(t, err)

	// Verify the group info was updated correctly
	groupInfo, exists := mocks.GetGroupInfo(ctx)
	require.True(t, exists)
	require.Equal(t, uint64(1), groupInfo.GroupNumber)
	require.Equal(t, uint32(20), groupInfo.OperatorSetParam.MaxOperatorCount)
	require.Equal(t, uint32(200), groupInfo.OperatorSetParam.KickBipsOfOperatorStake)
	require.Equal(t, uint32(200), groupInfo.OperatorSetParam.KickBipsOfTotalStake)

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestSetGroupParamUnauthorized tests the unauthorized access scenario
func TestSetGroupParamUnauthorized(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Setup test environment
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{GroupNumber: 1})

	// Create message with unauthorized signer
	msg := createSetGroupParamMsg(unauthSigner)

	// Setup auth mock for unauthorized case
	ak.On("IsAuthorized",
		mock.Anything,
		unauthSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(false).Once()

	// Execute
	_, err := mocks.SetGroupParam(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestSetGroupParamGroupInfoNotExists tests the scenario where group info doesn't exist
func TestSetGroupParamGroupInfoNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Create message
	msg := createSetGroupParamMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute without setting up group info
	_, err := mocks.SetGroupParam(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, types.ErrDataEmpty)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestSetGroupParamRegistryRouterNotExists tests the scenario where registry router doesn't exist
func TestSetGroupParamRegistryRouterNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Setup group info without registry router
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{GroupNumber: 1})

	// Create message
	msg := createSetGroupParamMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute without registry router
	_, err := mocks.SetGroupParam(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, types.ErrDataEmpty)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestSetGroupParamPEVMCallFailure tests the handling of PEVM call failures
func TestSetGroupParamPEVMCallFailure(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Setup test environment
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{GroupNumber: 1})

	// Create message
	msg := createSetGroupParamMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock PEVM call failure
	pevmk.On("CallRegistryRouterToSetOperatorSetParams",
		mock.Anything,
		mock.Anything, // registry router address
		uint64(1),     // group number
		msg.OperatorSetParams,
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		false,
		fmt.Errorf("PEVM call failed"),
	).Once()

	// Execute
	_, err := mocks.SetGroupParam(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Contains(t, err.Error(), "PEVM call failed")

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}
