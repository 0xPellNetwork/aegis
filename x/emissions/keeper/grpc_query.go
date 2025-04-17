package keeper

import (
	"github.com/pell-chain/pellcore/x/emissions/types"
)

var _ types.QueryServer = Keeper{}
