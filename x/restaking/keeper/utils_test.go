package keeper_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/dvsdirectory.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	testkeeper "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/x/pevm/keeper"
	pevmkeeper "github.com/0xPellNetwork/aegis/x/pevm/keeper"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// get a valid chain id independently of the build flag
func getValidChainID(t *testing.T) int64 {
	list := chains.ChainsList()
	require.NotEmpty(t, list)
	require.NotNil(t, list[0])

	return list[0].Id
}

// require that a contract has been deployed by checking stored code is non-empty.
func assertContractDeployment(t *testing.T, k types.EVMKeeper, ctx sdk.Context, contractAddress common.Address) {
	acc := k.GetAccount(ctx, contractAddress)
	require.NotNil(t, acc)

	code := k.GetCode(ctx, common.BytesToHash(acc.CodeHash))
	require.NotEmpty(t, code)
}

func deploySystemContractsWithMockEvmKeeper(
	t *testing.T,
	ctx sdk.Context,
	k *pevmkeeper.Keeper,
	mockEVMKeeper *testkeeper.PevmMockEVMKeeper,
) (systemContract,
	connector,
	proxyAdmin,
	strategyManagerProxy,
	delegationManagerInteractorProxy,
	delegationManagerProxy,
	slasherProxy,
	dvsDirectoryProxy,
	registryRouter common.Address) {
	mockEVMKeeper.SetupMockEVMKeeperForSystemContractDeployment()
	return deploySystemContracts(t, ctx, k, mockEVMKeeper)
}

// deploySystemContracts deploys the system contracts and returns their addresses.
func deploySystemContracts(
	t *testing.T,
	ctx sdk.Context,
	k *pevmkeeper.Keeper,
	evmk types.EVMKeeper,
) (systemContract,
	connector,
	proxyAdmin,
	strategyManagerProxy,
	delegationManagerInteractorProxy,
	delegationManagerProxy,
	slasherProxy,
	dvsDirectoryProxy,
	registryRouter common.Address,
) {
	var err error

	systemContract, err = k.DeployPellSystemContract(ctx, types.ModuleAddressEVM)
	require.NoError(t, err)
	require.NotEmpty(t, systemContract)
	assertContractDeployment(t, evmk, ctx, systemContract)

	emptyContract, err := k.DeployPellEmptyContract(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, emptyContract)
	assertContractDeployment(t, evmk, ctx, emptyContract)

	proxyAdmin, err = k.DeployPellProxyAdmin(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, proxyAdmin)
	assertContractDeployment(t, evmk, ctx, proxyAdmin)

	connector, err = k.DeployPellConnector(ctx, systemContract, types.ModuleAddressEVM)
	require.NoError(t, err)
	require.NotEmpty(t, connector)
	assertContractDeployment(t, evmk, ctx, connector)

	strategyManagerProxy, err = k.DeployPellStrategyManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, strategyManagerProxy)
	assertContractDeployment(t, evmk, ctx, strategyManagerProxy)

	delegationManagerProxy, err = k.DeployPellDelegationManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, delegationManagerProxy)
	assertContractDeployment(t, evmk, ctx, delegationManagerProxy)

	slasherProxy, err = k.DeployPellSlasherProxy(ctx, emptyContract, proxyAdmin, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, slasherProxy)
	assertContractDeployment(t, evmk, ctx, slasherProxy)

	dvsDirectoryImpl, err := k.DeployPellDvsDirectory(ctx, delegationManagerProxy)
	require.NoError(t, err)
	require.NotEmpty(t, dvsDirectoryImpl)
	assertContractDeployment(t, evmk, ctx, dvsDirectoryImpl)

	dvsAbi, err := dvsdirectory.DVSDirectoryMetaData.GetAbi()
	require.NoError(t, err)
	require.NotEmpty(t, dvsAbi)

	data, err := dvsAbi.Pack("initialize", types.ModuleAddressEVM, []common.Address{types.ModuleAddressEVM}, types.ModuleAddressEVM, big.NewInt(0))
	require.NoError(t, err)
	require.NotEmpty(t, data)

	dvsDirectoryProxy, err = k.DeployPellDvsDirectoryProxy(ctx, dvsDirectoryImpl, proxyAdmin, data)
	require.NoError(t, err)
	require.NotEmpty(t, dvsDirectoryProxy)
	assertContractDeployment(t, evmk, ctx, dvsDirectoryProxy)

	registryRouter, err = k.DeployPellRegistryRouter(ctx, dvsDirectoryProxy, systemContract)
	require.NoError(t, err)
	require.NotEmpty(t, registryRouter)
	assertContractDeployment(t, evmk, ctx, registryRouter)

	return
}

