package keeper

import (
	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

var _ types.QueryServer = Keeper{}
