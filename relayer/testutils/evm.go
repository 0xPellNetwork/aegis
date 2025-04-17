package testutils

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/strategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// ParseReceipt parses a Deposit event from a receipt
func ParseReceiptStakerDeposited(receipt *ethtypes.Receipt, strategyManager *strategymanager.StrategyManager) *strategymanager.StrategyManagerDeposit {
	for _, log := range receipt.Logs {
		event, err := strategyManager.ParseDeposit(*log)
		if err == nil && event != nil {
			return event // found
		}
	}
	return nil
}

// ParseReceipt parses an Delegate event from a receipt
func ParseReceiptERC20Deposited(receipt *ethtypes.Receipt, delegationManager *delegationmanager.DelegationManager) *delegationmanager.DelegationManagerStakerDelegated {
	for _, log := range receipt.Logs {
		event, err := delegationManager.ParseStakerDelegated(*log)
		if err == nil && event != nil {
			return event // found
		}
	}
	return nil
}
