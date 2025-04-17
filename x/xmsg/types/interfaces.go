package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// CrossChainFee is a struct that contains the address and the fee for a cross-chain event
type CrossChainFee struct {
	Address sdk.AccAddress
	Fee     sdkmath.Int
}

// InternalEventLogHooks is an interface for handling event logs from other modules
type InternalEventLogHooks interface {
	HandleEventLogs(ctx sdk.Context, emittingContractAddr ethcommon.Address, logs []*ethtypes.Log, txOrigin string) error
}

// // EventHandler is an interface for handling pell evm events. internal handlers are
type EventHandler interface {
	// HandleEvent handles the event
	HandleEvent(ctx sdk.Context, epochNum uint64, emittingContractAddr ethcommon.Address, logs []*ethtypes.Log, txOrigin string) ([]*CrossChainFee, error)
	// GetContractAddress returns the contract address filter for the event handler
	GetContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	// ParseEvent parses the event from the log
	ParseEvent(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error)
}
