package keeper

import (
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/bridge/gatewaypevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/dvsdirectory.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/mocks/emptycontract.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pellstrategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/swap/gasswappevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/systemcontract.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/tokens/wpell.sol"
	"github.com/0xPellNetwork/contracts/pkg/openzeppelin/contracts/proxy/beacon/upgradeablebeacon.sol"
	"github.com/0xPellNetwork/contracts/pkg/openzeppelin/contracts/proxy/transparent/proxyadmin.sol"
	"github.com/0xPellNetwork/contracts/pkg/openzeppelin/contracts/proxy/transparent/transparentupgradeableproxy.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

// SetSystemContract set system contract in the store
func (k Keeper) SetSystemContract(ctx sdk.Context, sytemContract types.SystemContract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SystemContractKey))
	b := k.cdc.MustMarshal(&sytemContract)
	store.Set([]byte{0}, b)
}

// GetSystemContract returns system contract from the store
func (k Keeper) GetSystemContract(ctx sdk.Context) (val types.SystemContract, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SystemContractKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSystemContract removes system contract from the store
func (k Keeper) RemoveSystemContract(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SystemContractKey))
	store.Delete([]byte{0})
}

func (k Keeper) DeployPellSystemContract(ctx sdk.Context, owner common.Address) (common.Address, error) {
	system, _ := k.GetSystemContract(ctx)

	contractAddr, err := k.DeployContract(ctx, systemcontract.SystemContractMetaData, owner)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "failed to deploy SystemContract")
	}

	system.SystemContract = contractAddr.String()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployPellConnector(ctx sdk.Context, systemContract, pauserAddress common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, pellconnector.PellConnectorMetaData, systemContract, pauserAddress)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "PellConnector")
	}

	system, _ := k.GetSystemContract(ctx)
	system.Connector = contractAddr.Hex()
	k.SetSystemContract(ctx, system)
	return contractAddr, nil
}

func (k Keeper) DeployPellEmptyContract(ctx sdk.Context) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, emptycontract.EmptyContractMetaData)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "EmptyContract")
	}
	return contractAddr, nil
}

func (k Keeper) DeployPellProxyAdmin(ctx sdk.Context) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, proxyadmin.ProxyAdminMetaData)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "ProxyAdmin")
	}
	return contractAddr, nil
}

func (k Keeper) DeployPellDelegationManagerProxy(ctx sdk.Context, emptyContract, proxyAdmin common.Address, reserver []byte) (common.Address, error) {
	contractAddr, err := k.deployPellTransparentUpgradeableProxy(ctx, emptyContract, proxyAdmin, reserver)
	if err == nil {
		system, _ := k.GetSystemContract(ctx)
		system.DelegationManagerProxy = contractAddr.Hex()
		k.SetSystemContract(ctx, system)
	}

	return contractAddr, err
}

func (k Keeper) DeployPellStrategyManagerProxy(ctx sdk.Context, emptyContract, proxyAdmin common.Address, reserver []byte) (common.Address, error) {
	contractAddr, err := k.deployPellTransparentUpgradeableProxy(ctx, emptyContract, proxyAdmin, reserver)
	if err == nil {
		system, _ := k.GetSystemContract(ctx)
		system.StrategyManagerProxy = contractAddr.Hex()
		k.SetSystemContract(ctx, system)
	}

	return contractAddr, err
}

func (k Keeper) DeployPellSlasherProxy(ctx sdk.Context, emptyContract, proxyAdmin common.Address, reserver []byte) (common.Address, error) {
	contractAddr, err := k.deployPellTransparentUpgradeableProxy(ctx, emptyContract, proxyAdmin, reserver)
	if err == nil {
		system, _ := k.GetSystemContract(ctx)
		system.SlasherProxy = contractAddr.Hex()
		k.SetSystemContract(ctx, system)
	}

	return contractAddr, err
}

func (k Keeper) DeployPellDvsDirectoryProxy(ctx sdk.Context, contract, proxyAdmin common.Address, reserver []byte) (common.Address, error) {
	contractAddr, err := k.deployPellTransparentUpgradeableProxy(ctx, contract, proxyAdmin, reserver)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.DvsDirectoryProxy = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, err
}

func (k Keeper) deployPellTransparentUpgradeableProxy(ctx sdk.Context, emptyContract, proxyAdmin common.Address, reserver []byte) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, transparentupgradeableproxy.TransparentUpgradeableProxyMetaData, emptyContract, proxyAdmin, reserver)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "TransparentUpgradeableProxy")
	}
	return contractAddr, nil
}

func (k Keeper) DeployPellDelegationManager(ctx sdk.Context, strategyManagerProxy, slasherProxy, systemContract common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, pelldelegationmanager.PellDelegationManagerMetaData, strategyManagerProxy, slasherProxy, systemContract)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "PellDelegationManager")
	}
	return contractAddr, nil
}

func (k Keeper) DeployPellStrategyManager(ctx sdk.Context, delegationManagerProxy, slasherProxy, systemContract common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, pellstrategymanager.PellStrategyManagerMetaData, delegationManagerProxy, slasherProxy, systemContract)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "PellStrategyManager")
	}
	return contractAddr, nil
}

func (k Keeper) DeployPellSlasher(ctx sdk.Context, strategyManagerProxy, delegationManagerProxy, systemContract common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, pellstrategymanager.PellStrategyManagerMetaData, strategyManagerProxy, delegationManagerProxy, systemContract)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "PellSlahser")
	}
	return contractAddr, nil
}

