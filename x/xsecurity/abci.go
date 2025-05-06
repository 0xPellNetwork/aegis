package observer

import (
	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/keeper"
)

func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
}

// EndBlocker called at every block, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) ([]types.ValidatorUpdate, error) {
	return k.ProcessEpoch(ctx)
}
