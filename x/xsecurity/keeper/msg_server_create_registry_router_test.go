package keeper_test

import (
	"fmt"
	"testing"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	"github.com/ethereum/go-ethereum/accounts/abi"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	types "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

var (
	registryRouterFactoryMedaDataABI *abi.ABI
)

// Common constants used across registry router tests
const (
	chainApprover           = "0x1111111111111111111111111111111111111111"
	churnApprover           = "0x2222222222222222222222222222222222222222"
	ejector                 = "0x3333333333333333333333333333333333333333"
	pauser                  = "0x4444444444444444444444444444444444444444"
	unpauser                = "0x5555555555555555555555555555555555555555"
	registryRouterAddr      = "0x6666666666666666666666666666666666666666"
	stakeRegistryRouterAddr = "0x7777777777777777777777777777777777777777"
)

var registryRouterCreatedEventID string

func setupRegistryRouterFactoryMedaDataABI() {
	registryRouterFactoryMedaDataABI, _ = registryrouterfactory.RegistryRouterFactoryMetaData.GetAbi()

	registryRouterCreatedEventID = registryRouterFactoryMedaDataABI.Events["RegistryRouterCreated"].ID.String()
}

// createStandardRegistryRouterMsg creates a standard test message for creating a registry router
func createStandardRegistryRouterMsg(signer string) *types.MsgCreateRegistryRouter {
	return &types.MsgCreateRegistryRouter{
		Signer:              signer,
		ChainApprover:       chainApprover,
		ChurnApprover:       churnApprover,
		Ejector:             ejector,
		Pauser:              pauser,
		Unpauser:            unpauser,
		InitialPausedStatus: 0, // Unpause by default
	}
}

// createRegistryRouterReceipt creates a sample receipt with RegistryRouterCreated event
func createRegistryRouterReceipt() *evmtypes.MsgEthereumTxResponse {
	// Create data with embedded addresses
	// Actual implementation would have proper encoding of addresses
	data := make([]byte, 64)
	// Placeholder data - in a real test this would contain properly encoded addresses
	return &evmtypes.MsgEthereumTxResponse{
		Logs: []*evmtypes.Log{
			{
				Topics: []string{
					registryRouterCreatedEventID,
				},
				Data: data, // Mock data that would contain the embedded addresses
			},
		},
	}
}

// TestCreateRegistryRouterSuccess tests the successful creation of a registry router.
// It verifies that an authorized signer can successfully create a registry router
// when all conditions are met and proper parameters are provided.
func TestCreateRegistryRouterSuccess(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	setupRegistryRouterFactoryMedaDataABI()

	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Prepare receipt and message
	sampleReceipt := createRegistryRouterReceipt()
	msg := createStandardRegistryRouterMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock successful PEVM call
	pevmk.On("CallRegistryRouterFactory",
		mock.Anything,
		mock.Anything, // chainApprover
		mock.Anything, // churnApprover
		mock.Anything, // ejector
		mock.Anything, // pauser
		mock.Anything, // unpauser
		mock.Anything, // initialPausedStatus
	).Return(
		sampleReceipt,
		true,
		nil,
	).Once()

	// Execute
	_, err := mocks.CreateRegistryRouter(ctx, msg)

	// Verify results
	require.NoError(t, err)

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestCreateRegistryRouterUnauthorized tests the unauthorized access scenario.
// It verifies that an unauthorized signer cannot create a registry router and
// receives an appropriate authorization error.
func TestCreateRegistryRouterUnauthorized(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Prepare message with unauthorized signer
	msg := createStandardRegistryRouterMsg(unauthSigner)

	// Setup auth mock for unauthorized case
	ak.On("IsAuthorized",
		mock.Anything,
		unauthSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(false).Once()

	// Execute
	_, err := mocks.CreateRegistryRouter(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestCreateRegistryRouterAlreadyExists tests the scenario where a registry router already exists.
// It verifies that the function fails properly when attempting to create a registry router
// that has already been created.
func TestCreateRegistryRouterAlreadyExists(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	setupRegistryRouterFactoryMedaDataABI()

	// Prepare message
	msg := createStandardRegistryRouterMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	mocks.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      registryRouterAddr,
		StakeRegistryRouterAddress: stakeRegistryRouterAddr,
	})
	// Execute
	_, err := mocks.CreateRegistryRouter(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Equal(t, "registry router already exists", err.Error())

	// Verify mock expectations
	ak.AssertExpectations(t)
}

// TestCreateRegistryRouterPEVMCallFailure tests the handling of PEVM call failures.
// It verifies that the function properly handles errors returned from the PEVM module.
func TestCreateRegistryRouterPEVMCallFailure(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Prepare message
	msg := createStandardRegistryRouterMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock PEVM call failure
	pevmk.On("CallRegistryRouterFactory",
		mock.Anything,
		mock.Anything, // chainApprover
		mock.Anything, // churnApprover
		mock.Anything, // ejector
		mock.Anything, // pauser
		mock.Anything, // unpauser
		mock.Anything, // initialPausedStatus
	).Return(
		nil,
		false,
		fmt.Errorf("PEVM call failed"),
	).Once()

	// Execute
	_, err := mocks.CreateRegistryRouter(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Contains(t, err.Error(), "PEVM call failed")

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}

// TestCreateRegistryRouterEventParsingFailure tests the scenario where parsing the event from the receipt fails.
// It verifies that the function properly handles errors when extracting router addresses from receipt.
func TestCreateRegistryRouterEventParsingFailure(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)
	pevmk := keepertest.GetXSecurityPEVMKeeperMock(t, mocks)

	// Prepare receipt and message
	badReceipt := &evmtypes.MsgEthereumTxResponse{
		Logs: []*evmtypes.Log{
			{
				Topics: []string{
					"bad_event_id",
				},
				Data: []byte{},
			},
		},
	}

	msg := createStandardRegistryRouterMsg(authSigner)

	// Setup auth mock for authorized case
	ak.On("IsAuthorized",
		mock.Anything,
		authSigner,
		authoritytypes.PolicyType_GROUP_OPERATIONAL,
	).Return(true).Once()

	// Mock successful PEVM call
	pevmk.On("CallRegistryRouterFactory",
		mock.Anything,
		mock.Anything, // chainApprover
		mock.Anything, // churnApprover
		mock.Anything, // ejector
		mock.Anything, // pauser
		mock.Anything, // unpauser
		mock.Anything, // initialPausedStatus
	).Return(
		badReceipt,
		false,
		nil,
	).Once()

	// Execute
	_, err := mocks.CreateRegistryRouter(ctx, msg)

	// Verify results
	require.Error(t, err)
	require.Contains(t, err.Error(), "RegistryRouterCreated event not found in receipt")

	// Verify mock expectations
	ak.AssertExpectations(t)
	pevmk.AssertExpectations(t)
}
