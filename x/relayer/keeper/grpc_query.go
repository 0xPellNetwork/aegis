package keeper

import (
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

var _ types.QueryServer = Keeper{}
