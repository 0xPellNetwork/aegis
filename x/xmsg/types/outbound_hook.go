package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

type XmsgOutboundResultHook interface {
	ProcessXmsgOutboundResult(ctx sdk.Context, xmsg *Xmsg, ballotStatus relayertypes.BallotStatus)
}
