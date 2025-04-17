package types

// DONTCOVER

import (
	cosmoserrors "cosmossdk.io/errors"
)

// x/pevm module sentinel errors
var (
	ErrDataEmpty          = cosmoserrors.Register(ModuleName, 1101, "data empty")
	ErrInvalidDenominator = cosmoserrors.Register(ModuleName, 1102, "invalid denominator")
)
