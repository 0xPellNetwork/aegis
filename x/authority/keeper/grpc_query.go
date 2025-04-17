package keeper

import (
	"github.com/pell-chain/pellcore/x/authority/types"
)

var _ types.QueryServer = Keeper{}
