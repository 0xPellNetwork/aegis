package pellcore

import "time"

const (
	// DefaultBaseGasPrice is the default base gas price
	DefaultBaseGasPrice = 1_000_000

	// DefaultGasLimit is the default gas limit used for broadcasting txs
	DefaultGasLimit = 200_000

	// PostGasPriceGasLimit is the gas limit for voting new gas price
	PostGasPriceGasLimit = 1_500_000

	// AddTxHashToOutTxTrackerGasLimit is the gas limit for adding tx hash to out tx tracker
	AddTxHashToOutTxTrackerGasLimit = 200_000

	// PostTSSGasLimit is the gas limit for voting on TSS keygen
	PostTSSGasLimit = 500_000

	// PostVoteInboundExecutionGasLimit is the gas limit for voting on observed inbound tx and executing it
	PostVoteInboundExecutionGasLimit = 6_500_000

	// PostVoteInboundMessagePassingExecutionGasLimit is the gas limit for voting on, and executing ,observed inbound tx related to message passing (coin_type == pell)
	PostVoteInboundMessagePassingExecutionGasLimit = 4_000_000

	// AddOutboundTrackerGasLimit is the gas limit for adding tx hash to out tx tracker
	AddOutboundTrackerGasLimit = 200_000

	// PostBlameDataGasLimit is the gas limit for voting on blames
	PostBlameDataGasLimit = 200_000

	// PostAddPellTokenGasLimit is the gas limit for voting on observed pell token
	PostAddPellTokenGasLimit = 1_500_000

	// PostAddGasTokenGasLimit is the gas limit for voting on observed gas token
	PostAddGasTokenGasLimit = 1_500_000

	// DefaultRetryCount is the number of retries for broadcasting a tx
	DefaultRetryCount = 5

	// ExtendedRetryCount is an extended number of retries for broadcasting a tx, used in keygen operations
	ExtendedRetryCount = 15

	// DefaultRetryInterval is the interval between retries in seconds
	DefaultRetryInterval = 5

	// PostVoteInboundGasLimit is the gas limit for voting on observed inbound tx.
	// Supports up to 256 events, but there may be more in practice
	PostVoteInboundGasLimit = 1_500_000 * 256

	// PostVoteOutboundGasLimit is the gas limit for voting on observed outbound tx
	PostVoteOutboundGasLimit = 1_500_000

	// PostVoteOutboundRevertGasLimit is the gas limit for voting on observed outbound tx for revert (when outbound fails)
	// The value needs to be higher because reverting implies interacting with the EVM to perform swaps for the gas token
	PostVoteOutboundRevertGasLimit = 1_500_000

	// MonitorVoteInboundTxResultInterval is the interval between retries for monitoring tx result in seconds
	MonitorVoteInboundTxResultInterval = 5

	// MonitorVoteInboundTxResultRetryCount is the number of retries to fetch monitoring tx result
	MonitorVoteInboundTxResultRetryCount = 20

	// MonitorVoteOutboundTxResultInterval is the interval between retries for monitoring tx result in seconds
	MonitorVoteOutboundTxResultInterval = 5

	// MonitorVoteOutboundTxResultRetryCount is the number of retries to fetch monitoring tx result
	MonitorVoteOutboundTxResultRetryCount = 20
)

// constants for monitoring tx results
const (
	monitorInterval   = 5 * time.Second
	monitorRetryCount = 20
)
