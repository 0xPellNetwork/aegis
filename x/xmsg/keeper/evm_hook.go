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

type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k: k}
}

// evm hooks -----------------------------------------------------------------------------------------------------
// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.k.Logger(ctx).Info("PostTxProcessing", "xmsgIndex", inXmsgIndices, "msg", msg, "receipt", receipt)

	// handle event logs from this module
	for _, handler := range h.k.internalHandlers {
		crossChainFees, err := handler.HandleEvent(ctx, 0, *msg.To(), receipt.Logs, msg.From().Hex())
		if err != nil {
			return err
		}

		if err := h.k.DeductFees(ctx, crossChainFees); err != nil {
			return err
		}
	}

	return nil
}
