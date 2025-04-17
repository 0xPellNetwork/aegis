package keeper

import (
	"github.com/0xPellNetwork/aegis/x/emissions/types"
)

var _ types.QueryServer = Keeper{}
