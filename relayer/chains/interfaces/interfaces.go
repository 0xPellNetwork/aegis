package interfaces

import (
	"context"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/onrik/ethrpc"
	"github.com/rs/zerolog"
	"gitlab.com/thorchain/tss/go-tss/blame"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	keyinterfaces "github.com/0xPellNetwork/aegis/relayer/keys/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/outtxprocessor"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

type Order string

const (
	NoOrder    Order = ""
	Ascending  Order = "ASC"
	Descending Order = "DESC"
)

// ChainClient is the interface for chain clients
type ChainClient interface {
	Start(ctx context.Context)
	Stop()
	IsOutboundProcessed(ctx context.Context, xmsg *xmsgtypes.Xmsg, logger zerolog.Logger) (bool, bool, error)
	SetChainParams(relayertypes.ChainParams)
	GetChainParams() relayertypes.ChainParams
	OutboundID(nonce uint64) string
	WatchIntxTracker(ctx context.Context) error
}

// ChainSigner is the interface to sign transactions for a chain
type ChainSigner interface {
	TryProcessOutTx(
		ctx context.Context,
		xmsg *xmsgtypes.Xmsg,
		outTxMan *outtxprocessor.Processor,
		outTxID string,
		chainclient ChainClient,
		pellBridge PellCoreBridger,
		height uint64,
	)
	SetPellConnectorAddress(address ethcommon.Address)
	GetPellConnectorAddress() ethcommon.Address
}

// PellcoreVoter represents voter interface.
type PellcoreVoter interface {
	PostVoteBlockHeader(
		ctx context.Context,
		chainID int64,
		txhash []byte,
		height int64,
		header proofs.HeaderData,
	) (string, error)
	PostGasPrice(
		ctx context.Context,
		chain chains.Chain,
		gasPrice uint64,
		supply string,
		blockNum uint64,
	) (string, error)
	PostVoteInboundEvents(
		ctx context.Context,
		gasLimit, retryGasLimit uint64,
		msg []*xmsgtypes.MsgVoteOnObservedInboundTx,
	) (string, string, error)
	PostVoteOutbound(
		ctx context.Context,
		sendHash string,
		outTxHash string,
		outBlockHeight uint64,
		outTxGasUsed uint64,
		outTxEffectiveGasPrice *big.Int,
		outTxEffectiveGasLimit uint64,
		status chains.ReceiveStatus,
		failedReasonMsg string,
		chain chains.Chain,
		nonce uint64,
	) (string, string, error)
	PostBlameData(
		ctx context.Context,
		blame *blame.Blame,
		chainID int64,
		index string,
	) (string, error)
	PostAddTxHashToOutTxTracker(
		ctx context.Context,
		chainID int64,
		nonce uint64,
		txHash string,
		proof *proofs.Proof,
		blockHash string,
		txIndex int64,
	) (string, error)
	PostVoteOnPellRecharge(
		ctx context.Context,
		chain chains.Chain,
		voteIndex uint64,
	) (string, error)
	PostVoteOnGasRecharge(
		ctx context.Context,
		chain chains.Chain,
		voteIndex uint64,
	) (string, error)
	PostVoteInboundBlock(
		ctx context.Context,
		gasLimit, retryLimit uint64,
		block *xmsgtypes.MsgVoteInboundBlock,
		events []*xmsgtypes.MsgVoteOnObservedInboundTx,
	) ([]string, []string, error)
}

// PellCoreBridger is the interface to interact with PellCore
type PellCoreBridger interface {
	PellcoreVoter

	Chain() chains.Chain
	GetLogger() *zerolog.Logger
	GetKeys() keyinterfaces.ObserverKeys

	GetKeyGen(ctx context.Context) (relayertypes.Keygen, error)

	GetBlockHeight(ctx context.Context) (int64, error)
	GetLastBlockHeightByChain(ctx context.Context, chain chains.Chain) (*xmsgtypes.LastBlockHeight, error)
	GetBlockHeaderChainState(ctx context.Context, chainID int64) (*lightclienttypes.QueryChainStateResponse, error)

	ListPendingXmsg(ctx context.Context, chainID int64) ([]*xmsgtypes.Xmsg, uint64, error)
	ListPendingXmsgWithinRatelimit(ctx context.Context) (*xmsgtypes.QueryListPendingXmsgWithinRateLimitResponse, error)
	GetRateLimiterInput(ctx context.Context, window int64) (*xmsgtypes.QueryRateLimiterInputResponse, error)
	GetPendingNoncesByChain(ctx context.Context, chainID int64) (relayertypes.PendingNonces, error)

	GetXmsgByNonce(ctx context.Context, chainID int64, nonce uint64) (*xmsgtypes.Xmsg, error)
	GetOutTxTracker(ctx context.Context, chain chains.Chain, nonce uint64) (*xmsgtypes.OutTxTracker, error)
	GetAllOutTxTrackerByChain(ctx context.Context, chainID int64, order Order) ([]xmsgtypes.OutTxTracker, error)
	GetCrosschainFlags(ctx context.Context) (relayertypes.CrosschainFlags, error)
	GetRateLimiterFlags(ctx context.Context) (xmsgtypes.RateLimiterFlags, error)
	GetObserverList(ctx context.Context) ([]string, error)
	GetPellHotKeyBalance(ctx context.Context) (sdkmath.Int, error)
	GetInboundTrackersForChain(ctx context.Context, chainID int64) ([]xmsgtypes.InTxTracker, error)
	GetChainIndex(ctx context.Context, chainId int64) (*xmsgtypes.ChainIndex, error)
	GetPellRechargeOperationIndex(ctx context.Context, chainId int64) (*xmsgtypes.PellRechargeOperationIndex, error)
	GetGasRechargeOperationIndex(ctx context.Context, chainId int64) (*xmsgtypes.GasRechargeOperationIndex, error)

	Stop()
	OnBeforeStop(callback func())
}

// BTCRPCClient is the interface for BTC RPC client
type BTCRPCClient interface {
	GetNetworkInfo() (*btcjson.GetNetworkInfoResult, error)
	CreateWallet(name string, opts ...rpcclient.CreateWalletOpt) (*btcjson.CreateWalletResult, error)
	GetNewAddress(account string) (btcutil.Address, error)
	GenerateToAddress(numBlocks int64, address btcutil.Address, maxTries *int64) ([]*chainhash.Hash, error)
	GetBalance(account string) (btcutil.Amount, error)
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
	ListUnspent() ([]btcjson.ListUnspentResult, error)
	ListUnspentMinMaxAddresses(minConf int, maxConf int, addrs []btcutil.Address) ([]btcjson.ListUnspentResult, error)
	EstimateSmartFee(confTarget int64, mode *btcjson.EstimateSmartFeeMode) (*btcjson.EstimateSmartFeeResult, error)
	GetTransaction(txHash *chainhash.Hash) (*btcjson.GetTransactionResult, error)
	GetRawTransaction(txHash *chainhash.Hash) (*btcutil.Tx, error)
	GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error)
	GetBlockCount() (int64, error)
	GetBlockHash(blockHeight int64) (*chainhash.Hash, error)
	GetBlockVerbose(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error)
	GetBlockVerboseTx(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseTxResult, error)
	GetBlockHeader(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
}

// EVMRPCClient is the interface for EVM RPC client
type EVMRPCClient interface {
	bind.ContractBackend
	SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethtypes.Block, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error)
	TransactionByHash(ctx context.Context, hash ethcommon.Hash) (tx *ethtypes.Transaction, isPending bool, err error)
	TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*ethtypes.Receipt, error)
	TransactionSender(ctx context.Context, tx *ethtypes.Transaction, block ethcommon.Hash, index uint) (ethcommon.Address, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	BalanceAt(ctx context.Context, account ethcommon.Address, blockNumber *big.Int) (*big.Int, error)
}