type SystemContractDeployConfig struct {
	DeployConnector              bool
	DeployStrategyManagerProxy   bool
	DeployDelegationManagerProxy bool
	DeploySlasherProxy           bool
	DeployDvsDirectoryProxy      bool
}

// deploySystemContractsConfigurable deploys the system contracts and returns their addresses
// while having a possibility to skip some deployments to test different scenarios
func deploySystemContractsConfigurable(
	t *testing.T,
	ctx sdk.Context,
	k *pevmkeeper.Keeper,
	evmk types.EVMKeeper,
	config *SystemContractDeployConfig,
) (systemContract,
	connector,
	proxyAdmin,
	strategyManagerProxy,
	delegationManagerInteractorProxy,
	delegationManagerProxy,
	slasherProxy,
	dvsDirectoryProxy,
	registryRouter common.Address) {
	var err error

	systemContract, err = k.DeployPellSystemContract(ctx, types.ModuleAddressEVM)
	require.NoError(t, err)
	require.NotEmpty(t, systemContract)
	assertContractDeployment(t, evmk, ctx, systemContract)

	emptyContract, err := k.DeployPellEmptyContract(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, emptyContract)
	assertContractDeployment(t, evmk, ctx, emptyContract)

	proxyAdmin, err = k.DeployPellProxyAdmin(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, proxyAdmin)
	assertContractDeployment(t, evmk, ctx, proxyAdmin)

	if config.DeployConnector {
		connector, err = k.DeployPellConnector(ctx, systemContract, types.ModuleAddressEVM)
		require.NoError(t, err)
		require.NotEmpty(t, connector)
		assertContractDeployment(t, evmk, ctx, connector)
	}

	if config.DeployStrategyManagerProxy {
		strategyManagerProxy, err = k.DeployPellStrategyManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
		require.NoError(t, err)
		require.NotEmpty(t, strategyManagerProxy)
		assertContractDeployment(t, evmk, ctx, strategyManagerProxy)
	}

	if config.DeployDelegationManagerProxy {
		delegationManagerProxy, err = k.DeployPellDelegationManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
		require.NoError(t, err)
		require.NotEmpty(t, delegationManagerProxy)
		assertContractDeployment(t, evmk, ctx, delegationManagerProxy)
	}

	if config.DeploySlasherProxy {
		slasherProxy, err = k.DeployPellSlasherProxy(ctx, emptyContract, proxyAdmin, []byte{})
		require.NoError(t, err)
		require.NotEmpty(t, slasherProxy)
		assertContractDeployment(t, evmk, ctx, slasherProxy)
	}

	if config.DeployDvsDirectoryProxy {
		dvsDirectoryProxy, err = k.DeployPellDvsDirectoryProxy(ctx, emptyContract, proxyAdmin, []byte{})
		require.NoError(t, err)
		require.NotEmpty(t, dvsDirectoryProxy)
		assertContractDeployment(t, evmk, ctx, dvsDirectoryProxy)

		registryRouter, err = k.DeployPellRegistryRouter(ctx, dvsDirectoryProxy, systemContract)
		require.NoError(t, err)
		require.NotEmpty(t, registryRouter)
		assertContractDeployment(t, evmk, ctx, registryRouter)
	}

	return
}

func mockSuccessfulContractDeployment(ctx context.Context, t *testing.T, k *keeper.Keeper) {
	mockEVMKeeper := keepertest.GetPevmEVMMock(t, k)
	mockEVMKeeper.On("WithChainID", mock.Anything).Maybe().Return(ctx)
	mockEVMKeeper.On("ChainID").Maybe().Return(big.NewInt(1))
	mockEVMKeeper.On(
		"EstimateGas",
		mock.Anything,
		mock.Anything,
	).Return(&evmtypes.EstimateGasResponse{Gas: 5}, nil)
	mockEVMKeeper.MockEVMSuccessCallOnce()
}

func mockFailedContractDeployment(ctx context.Context, t *testing.T, k *keeper.Keeper) {
	mockEVMKeeper := keepertest.GetPevmEVMMock(t, k)
	mockEVMKeeper.On("WithChainID", mock.Anything).Maybe().Return(ctx)
	mockEVMKeeper.On("ChainID").Maybe().Return(big.NewInt(1))
	mockEVMKeeper.On(
		"EstimateGas",
		mock.Anything,
		mock.Anything,
	).Return(&evmtypes.EstimateGasResponse{Gas: 5}, nil)
	mockEVMKeeper.MockEVMFailCallOnce()
}
