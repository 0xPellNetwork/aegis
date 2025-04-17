package keeper

import (
	"github.com/pell-chain/pellcore/x/relayer/types"
)

var _ types.QueryServer = Keeper{}
