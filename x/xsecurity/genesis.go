package observer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/xsecurity/keeper"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// InitGenesis initializes the xsecurity module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
}

// ExportGenesis returns the xsecurity module's exported genesis state as raw JSON bytes.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{}
}
