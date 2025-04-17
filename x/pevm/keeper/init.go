package keeper

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/bridge/gatewaypevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pellstrategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/swap/gasswappevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/systemcontract.sol"
	"github.com/0xPellNetwork/contracts/pkg/openzeppelin/contracts/proxy/beacon/upgradeablebeacon.sol"
	"github.com/0xPellNetwork/contracts/pkg/openzeppelin/contracts/proxy/transparent/proxyadmin.sol"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	systemContractMetaDataABI                  *abi.ABI
	delegationManagerMetaDataABI               *abi.ABI
	proxyAdminMetaDataABI                      *abi.ABI
	pellStrategyManagerMetaDataABI             *abi.ABI
	gatewayPEVMMetaDataABI                     *abi.ABI
	gasSwapPEVMMetaDataABI                     *abi.ABI
	pellDelegationManagerInteractorMetaDataABI *abi.ABI
	registryRouterMetaDataABI                  *abi.ABI
	upgradeableBeaconMetaDataABI               *abi.ABI
	pellConnectorMetaDataABI                   *abi.ABI
	registryRouterFactoryMetaDataABI           *abi.ABI
	stakeRegistryRouterMetaDataABI             *abi.ABI
)

func init() {
	var err error
	systemContractMetaDataABI, err = systemcontract.SystemContractMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	proxyAdminMetaDataABI, err = proxyadmin.ProxyAdminMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	pellStrategyManagerMetaDataABI, err = pellstrategymanager.PellStrategyManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	delegationManagerMetaDataABI, err = pelldelegationmanager.PellDelegationManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	gatewayPEVMMetaDataABI, err = gatewaypevm.GatewayPEVMMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	gasSwapPEVMMetaDataABI, err = gasswappevm.GasSwapPEVMMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	registryRouterMetaDataABI, err = registryrouter.RegistryRouterMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	upgradeableBeaconMetaDataABI, err = upgradeablebeacon.UpgradeableBeaconMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	pellConnectorMetaDataABI, err = pellconnector.PellConnectorMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	registryRouterFactoryMetaDataABI, err = registryrouterfactory.RegistryRouterFactoryMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	stakeRegistryRouterMetaDataABI, err = stakeregistryrouter.StakeRegistryRouterMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}
