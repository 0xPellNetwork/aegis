package keeper

import (
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

var _ types.QueryServer = Keeper{}
