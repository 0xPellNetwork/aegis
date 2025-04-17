package types

import (
	"context"
	"math/big"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type RelayerKeeper interface {
	SetPevmKeeper(pevmKeeper relayertypes.PevmKeeper)
	GetSupportedChains(ctx sdk.Context) []*chains.Chain
	GetTSS(ctx sdk.Context) (tss relayertypes.TSS, found bool)
	GetSupportedChainFromChainID(ctx sdk.Context, chainID int64) *chains.Chain
	GetChainParamsByChainID(ctx sdk.Context, chainID int64) (*relayertypes.ChainParams, bool)
	IsInboundEnabled(ctx sdk.Context) (found bool)
}

type EVMKeeper interface {
	ChainID() *big.Int
	GetBlockBloomTransient(ctx sdk.Context) *big.Int
	GetLogSizeTransient(ctx sdk.Context) uint64
	WithChainID(ctx sdk.Context)
	SetBlockBloomTransient(ctx sdk.Context, bloom *big.Int)
	SetLogSizeTransient(ctx sdk.Context, logSize uint64)
	EstimateGas(c context.Context, req *evmtypes.EthCallRequest) (*evmtypes.EstimateGasResponse, error)
	ApplyMessage(
		ctx sdk.Context,
		msg core.Message,
		tracer vm.EVMLogger,
		commit bool,
	) (*evmtypes.MsgEthereumTxResponse, error)
	GetAccount(ctx sdk.Context, addr ethcommon.Address) *statedb.Account
	GetCode(ctx sdk.Context, codeHash ethcommon.Hash) []byte
	SetAccount(ctx sdk.Context, addr ethcommon.Address, account statedb.Account) error
}

type AuthorityKeeper interface {
	IsAuthorized(ctx sdk.Context, address string, policyType authoritytypes.PolicyType) bool
}

type PevmKeeper interface {
	GetPellDelegationManagerProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	CallUpdateDestinationAddressOnPellGateway(
		ctx context.Context,
		chainId int64,
		destinationAddress string,
	) error
	CallUpdateSourceAddressOnPellGateway(
		ctx context.Context,
		chainId int64,
		sourceAddress string,
	) error
	CallUpdateDestinationAddressOnGasSwapPEVM(
		ctx context.Context,
		chainId int64,
		destinationAddress string,
	) error
}

type XmsgKeeper interface {
	GetGasPrice(ctx sdk.Context, chainID int64) (val xmsgtypes.GasPrice, found bool)
	ProcessXmsg(ctx sdk.Context, xmsg xmsgtypes.Xmsg, receiverChain *chains.Chain) error
	ProcessPellSentEvent(
		ctx sdk.Context,
		event *pellconnector.PellConnectorPellSent,
		emittingContract ethcommon.Address,
		txOrigin string,
		tss observertypes.TSS,
	) error
	IsAllowedXmsgSender(ctx sdk.Context, builder string) bool
	XmsgAll(c context.Context, req *xmsgtypes.QueryAllXmsgRequest) (*xmsgtypes.QueryXmsgAllResponse, error)
	GetAllCrosschainEventFees(ctx sdk.Context) ([]xmsgtypes.CrosschainFeeParam, error)
	GetCrosschainEventFee(ctx sdk.Context, chainId int64) (xmsgtypes.CrosschainFeeParam, bool)
	DeductFees(ctx sdk.Context, fees []*xmsgtypes.CrossChainFee) error
}
