package keeper

import (
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

var _ types.QueryServer = Keeper{}