func (k Keeper) DeployPellDvsDirectory(ctx sdk.Context, delegation common.Address) (common.Address, error) {
	return k.DeployContract(ctx, dvsdirectory.DVSDirectoryMetaData, delegation)
}

func (k Keeper) DeployPellRegistryRouter(ctx sdk.Context, dvsDirectoryProxy, systemContractAddr common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, registryrouter.RegistryRouterMetaData, dvsDirectoryProxy, systemContractAddr)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.RegistryRouter = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployPellStakeRegistryRouter(ctx sdk.Context, delegationAddr common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, stakeregistryrouter.StakeRegistryRouterMetaData, delegationAddr)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.StakeRegistryRouter = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployRegistryRouterBeacon(ctx sdk.Context, registryRouter common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, upgradeablebeacon.UpgradeableBeaconMetaData, registryRouter)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.RegistryRouterBeacon = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployStakeRegistryRouterBeacon(ctx sdk.Context, stakeRegistryRouter common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, upgradeablebeacon.UpgradeableBeaconMetaData, stakeRegistryRouter)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.StakeRegistryRouterBeacon = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployPellRegistryRouterFactory(ctx sdk.Context, owner, registryRouterBeacon, stakeRegistryRouterBeacon common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, registryrouterfactory.RegistryRouterFactoryMetaData, owner, registryRouterBeacon, stakeRegistryRouterBeacon)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.RegistryRouterFactory = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployWrappedPell(ctx sdk.Context) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, wpell.WPELLMetaData)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.WrappedPell = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployGatewayPEVM(ctx sdk.Context, connectorAddr, systemContractAddr, wpellAddr common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, gatewaypevm.GatewayPEVMMetaData, connectorAddr, systemContractAddr, wpellAddr)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.Gateway = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) DeployGasSwapPEVM(ctx sdk.Context, connectorAddr, systemContractAddr common.Address) (common.Address, error) {
	contractAddr, err := k.DeployContract(ctx, gasswappevm.GasSwapPEVMMetaData, connectorAddr, systemContractAddr)
	if err != nil {
		return common.Address{}, err
	}

	system, _ := k.GetSystemContract(ctx)
	system.GasSwap = contractAddr.Hex()
	k.SetSystemContract(ctx, system)

	return contractAddr, nil
}

func (k Keeper) UpgradeRegistryRouterBeacon(ctx sdk.Context, to, upgradeTo common.Address) (*evmtypes.MsgEthereumTxResponse, error) {
	return k.callBeaconUpgrade(ctx, to, upgradeTo)
}

// GetPellSystemContractAddress returns the system contract address
// TODO : GetPellSystemContractAddress and other constant strings , can be declared as a constant string in types
// TODO Remove repetitive code
func (k *Keeper) GetPellSystemContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	// set the system contract
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	systemAddress := ethcommon.HexToAddress(system.SystemContract)
	return systemAddress, nil
}

// GetConnectorContractAddress returns the PellConnector contract address on PellChain
func (k *Keeper) GetPellConnectorContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	connectorAddress := ethcommon.HexToAddress(system.Connector)
	return connectorAddress, nil
}

// GetStrategyManagerProxyContractAddress returns the PellStrategyManagerProxy contract address on PellChain
func (k *Keeper) GetPellStrategyManagerProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	strategyManagerProxyAddress := ethcommon.HexToAddress(system.StrategyManagerProxy)
	return strategyManagerProxyAddress, nil
}

// GetPellDelegationManagerProxyContractAddress returns the PellDelegationManagerProxy contract address on PellChain
func (k *Keeper) GetPellDelegationManagerProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	delegationManagerProxyAddress := ethcommon.HexToAddress(system.DelegationManagerProxy)
	return delegationManagerProxyAddress, nil
}

func (k *Keeper) GetPellGatewayEVMContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	gatewayEvmAddr := ethcommon.HexToAddress(system.Gateway)
	return gatewayEvmAddr, nil
}

func (k *Keeper) GetGasSwapPEVMContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	addr := ethcommon.HexToAddress(system.GasSwap)
	return addr, nil
}

// GetPellSlasherProxyContractAddress returns the PellSlasherProxy contract address on PellChain
func (k *Keeper) GetPellSlasherProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	slasherProxyAddress := ethcommon.HexToAddress(system.SlasherProxy)
	return slasherProxyAddress, nil
}

// GetPellDvsDirectoryProxyContractAddress returns the PellSlasherProxy contract address on PellChain
func (k *Keeper) GetPellDvsDirectoryProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	dvsDirectoryProxyAddress := ethcommon.HexToAddress(system.DvsDirectoryProxy)
	return dvsDirectoryProxyAddress, nil
}

// GetPellDvsDirectoryProxyContractAddress returns the PellSlasherProxy contract address on PellChain
func (k *Keeper) GetPellRegistryRouterContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	registryRouterAddress := ethcommon.HexToAddress(system.RegistryRouter)
	return registryRouterAddress, nil
}

// ------ LST Token staking --------
func (k *Keeper) GetRegistryRouterFactoryContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	system, found := k.GetSystemContract(ctx)
	if !found {
		return ethcommon.Address{}, cosmoserrors.Wrapf(types.ErrStateVariableNotFound, "failed to get system contract variable")
	}
	addr := ethcommon.HexToAddress(system.RegistryRouterFactory)
	return addr, nil
}
