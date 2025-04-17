package types

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/coin"
	"github.com/pell-chain/pellcore/pkg/proofs"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
)

type StakingKeeper interface {
	GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type RelayerKeeper interface {
	SetChainParamsList(ctx sdk.Context, chainParams relayertypes.ChainParamsList)
	GetObserverSet(ctx context.Context) (val relayertypes.RelayerSet, found bool)
	GetBallot(ctx sdk.Context, index string) (val relayertypes.Ballot, found bool)
	GetChainParamsByChainID(ctx sdk.Context, chainID int64) (params *relayertypes.ChainParams, found bool)
	GetNodeAccount(ctx sdk.Context, address string) (nodeAccount relayertypes.NodeAccount, found bool)
	GetAllNodeAccount(ctx sdk.Context) (nodeAccounts []relayertypes.NodeAccount)
	SetNodeAccount(ctx sdk.Context, nodeAccount relayertypes.NodeAccount)
	IsInboundEnabled(ctx sdk.Context) (found bool)
	GetCrosschainFlags(ctx sdk.Context) (val relayertypes.CrosschainFlags, found bool)
	GetKeygen(ctx sdk.Context) (val relayertypes.Keygen, found bool)
	SetKeygen(ctx sdk.Context, keygen relayertypes.Keygen)
	SetCrosschainFlags(ctx sdk.Context, crosschainFlags relayertypes.CrosschainFlags)
	SetLastObserverCount(ctx sdk.Context, lbc *relayertypes.LastRelayerCount)
	AddVoteToBallot(ctx sdk.Context, ballot relayertypes.Ballot, address string, observationType relayertypes.VoteType) (relayertypes.Ballot, error)
	CheckIfFinalizingVote(ctx sdk.Context, ballot relayertypes.Ballot) (relayertypes.Ballot, bool)
	IsNonTombstonedObserver(ctx sdk.Context, address string) bool
	FindBallot(ctx sdk.Context, index string, chain *chains.Chain, observationType relayertypes.ObservationType) (ballot relayertypes.Ballot, isNew bool, err error)
	AddBallotToList(ctx sdk.Context, ballot relayertypes.Ballot)
	CheckIfTssPubkeyHasBeenGenerated(ctx sdk.Context, tssPubkey string) (relayertypes.TSS, bool)
	GetAllTSS(ctx sdk.Context) (list []relayertypes.TSS)
	GetTSS(ctx sdk.Context) (val relayertypes.TSS, found bool)
	SetTSS(ctx sdk.Context, tss relayertypes.TSS)
	SetTSSHistory(ctx sdk.Context, tss relayertypes.TSS)
	GetTssAddress(goCtx context.Context, req *relayertypes.QueryGetTssAddressRequest) (*relayertypes.QueryGetTssAddressResponse, error)
	GetChainParamsList(ctx sdk.Context) (val relayertypes.ChainParamsList, found bool)
	SetFundMigrator(ctx sdk.Context, fm relayertypes.TssFundMigratorInfo)
	GetFundMigrator(ctx sdk.Context, chainID int64) (val relayertypes.TssFundMigratorInfo, found bool)
	GetAllTssFundMigrators(ctx sdk.Context) (fms []relayertypes.TssFundMigratorInfo)
	RemoveAllExistingMigrators(ctx sdk.Context)
	SetChainNonces(ctx sdk.Context, chainNonces relayertypes.ChainNonces)
	GetChainNonces(ctx sdk.Context, index string) (val relayertypes.ChainNonces, found bool)
	GetAllChainNonces(ctx sdk.Context) (list []relayertypes.ChainNonces)
	SetNonceToXmsg(ctx sdk.Context, nonceToXmsg relayertypes.NonceToXmsg)
	GetNonceToXmsg(ctx sdk.Context, tss string, chainID int64, nonce int64) (val relayertypes.NonceToXmsg, found bool)
	GetAllPendingNonces(ctx sdk.Context) (list []relayertypes.PendingNonces, err error)
	GetPendingNonces(ctx sdk.Context, tss string, chainID int64) (val relayertypes.PendingNonces, found bool)
	SetPendingNonces(ctx sdk.Context, pendingNonces relayertypes.PendingNonces)
	SetTssAndUpdateNonce(ctx sdk.Context, tss relayertypes.TSS)
	RemoveFromPendingNonces(ctx sdk.Context, tss string, chainID int64, nonce int64)
	GetAllNonceToXmsg(ctx sdk.Context) (list []relayertypes.NonceToXmsg)
	VoteOnInboundBallot(
		ctx sdk.Context,
		senderChainID int64,
		receiverChainID int64,
		coinType coin.CoinType,
		voter string,
		ballotIndex string,
		inTxHash string,
	) (bool, bool, error)
	VoteOnInboundBlockBallot(
		ctx sdk.Context,
		chainId int64,
		voter string,
		ballotIndex string,
		blockHash string,
	) (bool, bool, error)
	VoteOnOutboundBallot(
		ctx sdk.Context,
		ballotIndex string,
		outTxChainID int64,
		receiveStatus chains.ReceiveStatus,
		voter string,
	) (bool, bool, relayertypes.Ballot, string, error)
	VoteOnAddPellTokenBallot(
		ctx sdk.Context,
		chainId int64,
		voter string,
		voteIndex uint64,
	) (bool, bool, error)
	VoteOnAddGasTokenBallot(
		ctx sdk.Context,
		chainId int64,
		voter string,
		voteIndex uint64,
	) (bool, bool, error)
	GetSupportedChainFromChainID(ctx sdk.Context, chainID int64) *chains.Chain
	GetSupportedChains(ctx sdk.Context) []*chains.Chain
	GetSupportedForeignChains(ctx sdk.Context) []*chains.Chain
	// TODO: remove this after the next upgrade
	DeleteBallot(ctx sdk.Context, index string)
}

type PevmKeeper interface {
	GetSystemContract(ctx sdk.Context) (val pevmtypes.SystemContract, found bool)
	PELLRevertAndCallContract(ctx sdk.Context,
		sender ethcommon.Address,
		to ethcommon.Address,
		inboundSenderChainID int64,
		destinationChainID int64,
		indexBytes [32]byte) (*evmtypes.MsgEthereumTxResponse, error)
	CallSyncDepositStateOnPellStrategyManager(
		ctx context.Context,
		from []byte,
		senderChainID int64,
		staker,
		strategy ethcommon.Address,
		shares *big.Int,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallSyncDelegatedStateOnPellDelegationManager(
		ctx context.Context,
		from []byte,
		senderChainID int64,
		staker,
		strategy ethcommon.Address,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallSyncWithdrawalStateOnPellDelegationManager(
		ctx context.Context,
		senderChainID int64,
		staker ethcommon.Address,
		withdrawalParam *WithdrawalQueued,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallSyncUndelegateStateOnPellDelegationManager(
		ctx context.Context,
		senderChainID int64,
		staker ethcommon.Address,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallBridgePellOnPellGateway(
		ctx context.Context,
		destinationChainId int64,
		receiver ethcommon.Address,
		amount *big.Int,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallSwapOnPellGasSwap(
		ctx context.Context,
		destinationChainId int64,
		amountIn *big.Int,
		receiver ethcommon.Address,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	GetPellConnectorContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	GetPellStrategyManagerProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	GetPellDelegationManagerProxyContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	GetPellGatewayEVMContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	GetGasSwapPEVMContractAddress(ctx sdk.Context) (ethcommon.Address, error)
	CallAddSupportedChainOnRegistryRouter(
		ctx sdk.Context,
		params *RegisterChainDVSToPell,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallProcessPellSent(ctx sdk.Context, action *PellSent, xmsgIndex string) (*evmtypes.MsgEthereumTxResponse, bool, error)
}

type AuthorityKeeper interface {
	IsAuthorized(ctx sdk.Context, address string, policyType authoritytypes.PolicyType) bool
}

type LightclientKeeper interface {
	VerifyProof(ctx sdk.Context, proof *proofs.Proof, chainID int64, blockHash string, txIndex int64) ([]byte, error)
}
