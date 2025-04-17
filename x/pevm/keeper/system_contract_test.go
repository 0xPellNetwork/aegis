package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	testkeeper "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

func TestKeeper_DeploySystemContracts(t *testing.T) {
	t.Run("system contract deployment should error if deploy contract fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseEVMMock: true,
		})
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		mockFailedContractDeployment(ctx, t, k)

		res, err := k.DeployPellSystemContract(ctx, types.ModuleAddressEVM)
		require.Error(t, err)
		require.Empty(t, res)

		// assertContractDeployment(t, mockEVMKeeper, ctx, rs)
		// _, err = k.CallMethodOnSystemContract(ctx, res, "updateModuleAddress", "StakingModule", types.ModuleAddressEVM)
		// require.NoError(t, err)
	})

	t.Run("strategyManagerImpl deployment should error if deploy contract fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseEVMMock: true,
		})
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		mockFailedContractDeployment(ctx, t, k)

		strategyManagerImpl, err := k.DeployPellStrategyManager(ctx, sample.EthAddress(), sample.EthAddress(), sample.EthAddress())
		require.Error(t, err)
		require.Empty(t, strategyManagerImpl)

		// assertContractDeployment(t, mockEVMKeeper, ctx, strategyManagerImpl)
		// _, err = k.CallMethodOnContractByProxyAdmin(ctx, proxyAdmin, strategyManagerProxy, strategyManagerImpl,
		// 	pellstrategymanager.PellStrategyManagerMetaData, "initialize", types.ModuleAddressEVM)
		// require.Error(t, err)
	})

	t.Run("delegationManagerImpl deployment should error if deploy contract fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseEVMMock: true,
		})
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		mockFailedContractDeployment(ctx, t, k)

		delegationManagerImpl, err := k.DeployPellDelegationManager(ctx, sample.EthAddress(), sample.EthAddress(), sample.EthAddress())
		require.Error(t, err)
		require.Empty(t, delegationManagerImpl)

		// assertContractDeployment(t, mockEVMKeeper, ctx, delegationManagerImpl)
		// _, err = k.CallMethodOnContractByProxyAdmin(ctx, proxyAdmin, delegationManagerProxy, delegationManagerImpl,
		// 	pelldelegationmanager.PellDelegationManagerMetaData, "initialize", types.ModuleAddressEVM)
		// require.Error(t, err)
	})

	t.Run("slasherImpl deployment should error if deploy contract fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseEVMMock: true,
		})
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		mockFailedContractDeployment(ctx, t, k)

		slasherImpl, err := k.DeployPellSlasher(ctx, sample.EthAddress(), sample.EthAddress(), sample.EthAddress())
		require.Error(t, err)
		require.Empty(t, slasherImpl)

		// assertContractDeployment(t, mockEVMKeeper, ctx, slasherImpl)
		// _, err = k.CallMethodOnContractByProxyAdmin(ctx, proxyAdmin, slasherProxy, slasherImpl,
		// 	pellslasher.PellSlasherMetaData, "initialize", types.ModuleAddressEVM)
		// require.Error(t, err)
	})

	t.Run("registryRouterFactory deployment should error if deploy contract fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseEVMMock: true,
		})
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		mockFailedContractDeployment(ctx, t, k)

		registryRouterFactory, err := k.DeployPellRegistryRouterFactory(ctx, sample.EthAddress(), sample.EthAddress(), sample.EthAddress())
		require.Error(t, err)
		require.Empty(t, registryRouterFactory)

		// assertContractDeployment(t, mockEVMKeeper, ctx, registryRouterFactory)
	})

	t.Run("can deploy the system contracts", func(t *testing.T) {
		k, ctx, sdkk, _ := testkeeper.PevmKeeper(t)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		// deploy the system contracts
		systemContract, connector, _, strategyManagerProxy, delegationManagerProxy, slasherProxy, dvsDirectoryProxy, registryRouter := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)

		// can find system contract address
		found, err := k.GetPellSystemContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, systemContract, found)

		// can find connector constract address
		found, err = k.GetPellConnectorContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, connector, found)

		// can find strategyManagerProxy contract address
		found, err = k.GetPellStrategyManagerProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, strategyManagerProxy, found)

		// can find the delegationManagerProxy contract address
		found, err = k.GetPellDelegationManagerProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, delegationManagerProxy, found)

		// can find the slasherProxy contract address
		found, err = k.GetPellSlasherProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, slasherProxy, found)

		// can find the dvsDirectoryProxy contract address
		found, err = k.GetPellDvsDirectoryProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, dvsDirectoryProxy, found)

		// can find the registryRouter contract address
		found, err = k.GetPellRegistryRouterContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, registryRouter, found)
	})

}

