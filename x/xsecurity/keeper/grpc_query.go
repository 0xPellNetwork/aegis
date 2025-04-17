package keeper

import (
	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

var _ types.QueryServer = Keeper{}
