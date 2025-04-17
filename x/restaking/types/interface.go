package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// // EventHandler is an interface for handling pell evm events. internal handlers are
type EventHandler interface {
	// HandleEvent handles the event
	HandleEvent(ctx sdk.Context, epochNum uint64, emittingContractAddr ethcommon.Address, logs []*ethtypes.Log, txOrigin string) ([]*xmsgtypes.CrossChainFee, error)
	// GetContractAddress returns the contract address filter for the event handler
	GetContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	// ParseEvent parses the event from the log
	ParseEvent(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error)
}

// MiddlewareEventSubscriber is an interface for handling middleware events
type MiddlewareEventSubscriber interface {
	// ProcessLogs parse event and builds the pell sent event from the middleware event
	ProcessLogs(ctx sdk.Context, _ uint64, toAddress ethcommon.Address, log *ethtypes.Log, txOrigin string) (*sdkmath.Int, error)
}
