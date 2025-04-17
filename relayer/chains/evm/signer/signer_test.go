package signer

import (
	goctx "context"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/evm/observer"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/config"
	pctx "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/db"
	"github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
	"github.com/0xPellNetwork/aegis/relayer/outtxprocessor"
	"github.com/0xPellNetwork/aegis/relayer/testutils"
	"github.com/0xPellNetwork/aegis/relayer/testutils/stub"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var (
	// Dummy addresses as they are just used as transaction data to be signed
	ConnectorAddress = sample.EthAddress()
)

func getAppcontext() goctx.Context {
	app := pctx.NewAppContext(config.New(), zerolog.Nop())
	ctx := pctx.WithAppContext(goctx.Background(), app)

	return ctx
}

func getNewEvmSigner(tss interfaces.TSSSigner) (*Signer, error) {
	ctx := getAppcontext()

	// use default mock TSS if not provided
	if tss == nil {
		tss = stub.NewTSSMainnet()
	}

	connectorAddress := ConnectorAddress
	logger := logs.Logger{}

	return NewEVMSigner(
		ctx,
		chains.BscMainnetChain(),
		stub.EVMRPCEnabled,
		tss,
		connectorAddress,
		logger,
		nil,
	)
}

// getNewEvmChainObserver creates a new EVM chain observer for testing
func getNewEvmChainClient(t *testing.T, tss interfaces.TSSSigner) (*observer.ChainClient, error) {
	ctx := getAppcontext()

	// use default mock TSS if not provided
	if tss == nil {
		tss = stub.NewTSSMainnet()
	}

	// prepare mock arguments to create observer
	evmClient := stub.NewMockEvmClient().WithBlockNumber(1000)
	evmJSONRPCClient := stub.NewMockJSONRPCClient()
	params := stub.MockChainParams(chains.BscMainnetChain().Id, 10)
	logger := logs.Logger{}
	ts := &metrics.TelemetryServer{}

	database, err := db.NewFromSqliteInMemory(true)
	require.NoError(t, err)

	return observer.NewEVMChainClient(
		ctx,
		config.EVMConfig{},
		params,
		evmClient,
		evmJSONRPCClient,
		stub.NewMockPellCoreBridge(),
		tss,
		database,
		logger,
		ts,
	)
}

func getNewOutTxProcessor() *outtxprocessor.Processor {
	logger := zerolog.Logger{}
	return outtxprocessor.NewOutTxProcessor(logger)
}

func getXmsg(t *testing.T) *xmsgtypes.Xmsg {
	return testutils.LoadXmsgByNonce(t, 56, 68270)
}

func getInvalidXmsg(t *testing.T) *xmsgtypes.Xmsg {
	xmsg := getXmsg(t)
	// modify receiver chain id to make it invalid
	xmsg.GetCurrentOutTxParam().ReceiverChainId = 13378337
	return xmsg
}

// verifyTxSignature is a helper function to verify the signature of a transaction
func verifyTxSignature(t *testing.T, tx *ethtypes.Transaction, tssPubkey []byte, signer ethtypes.Signer) {
	_, r, s := tx.RawSignatureValues()
	signature := append(r.Bytes(), s.Bytes()...)
	hash := signer.Hash(tx)

	verified := crypto.VerifySignature(tssPubkey, hash.Bytes(), signature)
	require.True(t, verified)
}

// verifyTxBodyBasics is a helper function to verify 'to' and 'nonce' of a transaction
func verifyTxBodyBasics(
	t *testing.T,
	tx *ethtypes.Transaction,
	to ethcommon.Address,
	nonce uint64,
) {
	require.Equal(t, to, *tx.To())
	require.Equal(t, nonce, tx.Nonce())
}

// func TestSigner_SetGetConnectorAddress(t *testing.T) {
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)
// 	// Get and compare
// 	require.Equal(t, ConnectorAddress, evmSigner.GetPellConnectorAddress())

// 	// Update and get again
// 	newConnector := sample.EthAddress()
// 	evmSigner.SetPellConnectorAddress(newConnector)
// 	require.Equal(t, newConnector, evmSigner.GetPellConnectorAddress())
// }

// func TestSigner_TryProcessOutTx(t *testing.T) {
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)
// 	xmsg := getXmsg(t)
// 	processorManager := getNewOutTxProcessor()
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)

// 	evmSigner.TryProcessOutTx(xmsg, processorManager, "123", mockChainClient, stub.NewMockPellCoreBridge(), 123)

// 	//Check if xmsg was signed and broadcasted
// 	list := evmSigner.GetReportedTxList()
// 	require.Len(t, *list, 1)
// }

// func TestSigner_SignOutboundTx(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct

// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignOutboundTx - should successfully sign", func(t *testing.T) {
// 		// Call SignOutboundTx
// 		tx, err := evmSigner.SignOutboundTx(txData)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := stub.NewTSSMainnet()
// 		_, r, s := tx.RawSignatureValues()
// 		signature := append(r.Bytes(), s.Bytes()...)
// 		hash := evmSigner.EvmSigner().Hash(tx)

