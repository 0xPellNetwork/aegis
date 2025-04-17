package keeper

import (
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

var _ types.QueryServer = Keeper{}
