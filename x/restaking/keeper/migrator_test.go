package keeper_test

import (
	"encoding/binary"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// TestMigration1_1to1_2 simulates pre-migration V1 state data, executes the migration, and verifies the V2 state data.
func TestMigration1_1to1_2(t *testing.T) {
	// Initialize the keeper and context with mocked dependencies.
	k, ctx := keepertest.RestakingKeeperWithMocks(
		t,
		keepertest.RestakingMockOptions{
			UseBankMock:      true,
			UseAccountMock:   true,
			UseAuthorityMock: true,
			UseEvmMock:       true,
			UseRelayerMock:   true,
			UsePevmMock:      true,
		},
	)

	// Generate a random registry router address.
	registryRouterAddr := GenerateRandomAddress()

	// ----------------------------------------------------
	// Simulate pre-migration V1 data.
	// ----------------------------------------------------

	// Create the old version GroupOperatorRegistration data.
	oldRegistration := &types.GroupOperatorRegistration{
		Operator:     "0x123456",
		OperatorId:   []byte{1, 2},
		GroupNumbers: []byte{1, 2},
		Socket:       "socket-1",
		PubkeyParams: &types.PubkeyRegistrationParams{
			PubkeyG1: &types.G1Point{
				X: 12345,
				Y: 67890,
			},
			PubkeyG2: &types.G2Point{
				X: []uint64{111, 222},
				Y: []uint64{333, 444},
			},
		},
	}
	err := k.AddGroupOperatorRegistrationV1(ctx, registryRouterAddr, oldRegistration)
	require.NoError(t, err)

	// Add DVS supported chain data (to test for nil pointer issues).
	dvsInfo := &types.DVSInfo{
		ChainId:          GenerateRandomUint64(),
		ServiceManager:   GenerateRandomAddress().String(),
		EjectionManager:  GenerateRandomAddress().String(),
		CentralScheduler: GenerateRandomAddress().String(),
		StakeManager:     GenerateRandomAddress().String(),
		BlsApkRegistry:   GenerateRandomAddress().String(),
		IndexRegistry:    GenerateRandomAddress().String(),
		OutboundState:    0,
	}
	err = k.AddDVSSupportedChain(ctx, registryRouterAddr, dvsInfo)
	require.NoError(t, err)

	// Add group data (to test for nil pointer issues).
	groupData := &types.Group{
		GroupNumber:        0,
		OperatorSetParam:   nil,
		MinimumStake:       0,
		PoolParams:         nil,
		GroupEjectionParam: nil,
	}
	err = k.AddGroupData(ctx, registryRouterAddr, groupData)
	require.NoError(t, err)

	// Add the registry router address (to test for nil pointer issues).
	err = k.AddRegistryRouterAddress(ctx, []common.Address{registryRouterAddr})
	require.NoError(t, err)

	// Retrieve and log the V1 data to ensure it is stored correctly.
	v1Data, found := k.GetGroupOperatorRegistrationListV1(ctx, registryRouterAddr)
	require.True(t, found)
	t.Log("V1 data: ", v1Data)

	// ----------------------------------------------------
	// Execute the migration.
	// ----------------------------------------------------
	err = k.MigrationStore(ctx)
	require.NoError(t, err)

	// Retrieve the migrated V2 data.
	v2Data, found := k.GetGroupOperatorRegistrationList(ctx, registryRouterAddr)
	require.True(t, found)
	t.Log("V2 data: ", v2Data)

	// Verify the number of operator registered information entries.
	require.Len(t, v2Data.OperatorRegisteredInfos, 1)

	// ----------------------------------------------------
	// Verify the migrated data.
	// ----------------------------------------------------
	migratedReg := v2Data.OperatorRegisteredInfos[0]
	require.Equal(t, "0x123456", migratedReg.Operator)
	require.Equal(t, []byte{1, 2}, migratedReg.OperatorId)
	require.Equal(t, []byte{1, 2}, migratedReg.GroupNumbers)
	require.Equal(t, "socket-1", migratedReg.Socket)

	// Verify PubkeyG1 data.
	require.Equal(t, sdkmath.NewIntFromUint64(12345), migratedReg.PubkeyParams.PubkeyG1.X)
	require.Equal(t, sdkmath.NewIntFromUint64(67890), migratedReg.PubkeyParams.PubkeyG1.Y)

	// Verify PubkeyG2 data.
	expectedG2X := []sdkmath.Int{sdkmath.NewIntFromUint64(111), sdkmath.NewIntFromUint64(222)}
	expectedG2Y := []sdkmath.Int{sdkmath.NewIntFromUint64(333), sdkmath.NewIntFromUint64(444)}
	require.Equal(t, expectedG2X, migratedReg.PubkeyParams.PubkeyG2.X)
	require.Equal(t, expectedG2Y, migratedReg.PubkeyParams.PubkeyG2.Y)
}

// GenerateRandomAddress generates a random Ethereum address.
func GenerateRandomAddress() common.Address {
	var address common.Address
	if _, err := rand.Read(address[:]); err != nil {
		panic("failed to generate random address")
	}
	return address
}

// GenerateRandomUint64 generates a random uint64 value.
func GenerateRandomUint64() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("failed to generate random uint64")
	}
	return binary.LittleEndian.Uint64(b[:])
}
