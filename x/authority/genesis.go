package authority

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/authority/keeper"
	"github.com/0xPellNetwork/aegis/x/authority/types"
)

// InitGenesis initializes the authority module's state from a provided genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetPolicies(ctx, genState.Policies)
}

// ExportGenesis returns the authority module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var genesis types.GenesisState

	policies, found := k.GetPolicies(ctx)
	if found {
		genesis.Policies = policies
	}

	return &genesis
}
