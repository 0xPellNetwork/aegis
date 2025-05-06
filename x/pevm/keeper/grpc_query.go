package keeper

import (
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

var _ types.QueryServer = Keeper{}
