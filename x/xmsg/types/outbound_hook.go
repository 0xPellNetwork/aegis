package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
)

type XmsgOutboundResultHook interface {
	ProcessXmsgOutboundResult(ctx sdk.Context, xmsg *Xmsg, ballotStatus relayertypes.BallotStatus)
}
