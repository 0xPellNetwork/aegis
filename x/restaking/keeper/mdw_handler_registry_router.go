package keeper

import (
	"fmt"
	"reflect"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ types.EventHandler = &MiddlewareEventHandler{}

type MiddlewareEventHandler struct {
	Keeper
	subscriberRegistry []types.MiddlewareEventSubscriber
}

func NewMiddlewareEventHandler(keeper Keeper) *MiddlewareEventHandler {
	return &MiddlewareEventHandler{
		Keeper: keeper,
	}
}

func (h *MiddlewareEventHandler) RegisterAllEventSubscriber() {
	h.subscriberRegistry = []types.MiddlewareEventSubscriber{
		NewRegistryRouterEventSubscriber(h.Keeper),
		NewRegistryRouterFactoryEventSubscriber(h.Keeper),
	}
}

// HandleEvent handles the event from the log
func (h *MiddlewareEventHandler) HandleEvent(ctx sdk.Context, _ uint64, toAddress ethcommon.Address, logs []*ethtypes.Log, txOrigin string) ([]*xmsgtypes.CrossChainFee, error) {
	crossChainFees := make([]*xmsgtypes.CrossChainFee, 0)
	totalFee := sdkmath.NewInt(0)

	for _, log := range logs {
		ctx.Logger().Debug("MiddlewareEventHandler: log", "log", log)
		if len(log.Topics) == 0 {
			ctx.Logger().Debug("MiddlewareEventHandler: log has no topics")
			continue
		}

		// Process the log for each subscriber
		for _, subscriber := range h.subscriberRegistry {
			ctx.Logger().Debug("MiddlewareEventHandler: subscriber", "subscriber", reflect.TypeOf(subscriber))

			fee, err := subscriber.ProcessLogs(ctx, 0, toAddress, log, txOrigin)
			if err != nil {
				ctx.Logger().Error("MiddlewareEventHandler: subscriber.ProcessLogs", "error", err, "subscriber", subscriber)
				continue
			}

			if fee != nil {
				totalFee = totalFee.Add(*fee)
			}
		}
	}

	if !totalFee.IsZero() {
		crossChainFees = append(crossChainFees, &xmsgtypes.CrossChainFee{
			Address: sdk.AccAddress(ethcommon.HexToAddress(txOrigin).Bytes()),
			Fee:     totalFee,
		})
	}

	return crossChainFees, nil
}

// GetContractAddress gets the contract address
func (h *MiddlewareEventHandler) GetContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	return ethcommon.Address{}, fmt.Errorf("not implemented")
}

// ParseEvent parses the event
func (h *MiddlewareEventHandler) ParseEvent(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
