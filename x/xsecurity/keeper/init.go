package keeper

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	registryRouterFactoryMedaDataABI *abi.ABI
	registryRouterMetaDataABI        *abi.ABI
)

func init() {
	var err error

	registryRouterMetaDataABI, err = registryrouter.RegistryRouterMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	registryRouterFactoryMedaDataABI, err = registryrouterfactory.RegistryRouterFactoryMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}
