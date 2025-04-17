package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddToOutTxTracker{}, "xmsg/AddToOutTxTracker", nil)
	cdc.RegisterConcrete(&MsgAddToInTxTracker{}, "xmsg/AddToInTxTracker", nil)
	cdc.RegisterConcrete(&MsgRemoveFromOutTxTracker{}, "xmsg/RemoveFromOutTxTracker", nil)
	cdc.RegisterConcrete(&MsgVoteGasPrice{}, "xmsg/VoteGasPrice", nil)
	cdc.RegisterConcrete(&MsgVoteOnObservedOutboundTx{}, "xmsg/VoteOnObservedOutboundTx", nil)
	cdc.RegisterConcrete(&MsgVoteOnObservedInboundTx{}, "xmsg/VoteOnObservedInboundTx", nil)
	cdc.RegisterConcrete(&MsgVoteInboundBlock{}, "xmsg/VoteInboundBlock", nil)
	cdc.RegisterConcrete(&MsgMigrateTssFunds{}, "xmsg/MigrateTssFunds", nil)
	cdc.RegisterConcrete(&MsgUpdateTssAddress{}, "xmsg/UpdateTssAddress", nil)
	cdc.RegisterConcrete(&MsgAbortStuckXmsg{}, "xmsg/AbortStuckXmsg", nil)
	cdc.RegisterConcrete(&MsgUpdateRateLimiterFlags{}, "xmsg/UpdateRateLimiterFlags", nil)
	cdc.RegisterConcrete(&MsgUpsertCrosschainFeeParams{}, "xmsg/UpsertCrosschainFeeParams", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddToOutTxTracker{},
		&MsgAddToInTxTracker{},
		&MsgRemoveFromOutTxTracker{},
		&MsgVoteGasPrice{},
		&MsgVoteOnObservedOutboundTx{},
		&MsgVoteOnObservedInboundTx{},
		&MsgMigrateTssFunds{},
		&MsgUpdateTssAddress{},
		&MsgAbortStuckXmsg{},
		&MsgUpdateRateLimiterFlags{},
		&MsgVoteInboundBlock{},
		&MsgUpsertCrosschainFeeParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
