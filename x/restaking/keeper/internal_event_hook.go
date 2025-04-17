package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ types.InternalEventLogHooks = Hooks{}

type Hooks struct {
	Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{Keeper: k}
}

// HandleEventLogs is a wrapper for calling the EVM PostTxProcessing hook on
func (h Hooks) HandleEventLogs(ctx sdk.Context, emittingContractAddr ethcommon.Address, logs []*ethtypes.Log, txOrigin string) error {
	epochNum := h.GetEpochNumber(ctx)

	for _, handler := range h.eventHandler {
		fees, err := handler.HandleEvent(ctx, epochNum, emittingContractAddr, logs, txOrigin)
		if err != nil {
			return err
		}

		if err := h.xmsgKeeper.DeductFees(ctx, fees); err != nil {
			return err
		}
	}

	return nil
}
