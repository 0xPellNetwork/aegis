package sample

import (
	"github.com/pell-chain/pellcore/x/pevm/types"
)

func SystemContract_pell() *types.SystemContract {
	return &types.SystemContract{
		SystemContract:         EthAddress().String(),
		Connector:              EthAddress().String(),
		DelegationManagerProxy: EthAddress().String(),
		StrategyManagerProxy:   EthAddress().String(),
		SlasherProxy:           EthAddress().String(),
		DvsDirectoryProxy:      EthAddress().String(),
		RegistryRouter:         EthAddress().String(),
		RegistryRouterFactory:  EthAddress().String(),
	}
}
