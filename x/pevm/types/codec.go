package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDeploySystemContracts{}, "pevm/MsgDeploySystemContracts", nil)
	cdc.RegisterConcrete(&MsgUpgradeSystemContracts{}, "pevm/MsgUpgradeSystemContracts", nil)
	cdc.RegisterConcrete(&MsgDeployGatewayContract{}, "pevm/MsgDeployGatewayContract", nil)
	cdc.RegisterConcrete(&MsgDeployConnectorContract{}, "pevm/MsgDeployConnectorContract", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeploySystemContracts{},
		&MsgUpgradeSystemContracts{},
		&MsgDeployGatewayContract{},
		&MsgDeployConnectorContract{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