func TestKeeper_GetSystemContract(t *testing.T) {
	t.Run("should get and remove system contract", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.SetSystemContract(ctx, types.SystemContract{SystemContract: "test"})
		val, found := k.GetSystemContract(ctx)
		require.True(t, found)
		require.Equal(t, types.SystemContract{SystemContract: "test"}, val)

		// can remove contract
		k.RemoveSystemContract(ctx)
		_, found = k.GetSystemContract(ctx)
		require.False(t, found)
	})
}

func TestKeeper_GetPellSystemContractAddress(t *testing.T) {
	t.Run("should fail to get system contract address if system contracts are not deployed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, err := k.GetPellSystemContractAddress(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrStateVariableNotFound)
	})

	t.Run("should get system contract address if system contracts are deployed", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		systemContract, _, _, _, _, _, _, _ := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)
		found, err := k.GetPellSystemContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, systemContract, found)
	})
}

func TestKeeper_GetPellConnectorContractAddress(t *testing.T) {
	t.Run("should fail to get connector contract address if system contracts are not deployed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, err := k.GetPellConnectorContractAddress(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrStateVariableNotFound)
	})

	t.Run("should get connector contract address if system contracts are deployed", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, connector, _, _, _, _, _, _ := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)
		found, err := k.GetPellConnectorContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, connector, found)
	})
}

func TestKeeper_GetStrategyManagerProxyAddress(t *testing.T) {
	t.Run("should fail to get strategyManagerProxy contract address if system contracts are not deployed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, err := k.GetPellStrategyManagerProxyContractAddress(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrStateVariableNotFound)
	})

	t.Run("should get strategyManagerProxy contract address if system contracts are deployed", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, _, _, strategyManagerProxy, _, _, _, _ := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)
		found, err := k.GetPellStrategyManagerProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, strategyManagerProxy, found)
	})
}

func TestKeeper_GetPellDelegationManagerProxyAddress(t *testing.T) {
	t.Run("should fail to get delegationManagerProxy contract address if system contracts are not deployed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, err := k.GetPellDelegationManagerProxyContractAddress(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrStateVariableNotFound)
	})

	t.Run("should get delegationManagerProxy contract address if system contracts are deployed", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, _, _, _, delegationManagerProxy, _, _, _ := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)
		found, err := k.GetPellDelegationManagerProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, delegationManagerProxy, found)
	})

}

func TestKeeper_GetPellSlasherProxyAddress(t *testing.T) {
	t.Run("should fail to get slasherProxy contract address if system contracts are not deployed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, err := k.GetPellDelegationManagerProxyContractAddress(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrStateVariableNotFound)
	})

	t.Run("should get slasherProxy contract address if system contracts are deployed", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, _, _, _, _, slasherProxy, _, _ := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)
		found, err := k.GetPellSlasherProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, slasherProxy, found)
	})

}

func TestKeeper_GetDvsDirectoryProxyAddress(t *testing.T) {
	t.Run("should fail to get dvsDirectoryProxy contract address if system contracts are not deployed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, err := k.GetPellDvsDirectoryProxyContractAddress(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrStateVariableNotFound)
	})

	t.Run("should get dvsDirectoryProxy contract address if system contracts are deployed", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeper(t)
		k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)

		_, _, _, _, _, _, dvsDirectoryProxy, _ := deploySystemContracts(t, ctx, k, sdkk.EvmKeeper)
		found, err := k.GetPellDvsDirectoryProxyContractAddress(ctx)
		require.NoError(t, err)
		require.Equal(t, dvsDirectoryProxy, found)
	})

}
