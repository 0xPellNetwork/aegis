package evm

import "time"

const (
	// PellBlockTime is the block time of the Pell network
	PellBlockTime = 6500 * time.Millisecond

	// OutTxInclusionTimeout is the timeout for waiting for an outtx to be included in a block
	OutTxInclusionTimeout = 20 * time.Minute

	// OutTxTrackerReportTimeout is the timeout for waiting for an outtx tracker report
	OutTxTrackerReportTimeout = 10 * time.Minute

	// EthTransferGasLimit is the gas limit for a standard ETH transfer
	EthTransferGasLimit = 21000

	// [signature]
	TopicsPellStakerDeposited = 1

	// [signature, stakerAddress, operatorAddress]
	TopicsPellStakerDelegated = 3

	// [signature]
	TopicsPellWithdrawalQueued = 1

	// [signature, stakerAddress, operatorAddress]
	TopicsPellStakerUndelegated = 3

	TopicsCentralSchedulerToPell        = 1
	TopicsRegisterStakeManagerToPell    = 1
	TopicsRegisterEjectionManagerToPell = 1
	TopicsPellSent                      = 3
)
