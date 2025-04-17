package keeper

import (
	"github.com/pell-chain/pellcore/x/lightclient/types"
)

var _ types.QueryServer = Keeper{}
