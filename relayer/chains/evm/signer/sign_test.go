package signer

// import (
// 	ethtypes "github.com/ethereum/go-ethereum/core/types"
// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/0xPellNetwork/aegis/relayer/testutils/mocks"
// 	"math/big"
// 	"testing"
// )

// func TestSigner_SignConnectorOnReceive(t *testing.T) {
// 	ctx := makeCtx(t)

// 	// Setup evm signer
// 	tss := mocks.NewTSSMainnet()
// 	evmSigner, err := getNewEvmSigner(tss)
// 	require.NoError(t, err)

// 	// Setup txData struct

// 	xmsg := getXmsg(t)
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutboundData(ctx, xmsg, 123, zerolog.Logger{})
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignConnectorOnReceive - should successfully sign", func(t *testing.T) {
// 		// Call SignConnectorOnReceive
// 		tx, err := evmSigner.SignConnectorOnReceive(ctx, txData)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := mocks.NewTSSMainnet()
// 		verifyTxSignature(t, tx, tss.Pubkey(), evmSigner.EvmSigner())
// 	})
// 	t.Run("SignConnectorOnReceive - should fail if keysign fails", func(t *testing.T) {
// 		// Pause tss to make keysign fail
// 		tss.Pause()

// 		// Call SignConnectorOnReceive
// 		tx, err := evmSigner.SignConnectorOnReceive(ctx, txData)
// 		require.ErrorContains(t, err, "sign onReceive error")
// 		require.Nil(t, tx)
// 		tss.Unpause()
// 	})

// 	t.Run("SignOutbound - should successfully sign LegacyTx", func(t *testing.T) {
// 		// Call SignOutbound
// 		tx, err := evmSigner.SignConnectorOnReceive(ctx, txData)
// 		require.NoError(t, err)

// 		// Verify Signature
// 		tss := mocks.NewTSSMainnet()
// 		verifyTxSignature(t, tx, tss.Pubkey(), evmSigner.EvmSigner())

// 		// check that by default tx type is legacy tx
// 		assert.Equal(t, ethtypes.LegacyTxType, int(tx.Type()))
// 	})

// 	t.Run("SignOutbound - should successfully sign DynamicFeeTx", func(t *testing.T) {
// 		// ARRANGE
// 		const (
// 			gwei        = 1_000_000_000
// 			priorityFee = 1 * gwei
// 			gasPrice    = 3 * gwei
// 		)

// 		// Given a Xmsg with gas price and priority fee
// 		xmsg := getXmsg(t)
// 		xmsg.OutboundParams[0].GasPrice = big.NewInt(gasPrice).String()
// 		xmsg.OutboundParams[0].GasPriorityFee = big.NewInt(priorityFee).String()

// 		// Given outbound data
// 		txData, skip, err := NewOutboundData(ctx, xmsg, 123, makeLogger(t))
// 		require.False(t, skip)
// 		require.NoError(t, err)

// 		// Given a working TSS
// 		tss.Unpause()

// 		// ACT
// 		tx, err := evmSigner.SignConnectorOnReceive(ctx, txData)
// 		require.NoError(t, err)

// 		// ASSERT
// 		verifyTxSignature(t, tx, mocks.NewTSSMainnet().Pubkey(), evmSigner.EvmSigner())

// 		// check that by default tx type is a dynamic fee tx
// 		assert.Equal(t, ethtypes.DynamicFeeTxType, int(tx.Type()))

// 		// check that the gasPrice & priorityFee are set correctly
// 		assert.Equal(t, int64(gasPrice), tx.GasFeeCap().Int64())
// 		assert.Equal(t, int64(priorityFee), tx.GasTipCap().Int64())
// 	})
// }

// func TestSigner_SignCancel(t *testing.T) {
// 	ctx := makeCtx(t)

// 	// Setup evm signer
// 	tss := mocks.NewTSSMainnet()
// 	evmSigner, err := getNewEvmSigner(tss)
// 	require.NoError(t, err)

// 	// Setup txData struct
// 	xmsg := getXmsg(t)
// 	require.NoError(t, err)
// 	txData, skip, err := NewOutboundData(ctx, xmsg, 123, zerolog.Logger{})
// 	require.False(t, skip)
// 	require.NoError(t, err)

// 	t.Run("SignCancel - should successfully sign", func(t *testing.T) {
// 		// Call SignConnectorOnRevert
// 		tx, err := evmSigner.SignCancel(ctx, txData)
// 		require.NoError(t, err)

// 		// Verify tx signature
// 		tss := mocks.NewTSSMainnet()
// 		verifyTxSignature(t, tx, tss.Pubkey(), evmSigner.EvmSigner())

// 		// Verify tx body basics
// 		// Note: Cancel tx sends 0 gas token to TSS self address
// 		verifyTxBodyBasics(t, tx, evmSigner.TSS().EVMAddress(), txData.nonce, big.NewInt(0))
// 	})
// 	t.Run("SignCancel - should fail if keysign fails", func(t *testing.T) {
// 		// Pause tss to make keysign fail
// 		tss.Pause()

// 		// Call SignCancel
// 		tx, err := evmSigner.SignCancel(ctx, txData)
// 		require.ErrorContains(t, err, "SignCancel error")
// 		require.Nil(t, tx)
// 	})
// }
