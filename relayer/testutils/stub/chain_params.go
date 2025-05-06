package stub

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v2/strategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/0xPellNetwork/aegis/relayer/testutils"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func MockChainParams(chainID int64, confirmation uint64) relayertypes.ChainParams {
	return relayertypes.ChainParams{
		ChainId:                  chainID,
		ConfirmationCount:        confirmation,
		ConnectorContractAddress: testutils.ConnectorAddresses[chainID].Hex(),
		IsSupported:              true,
	}
}

func MockPellConnector(chainID int64) *pellconnector.PellConnector {
	connector, err := pellconnector.NewPellConnector(testutils.ConnectorAddresses[chainID], &ethclient.Client{})
	if err != nil {
		panic(err)
	}
	return connector
}

func MockStrategyManager(chainID int64) *strategymanager.StrategyManager {
	sm, err := strategymanager.NewStrategyManager(testutils.StrategyManagerAddresses[chainID], &ethclient.Client{})
	if err != nil {
		panic(err)
	}
	return sm
}

func MockDelegationManager(chainID int64) *delegationmanager.DelegationManager {
	dm, err := delegationmanager.NewDelegationManager(testutils.DelegationManagerAddresses[chainID], &ethclient.Client{})
	if err != nil {
		panic(err)
	}
	return dm
}
