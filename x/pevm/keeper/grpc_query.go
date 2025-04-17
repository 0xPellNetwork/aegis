package keeper

import (
	"github.com/pell-chain/pellcore/x/pevm/types"
)

var _ types.QueryServer = Keeper{}