// EVMJSONRPCClient is the interface for EVM JSON RPC client
type EVMJSONRPCClient interface {
	EthGetBlockByNumber(number int, withTransactions bool) (*ethrpc.Block, error)
	EthGetTransactionByHash(hash string) (*ethrpc.Transaction, error)
}

type TSSSigner interface {
	Pubkey() []byte
	// Sign: Specify optionalPubkey to use a different pubkey than the current pubkey set during keygen
	Sign(ctx context.Context, data []byte, height uint64, nonce uint64, chain *chains.Chain, optionalPubkey string) ([65]byte, error)
	EVMAddress() ethcommon.Address
	PubKeyCompressedBytes() []byte
}

// ChainEventHandler defines an interface for handling chain events,
// such as scanning blockchain events within specific block ranges
// and managing event storage.
//
// Methods:
//   - Execute: Processes blockchain events within a block range.
type ChainEventHandler interface {
	// HandleBlocks scans blockchain events within the specified block range
	// and returns the highest block height processed.
	//
	// Parameters:
	//   - ctx: Context for managing request-scoped values, deadlines, and cancellations.
	//   - startBlock: The starting block height for scanning.
	//   - toBlock: The ending block height for scanning.
	//   - eventStore: A pointer to a map where the key is the block height (int64),
	//     and the value is a slice of `MsgVoteOnObservedInboundTx` messages observed at that block.
	//
	// Returns:
	//   - uint64: The highest block height scanned.
	//   - error: An error, if any occurs during execution.
	HandleBlocks(startBlock, toBlock uint64, eventStore *map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) (uint64, error)
	// CheckAndBuildInboundVoteMsg validates and builds an inbound vote message.
	//
	// Parameters:
	//   - ctx: Context for managing request-scoped values, deadlines, and cancellations.
	//   - tx: The Ethereum transaction to validate.
	//   - receipt: The Ethereum transaction receipt to validate.
	//
	// Returns:
	//   - []*xmsgtypes.MsgVoteOnObservedInboundTx: A slice of `MsgVoteOnObservedInboundTx` messages.
	//   - error: An error, if any occurs during execution.
	CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error)
}

