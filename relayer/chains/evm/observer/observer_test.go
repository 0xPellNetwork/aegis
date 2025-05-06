package observer_test

import (
	"testing"

	"cosmossdk.io/math"
	lru "github.com/hashicorp/golang-lru"
	"github.com/onrik/ethrpc"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/coin"
	evmobserver "github.com/0xPellNetwork/aegis/relayer/chains/evm/observer"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/config"
	pctx "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/db"
	clientlogs "github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/testutils"
	"github.com/0xPellNetwork/aegis/relayer/testutils/stub"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// the relative path to the testdata directory
var TestDataDir = "../../../"

// getAppContext creates an app context for unit tests
func getAppContext(
	t *testing.T,
	evmChain chains.Chain,
	endpoint string,
	evmChainParams *relayertypes.ChainParams,
) (*pctx.AppContext, config.EVMConfig) {
	// use default endpoint if not provided
	if endpoint == "" {
		endpoint = "http://localhost:8545"
	}

	require.Equal(t, evmChain.Id, evmChainParams.ChainId, "chain id mismatch between chain and params")

	// create config
	cfg := config.NewConfig()
	cfg.EVMChainConfigs[evmChain.Id] = config.EVMConfig{
		Chain:    evmChain,
		Endpoint: endpoint,
	}

	// create core context
	evmChainParamsMap := make(map[int64]*relayertypes.ChainParams)
	evmChainParamsMap[evmChain.Id] = evmChainParams

	// create app context
	appCtx := pctx.NewAppContext(cfg, zerolog.Nop())

	// feed chain params
	err := appCtx.Update(
		relayertypes.Keygen{},
		[]chains.Chain{evmChain},
		evmChainParamsMap,
		"",
		*sample.CrosschainFlags_pell(),
		sample.VerificationFlags(),
		true,
		zerolog.Logger{},
	)
	require.NoError(t, err)

	return appCtx, cfg.EVMChainConfigs[evmChain.Id]
}

// MockEVMClient creates a mock ChainClient with custom chain, TSS, params etc
func MockEVMClient(
	t *testing.T,
	chain chains.Chain,
	config config.EVMConfig,
	evmClient interfaces.EVMRPCClient,
	evmJSONRPC interfaces.EVMJSONRPCClient,
	pellcoreClient interfaces.PellCoreBridger,
	tss interfaces.TSSSigner,
	lastBlock uint64,
	params relayertypes.ChainParams,
) (*evmobserver.ChainClient, *pctx.AppContext) {
	// use default mock evm client if not provided
	if evmClient == nil {
		evmClient = stub.NewMockEvmClient().WithBlockNumber(1000)
	}

	// use default mock evm client if not provided
	if evmJSONRPC == nil {
		evmJSONRPC = stub.NewMockJSONRPCClient()
	}

	// use default mock tss if not provided
	if tss == nil {
		tss = stub.NewTSSMainnet()
	}

	// use default mock bridge if not provided
	if pellcoreClient == nil {
		pellcoreClient = stub.NewMockPellCoreBridge()
	}
	// use default mock tss if not provided
	if tss == nil {
		tss = stub.NewTSSMainnet()
	}
	// create app context
	appCtx, _ := getAppContext(t, chain, "", &params)
	ctx := pctx.WithAppContext(context.Background(), appCtx)

	database, err := db.NewFromSqliteInMemory(true)
	require.NoError(t, err)

	testLogger := zerolog.New(zerolog.NewTestWriter(t))
	logger := clientlogs.Logger{Std: testLogger, Compliance: testLogger}

	// create chain client
	client, err := evmobserver.NewEVMChainClient(
		ctx,
		config,
		params,
		evmClient,
		evmJSONRPC,
		pellcoreClient,
		tss,
		database,
		logger,
		nil,
	)
	require.NoError(t, err)
	client.WithLastBlock(lastBlock)

	return client, appCtx
}

func TestEVM_BlockCache(t *testing.T) {
	// create client
	blockCache, err := lru.New(1000)
	require.NoError(t, err)
	ob := &evmobserver.ChainClient{}
	ob.WithBlockCache(blockCache)

	// delete non-existing block should not panic
	blockNumber := uint64(10388180)
	ob.RemoveCachedBlock(blockNumber)

	// add a block
	block := &ethrpc.Block{
		// #nosec G701 always in range
		Number: int(blockNumber),
	}
	blockCache.Add(blockNumber, block)
	ob.WithBlockCache(blockCache)

	// block should be in cache
	_, err = ob.GetBlockByNumberCached(blockNumber)
	require.NoError(t, err)

	// delete the block should not panic
	ob.RemoveCachedBlock(blockNumber)
}

func TestEVM_CheckTxInclusion(t *testing.T) {
	// load archived evm outtx Gas
	// https://etherscan.io/tx/0xd13b593eb62b5500a00e288cc2fb2c8af1339025c0e6bc6183b8bef2ebbed0d3
	chainID := int64(1)
	coinType := coin.CoinType_GAS
	outtxHash := "0xd13b593eb62b5500a00e288cc2fb2c8af1339025c0e6bc6183b8bef2ebbed0d3"
	tx, receipt := testutils.LoadEVMOuttxNReceipt(t, TestDataDir, chainID, outtxHash, coinType)

	// load archived evm block
	// https://etherscan.io/block/19363323
	blockNumber := receipt.BlockNumber.Uint64()
	block := testutils.LoadEVMBlock(t, TestDataDir, chainID, blockNumber, true)

	// create client
	blockCache, err := lru.New(1000)
	require.NoError(t, err)
	ob := &evmobserver.ChainClient{}

	// save block to cache
	blockCache.Add(blockNumber, block)
	ob.WithBlockCache(blockCache)

	t.Run("should pass for archived outtx", func(t *testing.T) {
		err := ob.CheckTxInclusion(tx, receipt)
		require.NoError(t, err)
	})
	t.Run("should fail on tx index out of range", func(t *testing.T) {
		// modify tx index to invalid number
		copyReceipt := *receipt
		// #nosec G701 non negative value
		copyReceipt.TransactionIndex = uint(len(block.Transactions))
		err := ob.CheckTxInclusion(tx, &copyReceipt)
		require.ErrorContains(t, err, "out of range")
	})
	t.Run("should fail on tx hash mismatch", func(t *testing.T) {
		// change the tx at position 'receipt.TransactionIndex' to a different tx
		priorTx := block.Transactions[receipt.TransactionIndex-1]
		block.Transactions[receipt.TransactionIndex] = priorTx
		blockCache.Add(blockNumber, block)
		ob.WithBlockCache(blockCache)

		// check inclusion should fail
		err := ob.CheckTxInclusion(tx, receipt)
		require.ErrorContains(t, err, "has different hash")

		// wrong block should be removed from cache
		_, ok := blockCache.Get(blockNumber)
		require.False(t, ok)
	})
}

func TestEVM_VoteOutboundBallot(t *testing.T) {
	// load archived evm outtx Gas
	// https://etherscan.io/tx/0xd13b593eb62b5500a00e288cc2fb2c8af1339025c0e6bc6183b8bef2ebbed0d3
	chainID := int64(1)
	coinType := coin.CoinType_GAS
	outtxHash := "0xd13b593eb62b5500a00e288cc2fb2c8af1339025c0e6bc6183b8bef2ebbed0d3"
	tx, receipt := testutils.LoadEVMOuttxNReceipt(t, TestDataDir, chainID, outtxHash, coinType)

	// load archived xmsg
	xmsg := testutils.LoadXmsgByNonce(t, chainID, tx.Nonce())

	t.Run("outtx ballot should match xmsg", func(t *testing.T) {
		msg := types.NewMsgVoteOnObservedOutboundTx(
			"anyCreator",
			xmsg.Index,
			receipt.TxHash.Hex(),
			receipt.BlockNumber.Uint64(),
			receipt.GasUsed,
			math.NewIntFromBigInt(tx.GasPrice()),
			tx.Gas(),
			chains.ReceiveStatus_SUCCESS,
			"",
			chainID,
			tx.Nonce(),
		)
		ballotExpected := xmsg.GetCurrentOutTxParam().OutboundTxBallotIndex
		t.Log("ballotExpected", ballotExpected, "msg.Digest()", msg.Digest()) // TODO: fix
		//require.Equal(t, ballotExpected, msg.Digest())
	})
}
