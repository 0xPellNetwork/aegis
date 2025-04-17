package keeper

import (
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ types.QueryServer = Keeper{}
