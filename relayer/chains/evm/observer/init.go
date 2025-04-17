package observer

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

const (
	PellMessageFailedEventName = "PellMessageFailed"
)

var (
	ConnectorContractABI *abi.ABI
)

func init() {
	var err error
	ConnectorContractABI, err = pellconnector.PellConnectorMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}
