package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddObserver{}, "relayer/AddObserver", nil)
	cdc.RegisterConcrete(&MsgUpsertChainParams{}, "relayer/UpdateChainParams", nil)
	cdc.RegisterConcrete(&MsgRemoveChainParams{}, "relayer/RemoveChainParams", nil)
	cdc.RegisterConcrete(&MsgVoteBlockHeader{}, "relayer/VoteBlockHeader", nil)
	cdc.RegisterConcrete(&MsgAddBlameVote{}, "relayer/AddBlameVote", nil)
	cdc.RegisterConcrete(&MsgUpsertCrosschainFlags{}, "relayer/UpdateCrosschainFlags", nil)
	cdc.RegisterConcrete(&MsgUpdateKeygen{}, "relayer/UpdateKeygen", nil)
	cdc.RegisterConcrete(&MsgUpdateObserver{}, "relayer/UpdateObserver", nil)
	cdc.RegisterConcrete(&MsgResetChainNonces{}, "relayer/ResetChainNonces", nil)
	cdc.RegisterConcrete(&MsgVoteTSS{}, "relayer/VoteTSS", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddObserver{},
		&MsgUpsertChainParams{},
		&MsgRemoveChainParams{},
		&MsgAddBlameVote{},
		&MsgUpsertCrosschainFlags{},
		&MsgUpdateKeygen{},
		&MsgVoteBlockHeader{},
		&MsgUpdateObserver{},
		&MsgResetChainNonces{},
		&MsgVoteTSS{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
