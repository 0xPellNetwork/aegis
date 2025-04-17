package types

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

type RelayerKeeper interface {
	GetMaturedBallotList(ctx sdk.Context) []string
	GetBallot(ctx sdk.Context, index string) (val relayertypes.Ballot, found bool)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// Methods imported from bank should be defined here
}

type StakingKeeper interface {
	BondedRatio(ctx context.Context) (sdkmath.LegacyDec, error)
}

// ParamStore defines the expected paramstore methods to store and load Params (noalias)
type ParamStore interface {
	GetParamSetIfExists(ctx sdk.Context, ps paramstypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramstypes.ParamSet)
	WithKeyTable(table paramstypes.KeyTable) paramstypes.Subspace
	HasKeyTable() bool
}
