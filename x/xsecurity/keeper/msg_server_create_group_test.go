package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	types "github.com/pell-chain/pellcore/x/xsecurity/types"
)

// Common constants used across tests
const (
	authSigner       = "auth_signer"
	unauthSigner     = "unauth_signer"
	routerAddress    = "0x1234567890123456789012345678901234567890"
	stakingAddress   = "0x1234567890123456789012345678901234567891"
	syncGroupEventID = "0xd504611e9ae12b362445db90e7f296d92e9fe36b3fdb6d97069ac1162300a19f"
	zeroInHex        = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

// createStandardTestMsg creates a standard test message with given signer
func createStandardTestMsg(signer string) *types.MsgCreateGroup {
	return &types.MsgCreateGroup{
		Signer: signer,
		OperatorSetParams: &restakingtypes.OperatorSetParam{
			MaxOperatorCount:        10,
			KickBipsOfOperatorStake: 100,
			KickBipsOfTotalStake:    100,
		},
		MinStake: sdkmath.NewInt(1000),
		PoolParams: []*restakingtypes.PoolParams{
			{
				ChainId:    1,
				Pool:       "pool1",
				Multiplier: 1,
			},
		},
		GroupEjectionParams: &restakingtypes.GroupEjectionParam{
			RateLimitWindow:       100,
			EjectableStakePercent: 10,
		},
	}
}

// createSampleReceipt creates a sample receipt with SyncCreateGroup event
func createSampleReceipt() *evmtypes.MsgEthereumTxResponse {
	return &evmtypes.MsgEthereumTxResponse{
		Logs: []*evmtypes.Log{
			{
				Topics: []string{
					syncGroupEventID,
					zeroInHex, // Group number 0 in hex
				},
			},
		},
	}
}

// TestCreateGroupSuccess tests the successful creation of a new group.
// It verifies that an authorized signer can successfully create a new group
// when all conditions are met and proper parameters are provided.
func TestCreateGroupSuccess(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Set up registry router address
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})

	// Prepare receipt and message
	sampleReceipt := createSampleReceipt()
	msg := createStandardTestMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock successful PEVM call
	pevmk.On("CallRegistryRouterToCreateGroup",
		mock.Anything,
		mock.Anything, // registry router address
		mock.Anything, // operator set params
		mock.Anything, // min stake
		mock.Anything, // pool params
		mock.Anything, // group ejection params
	).Return(
		sampleReceipt,
		true, // No return data needed for this test
		nil,
	).Once()

	// Execute
	_, err := mocks.CreateGroup(ctx, msg)

	// Verify results
	require.NoError(t, err)

	// Verify that group info was set correctly
	groupInfo, exists := mocks.GetGroupInfo(ctx)
	require.True(t, exists)
	require.Equal(t, uint64(0), groupInfo.GroupNumber)

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestCreateGroupUnauthorized tests the unauthorized access scenario.
// It verifies that an unauthorized signer cannot create a group and
// receives an appropriate authorization error.
func TestCreateGroupUnauthorized(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Set up registry router address
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})

	// Prepare message with unauthorized signer
	msg := createStandardTestMsg(unauthSigner)

	// Setup auth mock for unauthorized case
	ak.On("IsAuthorized",
		mock.Anything,
		unauthSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(false).Once()

	// Execute
	_, err := mocks.CreateGroup(ctx, msg)

	// Verify results
	require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestCreateGroupRegistryRouterNotExists tests the scenario where registry router is not configured.
// It verifies that the function fails properly when the registry router address is not set up.
func TestCreateGroupRegistryRouterNotExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Prepare message
	msg := createStandardTestMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute (without setting registry router address)
	_, err := mocks.CreateGroup(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Equal(t, "registry router not exists", err.Error())

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestCreateGroupAlreadyExists tests the scenario where a group already exists.
// It verifies that the function fails properly when attempting to create
// a group that has already been created.
func TestCreateGroupAlreadyExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Set up registry router address
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})

	// Set up existing group info
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber:  0,
		MinimumStake: sdkmath.NewInt(1000),
	})

	// Prepare message
	msg := createStandardTestMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Execute
	_, err := mocks.CreateGroup(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Equal(t, "group already exists", err.Error())

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestCreateGroupPEVMCallFailure tests the handling of PEVM call failures.
// It verifies that the function properly handles errors returned from the PEVM module.
func TestCreateGroupPEVMCallFailure(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Set up registry router address
	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      routerAddress,
		StakeRegistryRouterAddress: stakingAddress,
	})

	// Prepare message
	msg := createStandardTestMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock PEVM call failure
	pevmk.On("CallRegistryRouterToCreateGroup",
		mock.Anything,
		mock.Anything, // registry router address
		mock.Anything, // operator set params
		mock.Anything, // min stake
		mock.Anything, // pool params
		mock.Anything, // group ejection params
	).Return(
		&evmtypes.MsgEthereumTxResponse{},
		false,
		fmt.Errorf("PEVM call failed"),
	).Once()

	// Execute
	_, err := mocks.CreateGroup(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Contains(t, err.Error(), "PEVM call failed")

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}
