package types

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/core/vm"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// TODO USE string constant
var (
	BigIntZero   = big.NewInt(0)
	PEVMGasLimit = big.NewInt(1_000_000)
)

// IsRevertError checks if an error is a evm revert error
func IsRevertError(err error) bool {
	return err != nil && strings.Contains(err.Error(), vm.ErrExecutionReverted.Error())
}

// IsContractReverted checks if the contract execution is reverted
func IsContractReverted(res *evmtypes.MsgEthereumTxResponse, err error) bool {
	return IsRevertError(err) || (res != nil && res.Failed())
}
