package signer

// import (
// 	"math/big"
// 	"testing"

// 	ethcommon "github.com/ethereum/go-ethereum/common"
// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/require"
// 	"github.com/0xPellNetwork/aegis/pkg/chains"
// 	"github.com/0xPellNetwork/aegis/x/xmsg/types"
// )

// func TestSigner_SetChainAndSender(t *testing.T) {
// 	// setup inputs
// 	xmsg := getXmsg(t)
// 	txData := &OutBoundTransactionData{}
// 	logger := zerolog.Logger{}

// 	t.Run("SetChainAndSender PendingRevert", func(t *testing.T) {
// 		xmsg.XmsgStatus.Status = types.XmsgStatus_PendingRevert
// 		skipTx := txData.SetChainAndSender(xmsg, logger)

// 		require.False(t, skipTx)
// 		require.Equal(t, ethcommon.HexToAddress(xmsg.InboundTxParams.Sender), txData.to)
// 		require.Equal(t, big.NewInt(xmsg.InboundTxParams.SenderChainId), txData.toChainID)
// 	})

// 	t.Run("SetChainAndSender PendingOutBound", func(t *testing.T) {
// 		xmsg.XmsgStatus.Status = types.XmsgStatus_PendingOutbound
// 		skipTx := txData.SetChainAndSender(xmsg, logger)

// 		require.False(t, skipTx)
// 		require.Equal(t, ethcommon.HexToAddress(xmsg.GetCurrentOutTxParam().Receiver), txData.to)
// 		require.Equal(t, big.NewInt(xmsg.GetCurrentOutTxParam().ReceiverChainId), txData.toChainID)
// 	})

// 	t.Run("SetChainAndSender Should skip xmsg", func(t *testing.T) {
// 		xmsg.XmsgStatus.Status = types.XmsgStatus_PendingInbound
// 		skipTx := txData.SetChainAndSender(xmsg, logger)
// 		require.True(t, skipTx)
// 	})
// }

// func TestSigner_SetupGas(t *testing.T) {
// 	xmsg := getXmsg(t)
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	txData := &OutBoundTransactionData{}
// 	logger := zerolog.Logger{}

// 	t.Run("SetupGas_success", func(t *testing.T) {
// 		chain := chains.BscMainnetChain()
// 		err := txData.SetupGas(xmsg, logger, evmSigner.EvmClient(), &chain)
// 		require.NoError(t, err)
// 	})

// 	t.Run("SetupGas_error", func(t *testing.T) {
// 		xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = "invalidGasPrice"
// 		chain := chains.BscMainnetChain()
// 		err := txData.SetupGas(xmsg, logger, evmSigner.EvmClient(), &chain)
// 		require.ErrorContains(t, err, "cannot convert gas price")
// 	})
// }

// func TestSigner_NewOutBoundTransactionData(t *testing.T) {
// 	// Setup evm signer
// 	evmSigner, err := getNewEvmSigner()
// 	require.NoError(t, err)

// 	mockChainClient, err := getNewEvmChainClient()
// 	require.NoError(t, err)

// 	t.Run("NewOutBoundTransactionData success", func(t *testing.T) {
// 		xmsg := getXmsg(t)
// 		_, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 		require.False(t, skip)
// 		require.NoError(t, err)
// 	})

// 	t.Run("NewOutBoundTransactionData skip", func(t *testing.T) {
// 		xmsg := getXmsg(t)
// 		xmsg.XmsgStatus.Status = types.XmsgStatus_Aborted
// 		_, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 		require.NoError(t, err)
// 		require.True(t, skip)
// 	})

// 	t.Run("NewOutBoundTransactionData unknown chain", func(t *testing.T) {
// 		xmsg := getInvalidXmsg(t)
// 		require.NoError(t, err)
// 		_, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 		require.ErrorContains(t, err, "unknown chain")
// 		require.True(t, skip)
// 	})

// 	t.Run("NewOutBoundTransactionData setup gas error", func(t *testing.T) {
// 		xmsg := getXmsg(t)
// 		require.NoError(t, err)
// 		xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = "invalidGasPrice"
// 		_, skip, err := NewOutBoundTransactionData(xmsg, mockChainClient, evmSigner.EvmClient(), zerolog.Logger{}, 123)
// 		require.True(t, skip)
// 		require.ErrorContains(t, err, "cannot convert gas price")
// 	})
// }
