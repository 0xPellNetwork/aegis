package keeper

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/omnioperatorsharesmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/centralscheduler.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/ejectionmanager.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/operatorstakemanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	omniOperatorSharesManagerMetaDataABI *abi.ABI
	delegationManagerMetaDataABI         *abi.ABI
	registryRouterMetaDataABI            *abi.ABI
	stakeRegistryRouterMetaDataABI       *abi.ABI
	centralschedulerMetaDataABI          *abi.ABI
	operatorstakemanagerMetaDataABI      *abi.ABI
	ejectionManagerMetaDataABI           *abi.ABI
)

func init() {
	var err error
	delegationManagerMetaDataABI, err = delegationmanager.DelegationManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	omniOperatorSharesManagerMetaDataABI, err = omnioperatorsharesmanager.OmniOperatorSharesManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	registryRouterMetaDataABI, err = registryrouter.RegistryRouterMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	stakeRegistryRouterMetaDataABI, err = stakeregistryrouter.StakeRegistryRouterMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	centralschedulerMetaDataABI, err = centralscheduler.CentralSchedulerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	operatorstakemanagerMetaDataABI, err = operatorstakemanager.OperatorStakeManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	ejectionManagerMetaDataABI, err = ejectionmanager.EjectionManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}
