package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	_ evmtypes.EvmHooks = Hooks{}
)

// evm hooks -----------------------------------------------------------------------------------------------------
// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	h.Logger(ctx).Info("PostTxProcessing start", "msg", msg, "receipt", receipt)

	// handle event logs from other modules
	for _, handler := range h.eventHandler {
		crossChainFees, err := handler.HandleEvent(ctx, 0, *msg.To(), receipt.Logs, msg.From().Hex())
		if err != nil {
			return err
		}

		if err := h.xmsgKeeper.DeductFees(ctx, crossChainFees); err != nil {
			return err
		}
	}

	return nil
}
