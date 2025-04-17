package restaking_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/nullify"
	"github.com/pell-chain/pellcore/x/restaking"
	"github.com/pell-chain/pellcore/x/restaking/types"
)

// TestGenesis verifies that the restaking module can correctly import genesis state
// and then export it back with all data preserved.
func TestGenesis(t *testing.T) {
	// ---------- Prepare test data ----------

	// Create operator shares data
	operatorShares := []types.OperatorShares{
		{
			ChainId:  1,
			Operator: "0x1234567890123456789012345678901234567890",
			Strategy: "strategy1",
			Shares:   math.NewInt(1000),
		},
		{
			ChainId:  2,
			Operator: "0x0987654321098765432109876543210987654321",
			Strategy: "strategy2",
			Shares:   math.NewInt(2000),
		},
	}

	// Create registry router data with DVS info, groups, and operator registrations
	registryRouterData := []types.RegistryRouterData{
		{
			// Define registry router addresses
			RegistryRouterSet: types.RegistryRouterSet{
				RegistryRouterAddress:      "0xabcdef1234567890abcdef1234567890abcdef12",
				StakeRegistryRouterAddress: "0x1234abcdef1234567890abcdef1234567890abcd",
			},

			// Define DVS supported chains
			DvsInfoList: types.DVSInfoList{
				DvsInfos: []*types.DVSInfo{
					{
						// First chain configuration
						ChainId:          1,
						ServiceManager:   "0xaaaa567890123456789012345678901234567890",
						EjectionManager:  "0xbbbb567890123456789012345678901234567890",
						CentralScheduler: "0xcccc567890123456789012345678901234567890",
						StakeManager:     "0xdddd567890123456789012345678901234567890",
						BlsApkRegistry:   "0xeeee567890123456789012345678901234567890",
						IndexRegistry:    "0xffff567890123456789012345678901234567890",
						OutboundState:    types.OutboundStatus_OUTBOUND_STATUS_NORMAL,
					},
					{
						// Second chain configuration
						ChainId:          2,
						ServiceManager:   "0x1111567890123456789012345678901234567890",
						EjectionManager:  "0x2222567890123456789012345678901234567890",
						CentralScheduler: "0x3333567890123456789012345678901234567890",
						StakeManager:     "0x4444567890123456789012345678901234567890",
						BlsApkRegistry:   "0x5555567890123456789012345678901234567890",
						IndexRegistry:    "0x6666567890123456789012345678901234567890",
						OutboundState:    types.OutboundStatus_OUTBOUND_STATUS_SYNCING,
					},
				},
			},

			// Define groups for operators
			GroupList: types.GroupList{
				Groups: []*types.Group{
					{
						// First group configuration
						GroupNumber: 1,
						OperatorSetParam: &types.OperatorSetParam{
							MaxOperatorCount:        10,
							KickBipsOfOperatorStake: 1000,
							KickBipsOfTotalStake:    500,
						},
						MinimumStake: 1000000,
						PoolParams: []*types.PoolParams{
							{
								ChainId:    1,
								Pool:       "0x7777567890123456789012345678901234567890",
								Multiplier: 200,
							},
						},
						GroupEjectionParam: &types.GroupEjectionParam{
							RateLimitWindow:       3600,
							EjectableStakePercent: 20,
						},
					},
					{
						// Second group configuration
						GroupNumber: 2,
						OperatorSetParam: &types.OperatorSetParam{
							MaxOperatorCount:        20,
							KickBipsOfOperatorStake: 2000,
							KickBipsOfTotalStake:    1000,
						},
						MinimumStake: 2000000,
						PoolParams: []*types.PoolParams{
							{
								ChainId:    2,
								Pool:       "0x8888567890123456789012345678901234567890",
								Multiplier: 300,
							},
						},
						GroupEjectionParam: &types.GroupEjectionParam{
							RateLimitWindow:       7200,
							EjectableStakePercent: 30,
						},
					},
				},
			},

			// Define operator registrations with their BLS keys
			GroupOperatorRegistrationList: types.GroupOperatorRegistrationListV2{
				OperatorRegisteredInfos: []*types.GroupOperatorRegistrationV2{
					{
						// First operator registration
						Operator:     "0x1234567890123456789012345678901234567890",
						OperatorId:   []byte{0x01, 0x02, 0x03, 0x04},
						GroupNumbers: []byte{0x01},
						Socket:       "127.0.0.1:8000",
						PubkeyParams: &types.PubkeyRegistrationParamsV2{
							// BLS public key components
							PubkeyG1: &types.G1PointV2{
								X: math.NewInt(123456),
								Y: math.NewInt(654321),
							},
							PubkeyG2: &types.G2PointV2{
								X: []math.Int{math.NewInt(111111), math.NewInt(222222)},
								Y: []math.Int{math.NewInt(333333), math.NewInt(444444)},
							},
						},
					},
					{
						// Second operator registration
						Operator:     "0x0987654321098765432109876543210987654321",
						OperatorId:   []byte{0x05, 0x06, 0x07, 0x08},
						GroupNumbers: []byte{0x02},
						Socket:       "127.0.0.1:9000",
						PubkeyParams: &types.PubkeyRegistrationParamsV2{
							// BLS public key components
							PubkeyG1: &types.G1PointV2{
								X: math.NewInt(789012),
								Y: math.NewInt(210987),
							},
							PubkeyG2: &types.G2PointV2{
								X: []math.Int{math.NewInt(555555), math.NewInt(666666)},
								Y: []math.Int{math.NewInt(777777), math.NewInt(888888)},
							},
						},
					},
				},
			},
		},
	}

	// Combine data into genesis state
	genesisState := types.GenesisState{
		OperatorShare:      operatorShares,
		RegistryRouterData: registryRouterData,
	}

	// ---------- Set up test environment ----------

	// Create keeper with all dependencies mocked
	k, ctx := keepertest.RestakingKeeperAllMocks(t)

	// Mock EVM keeper behavior to allow contract interactions
	evmMock := keepertest.GetRestakingEvmMock(t, k)
	evmMock.On("GetAccount", ctx, mock.Anything).Return(nil).Maybe()

	// ---------- Execute test ----------

	// Import genesis state into the module
	restaking.InitGenesis(ctx, *k, genesisState)

	// Export genesis state from the module
	got := restaking.ExportGenesis(ctx, *k)
	require.NotNil(t, got, "Exported genesis state should not be nil")

	// ---------- Verify results ----------

	// Fill nil fields to ensure proper equality comparison
	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// Compare top-level fields
	require.Equal(t, genesisState.OperatorShare, got.OperatorShare,
		"Operator shares should be preserved after import/export")
	require.Equal(t, len(genesisState.RegistryRouterData), len(got.RegistryRouterData),
		"Number of registry router data entries should match")

	// Detailed comparison of all nested data structures
	for i, routerData := range genesisState.RegistryRouterData {
		// Compare DVSInfo list
		require.Equal(t, len(routerData.DvsInfoList.DvsInfos), len(got.RegistryRouterData[i].DvsInfoList.DvsInfos),
			"Number of DVS infos should match")
		for j, dvsInfo := range routerData.DvsInfoList.DvsInfos {
			require.Equal(t, dvsInfo.ChainId, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].ChainId)
			require.Equal(t, dvsInfo.ServiceManager, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].ServiceManager)
			require.Equal(t, dvsInfo.EjectionManager, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].EjectionManager)
			require.Equal(t, dvsInfo.CentralScheduler, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].CentralScheduler)
			require.Equal(t, dvsInfo.StakeManager, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].StakeManager)
			require.Equal(t, dvsInfo.BlsApkRegistry, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].BlsApkRegistry)
			require.Equal(t, dvsInfo.IndexRegistry, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].IndexRegistry)
			require.Equal(t, dvsInfo.OutboundState, got.RegistryRouterData[i].DvsInfoList.DvsInfos[j].OutboundState)
		}

		// Compare Group list
		require.Equal(t, len(routerData.GroupList.Groups), len(got.RegistryRouterData[i].GroupList.Groups),
			"Number of groups should match")
		for j, group := range routerData.GroupList.Groups {
			require.Equal(t, group.GroupNumber, got.RegistryRouterData[i].GroupList.Groups[j].GroupNumber)
			require.Equal(t, group.MinimumStake, got.RegistryRouterData[i].GroupList.Groups[j].MinimumStake)

			// Compare OperatorSetParam
			if group.OperatorSetParam != nil {
				require.Equal(t, group.OperatorSetParam.MaxOperatorCount,
					got.RegistryRouterData[i].GroupList.Groups[j].OperatorSetParam.MaxOperatorCount)
				require.Equal(t, group.OperatorSetParam.KickBipsOfOperatorStake,
					got.RegistryRouterData[i].GroupList.Groups[j].OperatorSetParam.KickBipsOfOperatorStake)
				require.Equal(t, group.OperatorSetParam.KickBipsOfTotalStake,
					got.RegistryRouterData[i].GroupList.Groups[j].OperatorSetParam.KickBipsOfTotalStake)
			}

			// Compare PoolParams
			require.Equal(t, len(group.PoolParams), len(got.RegistryRouterData[i].GroupList.Groups[j].PoolParams),
				"Number of pool params should match")
			for k, poolParam := range group.PoolParams {
				require.Equal(t, poolParam.ChainId, got.RegistryRouterData[i].GroupList.Groups[j].PoolParams[k].ChainId)
				require.Equal(t, poolParam.Pool, got.RegistryRouterData[i].GroupList.Groups[j].PoolParams[k].Pool)
				require.Equal(t, poolParam.Multiplier, got.RegistryRouterData[i].GroupList.Groups[j].PoolParams[k].Multiplier)
			}

			// Compare GroupEjectionParam
			if group.GroupEjectionParam != nil {
				require.Equal(t, group.GroupEjectionParam.RateLimitWindow,
					got.RegistryRouterData[i].GroupList.Groups[j].GroupEjectionParam.RateLimitWindow)
				require.Equal(t, group.GroupEjectionParam.EjectableStakePercent,
					got.RegistryRouterData[i].GroupList.Groups[j].GroupEjectionParam.EjectableStakePercent)
			}
		}

		// Compare GroupOperatorRegistration list
		require.Equal(t, len(routerData.GroupOperatorRegistrationList.OperatorRegisteredInfos),
			len(got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos),
			"Number of operator registrations should match")
		for j, info := range routerData.GroupOperatorRegistrationList.OperatorRegisteredInfos {
			require.Equal(t, info.Operator, got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].Operator)
			require.Equal(t, info.OperatorId, got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].OperatorId)
			require.Equal(t, info.GroupNumbers, got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].GroupNumbers)
			require.Equal(t, info.Socket, got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].Socket)

			// Compare PubkeyParams
			if info.PubkeyParams != nil {
				// Compare G1Point coordinates
				if info.PubkeyParams.PubkeyG1 != nil {
					require.Equal(t, info.PubkeyParams.PubkeyG1.X.String(),
						got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].PubkeyParams.PubkeyG1.X.String(),
						"G1 X-coordinate should match")
					require.Equal(t, info.PubkeyParams.PubkeyG1.Y.String(),
						got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].PubkeyParams.PubkeyG1.Y.String(),
						"G1 Y-coordinate should match")
				}

				// Compare G2Point coordinates
				if info.PubkeyParams.PubkeyG2 != nil {
					// X coordinate array comparison
					require.Equal(t, len(info.PubkeyParams.PubkeyG2.X),
						len(got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].PubkeyParams.PubkeyG2.X),
						"G2 X-coordinate array length should match")
					for k, x := range info.PubkeyParams.PubkeyG2.X {
						require.Equal(t, x.String(),
							got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].PubkeyParams.PubkeyG2.X[k].String(),
							"G2 X-coordinate elements should match")
					}

					// Y coordinate array comparison
					require.Equal(t, len(info.PubkeyParams.PubkeyG2.Y),
						len(got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].PubkeyParams.PubkeyG2.Y),
						"G2 Y-coordinate array length should match")
					for k, y := range info.PubkeyParams.PubkeyG2.Y {
						require.Equal(t, y.String(),
							got.RegistryRouterData[i].GroupOperatorRegistrationList.OperatorRegisteredInfos[j].PubkeyParams.PubkeyG2.Y[k].String(),
							"G2 Y-coordinate elements should match")
					}
				}
			}
		}
	}
}
