package keeper

import (
	"context"
	"math/big"

	cosmoserror "cosmossdk.io/errors"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/dvsdirectory.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pellslasher.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pellstrategymanager.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

// DeploySystemContracts deploy new instances of the system contracts
//
// Authorized: admin policy group 2.
func (k msgServer) DeploySystemContracts(goCtx context.Context, msg *types.MsgDeploySystemContracts) (*types.MsgDeploySystemContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, cosmoserror.Wrap(authoritytypes.ErrUnauthorized, "System contract deployment can only be executed by the correct policy account")
	}

	// system contract
	systemContract, err := k.DeployPellSystemContract(ctx, types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy SystemContract")
	}

	// connector contract
	connector, err := k.DeployPellConnector(ctx, systemContract, types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellConnector")
	}

	// empty contract
	emptyContract, err := k.DeployPellEmptyContract(ctx)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy EmptyContract")
	}

	// proxy_admin contract
	proxyAdmin, err := k.DeployPellProxyAdmin(ctx)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy ProxyAdmin")
	}

	// strategy_manager_proxy contract
	strategyManagerProxy, err := k.DeployPellStrategyManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellStrategyManagerProxy")
	}

	// delegation_manager_proxy contract
	delegationManagerProxy, err := k.DeployPellDelegationManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellDelegationManagerProxy")
	}

	// slash_proxy contract
	slasherProxy, err := k.DeployPellSlasherProxy(ctx, emptyContract, proxyAdmin, []byte{})
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellSlasherProxy")
	}

	// strategy_manager_impl contract
	strategyManagerImpl, err := k.DeployPellStrategyManager(ctx, delegationManagerProxy, slasherProxy, systemContract)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellStrategyManager")
	}

	// delegation_manager_impl contract
	delegationManagerImpl, err := k.DeployPellDelegationManager(ctx, strategyManagerProxy, slasherProxy, systemContract)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellDelegationManager")
	}

	// slasher_impl contract
	slasherImpl, err := k.DeployPellSlasher(ctx, strategyManagerProxy, delegationManagerProxy, systemContract)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellSlasherManager")
	}

	// initialize strategy manager contract
	_, err = k.CallMethodOnContractByProxyAdmin(ctx, proxyAdmin, strategyManagerProxy, strategyManagerImpl,
		pellstrategymanager.PellStrategyManagerMetaData, "initialize", types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to call initialize on StrategyManager")
	}

	// initialize delegation manager contract
	_, err = k.CallMethodOnContractByProxyAdmin(ctx, proxyAdmin, delegationManagerProxy, delegationManagerImpl,
		pelldelegationmanager.PellDelegationManagerMetaData, "initialize", types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to call initialize on DelegationManager")
	}

	// initialize slash contract
	_, err = k.CallMethodOnContractByProxyAdmin(ctx, proxyAdmin, slasherProxy, slasherImpl,
		pellslasher.PellSlasherMetaData, "initialize", types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to call initialize on Slasher")
	}

	// dvs_directory_impl contract
	dvsDirectoryImpl, err := k.DeployPellDvsDirectory(ctx, delegationManagerProxy)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy PellStrategyManager")
	}

	dvsAbi, err := dvsdirectory.DVSDirectoryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := dvsAbi.Pack("initialize", types.ModuleAddressEVM, []common.Address{types.ModuleAddressEVM}, types.ModuleAddressEVM, big.NewInt(0))
	if err != nil {
		return nil, err
	}

	// dvs_directory_proxy contract
	dvsDirectoryProxy, err := k.DeployPellDvsDirectoryProxy(ctx, dvsDirectoryImpl, proxyAdmin, data)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy pellDvsDirectoryProxy")
	}

	// registry router contract
	registryRouter, err := k.DeployPellRegistryRouter(ctx, dvsDirectoryProxy, systemContract)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy pellRegistryRouter")
	}

	registryRouterBeacon, err := k.DeployRegistryRouterBeacon(ctx, registryRouter)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy registryRouterBeacon")
	}

	stakeRegistryRouter, err := k.DeployPellStakeRegistryRouter(ctx, delegationManagerProxy)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy pellStakeRegistryRouter")
	}

	// stakeRegistryRouterBeacon contract
	stakeRegistryRouterBeacon, err := k.DeployStakeRegistryRouterBeacon(ctx, stakeRegistryRouter)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy stakeRegistryRouterBeacon")
	}

	// registry_router_factory contract
	registryRouterFactory, err := k.DeployPellRegistryRouterFactory(ctx, common.HexToAddress(msg.Signer), registryRouterBeacon, stakeRegistryRouterBeacon)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy pellRegistryRouterFactory")
	}

	wpell, err := k.DeployWrappedPell(ctx)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy WrappedPell")
	}

	gatewayPEVM, err := k.DeployGatewayPEVM(ctx, connector, systemContract, wpell)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy GatewayEVM")
	}

	gasSwap, err := k.DeployGasSwapPEVM(ctx, connector, types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy GasSwap")
	}

	// update systemcontract with owner
	_, err = k.CallMethodOnSystemContract(ctx, systemContract, "updateModuleAddress", "StakingModule", types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to call updateModuleAddress on SystemContract")
	}

	pellChainIDInt, err := chains.CosmosToEthChainID(ctx.ChainID())
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to get pell chain id on DeploySystemContracts")
	}

	pellChainID := big.NewInt(pellChainIDInt)
	// update gateway source address and destination address
	_, err = k.CallMethodOnGateway(ctx, gatewayPEVM, "updateDestinationAddress", pellChainID, connector.Bytes())
	if err != nil {
		k.Logger(ctx).Error("failed to call updateDestinationAddress on gateway contract", "error", err)
		return nil, cosmoserror.Wrapf(err, "failed to call updateDestinationAddress on gateway contract")
	}

	_, err = k.CallMethodOnGateway(ctx, gatewayPEVM, "updateSourceAddress", pellChainID, connector.Bytes())
	if err != nil {
		k.Logger(ctx).Error("failed to call updateSourceAddress on gateway contract", "error", err)
		return nil, cosmoserror.Wrapf(err, "failed to call updateSourceAddress on gateway contract")
	}

	err = ctx.EventManager().EmitTypedEvent(
		&types.EventSystemContractsDeployed{
			MsgTypeUrl:             sdk.MsgTypeURL(&types.MsgDeploySystemContracts{}),
			SystemContract:         systemContract.Hex(),
			Connector:              connector.Hex(),
			EmptyContract:          emptyContract.Hex(),
			ProxyAdmin:             proxyAdmin.Hex(),
			DelegationManagerProxy: delegationManagerProxy.Hex(),
			StrategyManagerProxy:   strategyManagerProxy.Hex(),
			SlasherProxy:           slasherProxy.Hex(),
			DelegationManagerImpl:  delegationManagerImpl.Hex(),
			StrategyManagerImpl:    strategyManagerImpl.Hex(),
			SlasherImpl:            slasherImpl.Hex(),
			DvsDirectory:           dvsDirectoryImpl.Hex(),
			DvsDirectoryProxy:      dvsDirectoryProxy.Hex(),
			RegistryRouter:         registryRouter.Hex(),
			RegistryRouterFactory:  registryRouterFactory.Hex(),
			Signer:                 msg.Signer,
			WrappedPell:            wpell.Hex(),
			Gateway:                gatewayPEVM.Hex(),
			GasSwap:                gasSwap.Hex(),
			StakeRegistryRouter:    stakeRegistryRouter.Hex(),
		},
	)
	if err != nil {
		k.Logger(ctx).Error("failed to emit event",
			"event", "EventSystemContractsDeployed",
			"error", err.Error(),
		)
		return nil, cosmoserror.Wrapf(types.ErrEmitEvent, "failed to emit event (%s)", err.Error())
	}

	return &types.MsgDeploySystemContractsResponse{
		SystemContract:         systemContract.Hex(),
		Connector:              connector.Hex(),
		EmptyContract:          emptyContract.Hex(),
		ProxyAdmin:             proxyAdmin.Hex(),
		DelegationManagerProxy: delegationManagerProxy.Hex(),
		StrategyManagerProxy:   strategyManagerProxy.Hex(),
		SlasherProxy:           slasherProxy.Hex(),
		DelegationManagerImpl:  delegationManagerImpl.Hex(),
		StrategyManagerImpl:    strategyManagerImpl.Hex(),
		SlasherImpl:            slasherImpl.Hex(),
		DvsDirectoryImpl:       dvsDirectoryImpl.Hex(),
		DvsDirectoryProxy:      dvsDirectoryProxy.Hex(),
		RegistryRouter:         registryRouter.Hex(),
		RegistryRouterFactory:  registryRouterFactory.Hex(),
		WrappedPell:            wpell.Hex(),
		Gateway:                gatewayPEVM.Hex(),
		GasSwap:                gasSwap.Hex(),
		StakeRegistryRouter:    stakeRegistryRouter.Hex(),
	}, nil
}