// 		verified := crypto.VerifySignature(tss.Pubkey(), hash.Bytes(), signature)
// 		require.True(t, verified)
// 	})
// }

// func TestSigner_SignRevertTx(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignRevertTx - should successfully sign", func(t *testing.T) {
// 		// Call SignRevertTx
// 		tx, err := evmSigner.SignRevertTx(txData)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := stub.NewTSSMainnet()
// 		_, r, s := tx.RawSignatureValues()
// 		signature := append(r.Bytes(), s.Bytes()...)
// 		hash := evmSigner.EvmSigner().Hash(tx)

// 		verified := crypto.VerifySignature(tss.Pubkey(), hash.Bytes(), signature)
// 		require.True(t, verified)
// 	})
// }

// func TestSigner_SignWithdrawTx(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignWithdrawTx - should successfully sign", func(t *testing.T) {
// 		// Call SignWithdrawTx
// 		tx, err := evmSigner.SignWithdrawTx(txData)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := stub.NewTSSMainnet()
// 		_, r, s := tx.RawSignatureValues()
// 		signature := append(r.Bytes(), s.Bytes()...)
// 		hash := evmSigner.EvmSigner().Hash(tx)

// 		verified := crypto.VerifySignature(tss.Pubkey(), hash.Bytes(), signature)
// 		require.True(t, verified)
// 	})
// }

// func TestSigner_SignCommandTx(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignCommandTx CmdWhitelistERC20", func(t *testing.T) {
// 		cmd := constant.CmdWhitelistERC20
// 		params := ConnectorAddress.Hex()
// 		// Call SignCommandTx
// 		tx, err := evmSigner.SignCommandTx(txData, cmd, params)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := stub.NewTSSMainnet()
// 		_, r, s := tx.RawSignatureValues()
// 		signature := append(r.Bytes(), s.Bytes()...)
// 		hash := evmSigner.EvmSigner().Hash(tx)

// 		verified := crypto.VerifySignature(tss.Pubkey(), hash.Bytes(), signature)
// 		require.True(t, verified)
// 	})

// 	t.Run("SignCommandTx CmdMigrateTssFunds", func(t *testing.T) {
// 		cmd := constant.CmdMigrateTssFunds
// 		// Call SignCommandTx
// 		tx, err := evmSigner.SignCommandTx(txData, cmd, "")
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := stub.NewTSSMainnet()
// 		_, r, s := tx.RawSignatureValues()
// 		signature := append(r.Bytes(), s.Bytes()...)
// 		hash := evmSigner.EvmSigner().Hash(tx)

// 		verified := crypto.VerifySignature(tss.Pubkey(), hash.Bytes(), signature)
// 		require.True(t, verified)
// 	})
// }

// func TestSigner_SignERC20WithdrawTx(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignERC20WithdrawTx - should successfully sign", func(t *testing.T) {
// 		// Call SignERC20WithdrawTx
// 		tx, err := evmSigner.SignERC20WithdrawTx(txData)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := stub.NewTSSMainnet()
// 		_, r, s := tx.RawSignatureValues()
// 		signature := append(r.Bytes(), s.Bytes()...)
// 		hash := evmSigner.EvmSigner().Hash(tx)

// 		verified := crypto.VerifySignature(tss.Pubkey(), hash.Bytes(), signature)
// 		require.True(t, verified)
// 	})
// }

// func TestSigner_BroadcastOutTx(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("BroadcastOutTx - should successfully broadcast", func(t *testing.T) {
// 		// Call SignERC20WithdrawTx
// 		tx, err := evmSigner.SignERC20WithdrawTx(txData)
// 		require.NoError(t, err)

// 		evmSigner.BroadcastOutTx(tx, xmsg, zerolog.Logger{}, sdktypes.AccAddress{}, stub.NewMockPellCoreBridge(), txData)

// 		//Check if xmsg was signed and broadcasted
// 		list := evmSigner.GetReportedTxList()
// 		require.Len(t, *list, 1)
// 	})
// }

// func TestSigner_getEVMRPC(t *testing.T) {
// 	t.Run("getEVMRPC error dialing", func(t *testing.T) {
// 		client, signer, err := getEVMRPC("invalidEndpoint")
// 		require.Nil(t, client)
// 		require.Nil(t, signer)
// 		require.Error(t, err)
// 	})
// }

// func TestSigner_SignerErrorMsg(t *testing.T) {
// 	xmsg := getXmsg(t)

// 	msg := SignerErrorMsg(xmsg)
// 	require.Contains(t, msg, "nonce 68270 chain 56")
// }

// func TestSigner_SignWhitelistERC20Cmd(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	tx, err := evmSigner.SignWhitelistERC20Cmd(txData, "")
// 	require.Nil(t, tx)
// 	require.ErrorContains(t, err, "invalid erc20 address")
// }
