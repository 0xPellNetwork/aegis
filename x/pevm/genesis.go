package pevm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/pevm/keeper"
	"github.com/pell-chain/pellcore/x/pevm/types"
)

// InitGenesis initializes the pevm module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if genState.SystemContract != nil {
		k.SetSystemContract(ctx, *genState.SystemContract)
	}

}

// ExportGenesis returns the pevm module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var genesis types.GenesisState

	system, found := k.GetSystemContract(ctx)
	if found {
		genesis.SystemContract = &system
	}

	return &genesis
}