// IEVMEventReactor defines an interface for processing events within a specified block range
// in an EVM-compatible blockchain. The main purpose is to handle all relevant events
// between two block heights.
//
// Methods:
//   - HandleBlocks: Processes all events from the given start block to the end block.
type IEVMEventReactor interface {
	// HandleBlocks processes all events from the given start block to the end block.
	//
	// Parameters:
	//   - startBlock: The starting block height for scanning.
	//   - toBlock: The ending block height for scanning.
	//
	// Returns:
	//   - map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx: A map where the key is the block height (uint64),
	//     and the value is a slice of `MsgVoteOnObservedInboundTx` messages observed at that block.
	//   - uint64: The highest block height scanned.
	HandleBlocks(startBlock, toBlock uint64) (map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx, uint64)
	// RegisterEventHandler registers a chain event handler to the reactor.
	//
	// Parameters:
	//   - handler: The chain event handler to register.
	RegisterEventHandler(handler ChainEventHandler)
	// CheckAndBuildInboundVoteMsg validates and builds an inbound vote message.
	//
	// Parameters:
	//   - ctx: Context for managing request-scoped values, deadlines, and cancellations.
	//   - tx: The Ethereum transaction to validate.
	//   - receipt: The Ethereum transaction receipt to validate.
	//   - lastBlock: The last block height.
	//
	// Returns:
	//   - []*xmsgtypes.MsgVoteOnObservedInboundTx: A slice of `MsgVoteOnObservedInboundTx` messages.
	//   - error: An error, if any occurs during execution.
	CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt, lastBlock uint64) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error)
}
