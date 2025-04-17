package keeper

import (
	"github.com/pell-chain/pellcore/x/restaking/types"
)

var _ types.QueryServer = Keeper{}
