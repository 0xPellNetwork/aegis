package keeper_test

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/pevm/keeper"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

func TestMsgServer_DeploySystemContracts(t *testing.T) {
	t.Run("admin can deploy system contracts", func(t *testing.T) {
		k, ctx, sdkk, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
		})

		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		res, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.NoError(t, err)
		require.NotNil(t, res)
		assertContractDeployment(t, sdkk.EvmKeeper, ctx, ethcommon.HexToAddress(res.SystemContract))
		assertContractDeployment(t, sdkk.EvmKeeper, ctx, ethcommon.HexToAddress(res.Connector))
		assertContractDeployment(t, sdkk.EvmKeeper, ctx, ethcommon.HexToAddress(res.DelegationManagerProxy))
		assertContractDeployment(t, sdkk.EvmKeeper, ctx, ethcommon.HexToAddress(res.StrategyManagerProxy))
		assertContractDeployment(t, sdkk.EvmKeeper, ctx, ethcommon.HexToAddress(res.SlasherProxy))
		assertContractDeployment(t, sdkk.EvmKeeper, ctx, ethcommon.HexToAddress(res.RegistryRouterFactory))
	})

	t.Run("non-admin cannot deploy system contracts", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
		})

		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		nonadmin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, nonadmin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(nonadmin))
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
	})

	t.Run("should fail if system contract deployment fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
			UseEVMMock:       true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// mock failed systemContract deployment
		mockFailedContractDeployment(ctx, t, k)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to deploy")
	})

	t.Run("should fail if connector contract deployment fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
			UseEVMMock:       true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// mock successful systemContract deployment
		mockSuccessfulContractDeployment(ctx, t, k)
		// mock failed connector deployment
		mockFailedContractDeployment(ctx, t, k)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to deploy")
	})

	t.Run("should fail if strategyManagerProxy contract deployment fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
			UseEVMMock:       true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// mock successful systemContract, connector deployment
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		// mock failed strategyManagerProxy  deployment
		mockFailedContractDeployment(ctx, t, k)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to deploy")
	})

	t.Run("should fail if delegationManagerInteractorProxy deployment fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
			UseEVMMock:       true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// mock successful systemContract, connector and strategyManagerProxy deployments
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		// mock failed delegationManagerInteractorProxy deployment
		mockFailedContractDeployment(ctx, t, k)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to deploy")
	})

	t.Run("should fail if delegationManagerProxy deployment fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
			UseEVMMock:       true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// mock successful systemContract, connector, strategyManagerProxy and delegationManagerInteractorProxy deployments
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		// mock failed delegationManagerProxy deployment
		mockFailedContractDeployment(ctx, t, k)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to deploy")
	})

	t.Run("should fail if slashProxy deployment fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.PevmKeeperWithMocks(t, keepertest.PevmMockOptions{
			UseAuthorityMock: true,
			UseEVMMock:       true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		_ = k.GetAuthKeeper().GetModuleAccount(ctx, types.ModuleName)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetPevmAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// mock successful systemContract, connector, strategyManagerProxy, delegationManagerInteractorProxy and delegationManagerProxy deployments
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		mockSuccessfulContractDeployment(ctx, t, k)
		// mock failed slashProxy deployment
		mockFailedContractDeployment(ctx, t, k)

		_, err := msgServer.DeploySystemContracts(ctx, types.NewMsgDeploySystemContracts(admin))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to deploy")
	})
}
