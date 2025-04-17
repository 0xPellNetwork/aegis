package keeper

import (
	"strconv"

	types2 "github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

const (
	EventTypeMsgServerAddBlameVote = iota
	EventTypeMsgServerVoteTss
	EventTypeVoteInbound
	EventTypeVoteInboundBlock
	EventTypeVoteOnGasTokenRecharge
	EventTypeVoteOnPellTokenRecharge
	EventTypeMsgServerVoteOutboundTx
)

func EmitEventBallotCreated(ctx sdk.Context, ballot types.Ballot, observationHash, observationChain string, eventType uint64) {
	err := ctx.EventManager().EmitTypedEvent(&types.EventBallotCreated{
		BallotIdentifier: ballot.BallotIdentifier,
		BallotType:       ballot.ObservationType.String(),
		ObservationHash:  observationHash,
		ObservationChain: observationChain,
		EventType:        eventType,
	})
	if err != nil {
		ctx.Logger().Error("failed to emit EventBallotCreated : %s", err.Error())
	}
}

func EmitEventKeyGenBlockUpdated(ctx sdk.Context, keygen *types.Keygen) {
	err := ctx.EventManager().EmitTypedEvents(&types.EventKeygenBlockUpdated{
		MsgTypeUrl:    sdk.MsgTypeURL(&types.MsgUpdateKeygen{}),
		KeygenBlock:   strconv.Itoa(int(keygen.BlockNumber)),
		KeygenPubkeys: types2.PrettyPrintStruct(keygen.GranteePubkeys),
	})
	if err != nil {
		ctx.Logger().Error("Error emitting EventKeygenBlockUpdated :", err)
	}
}

func EmitEventAddObserver(ctx sdk.Context, observerCount uint64, operatorAddress, pellclientGranteeAddress, pellclientGranteePubkey string) {
	err := ctx.EventManager().EmitTypedEvents(&types.EventNewRelayerAdded{
		MsgTypeUrl:               sdk.MsgTypeURL(&types.MsgAddObserver{}),
		ObserverAddress:          operatorAddress,
		PellclientGranteeAddress: pellclientGranteeAddress,
		PellclientGranteePubkey:  pellclientGranteePubkey,
		ObserverLastBlockCount:   observerCount,
	})
	if err != nil {
		ctx.Logger().Error("Error emitting EmitEventAddObserver :", err)
	}
}
