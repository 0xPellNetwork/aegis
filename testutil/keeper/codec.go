package keeper

import (
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	etherminttypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	emissionstypes "github.com/0xPellNetwork/aegis/x/emissions/types"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func NewCodec() *codec.ProtoCodec {
	registry := codectypes.NewInterfaceRegistry()

	cryptocodec.RegisterInterfaces(registry)
	authtypes.RegisterInterfaces(registry)
	authz.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
	slashingtypes.RegisterInterfaces(registry)
	upgradetypes.RegisterInterfaces(registry)
	distrtypes.RegisterInterfaces(registry)
	evidencetypes.RegisterInterfaces(registry)
	crisistypes.RegisterInterfaces(registry)
	evmtypes.RegisterInterfaces(registry)
	etherminttypes.RegisterInterfaces(registry)
	xmsgtypes.RegisterInterfaces(registry)
	emissionstypes.RegisterInterfaces(registry)
	pevmtypes.RegisterInterfaces(registry)
	relayertypes.RegisterInterfaces(registry)

	cdc := codec.NewProtoCodec(registry)
	return cdc
}
