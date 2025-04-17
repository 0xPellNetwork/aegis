package signer

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	connectorABI *abi.ABI
)

func init() {
	var err error
	connectorABI, err = pellconnector.PellConnectorMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}
