package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddPools{}, "xsecurity/MsgAddPools", nil)
	cdc.RegisterConcrete(&MsgRemovePools{}, "xsecurity/MsgRemovePools", nil)
	cdc.RegisterConcrete(&MsgCreateGroup{}, "xsecurity/MsgCreateGroup", nil)
	cdc.RegisterConcrete(&MsgCreateRegistryRouter{}, "xsecurity/MsgCreateRegistryRouter", nil)
	cdc.RegisterConcrete(&MsgRegisterOperator{}, "xsecurity/MsgRegisterOperator", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddPools{},
		&MsgRemovePools{},
		&MsgCreateGroup{},
		&MsgCreateRegistryRouter{},
		&MsgRegisterOperator{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
