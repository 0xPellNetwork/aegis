package keeper

import (
	"github.com/0xPellNetwork/aegis/x/authority/types"
)

var _ types.QueryServer = Keeper{}
