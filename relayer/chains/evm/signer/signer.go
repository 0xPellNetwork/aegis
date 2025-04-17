package signer

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/chains/base"
	"github.com/pell-chain/pellcore/relayer/chains/evm/observer"
	"github.com/pell-chain/pellcore/relayer/chains/interfaces"
	"github.com/pell-chain/pellcore/relayer/compliance"
	pctx "github.com/pell-chain/pellcore/relayer/context"
	"github.com/pell-chain/pellcore/relayer/logs"
	clientlogs "github.com/pell-chain/pellcore/relayer/logs"
	"github.com/pell-chain/pellcore/relayer/metrics"
	"github.com/pell-chain/pellcore/relayer/outtxprocessor"
	zbridge "github.com/pell-chain/pellcore/relayer/pellcore"
	"github.com/pell-chain/pellcore/relayer/testutils/stub"
	"github.com/pell-chain/pellcore/x/xmsg/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	// broadcastBackoff is the initial backoff duration for retrying broadcast
	broadcastBackoff = 1000 * time.Millisecond

	// broadcastRetries is the maximum number of retries for broadcasting a transaction
	broadcastRetries = 5
)

var (
	_ interfaces.ChainSigner = (*Signer)(nil)

	// zeroValue is for outbounds that carry no ETH (gas token) value
	zeroValue = big.NewInt(0)
)

// Signer deals with the signing EVM transactions and implements the ChainSigner interface
type Signer struct {
	*base.Signer

	// client is the EVM RPC client to interact with the EVM chain
	client interfaces.EVMRPCClient

	// ethSigner encapsulates EVM transaction signature handling
	ethSigner ethtypes.Signer

	// mu protects below fields from concurrent access
	pellConnectorAddress ethcommon.Address
}

func NewEVMSigner(
	ctx context.Context,
	chain chains.Chain,
	endpoint string,
	tss interfaces.TSSSigner,
	pellConnectorAddress ethcommon.Address,
	logger clientlogs.Logger,
	ts *metrics.TelemetryServer,
) (*Signer, error) {
	appContext, err := pctx.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get app context")
	}

	// create base signer
	baseSigner := base.NewSigner(chain, tss, appContext.PellCoreContext(), ts, logger)

	client, ethSigner, err := getEVMRPC(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	return &Signer{
		Signer:               baseSigner,
		client:               client,
		ethSigner:            ethSigner,
		pellConnectorAddress: pellConnectorAddress,
	}, nil
}

// SetPellConnectorAddress set the OmniOperatorSharesManager address
func (signer *Signer) SetPellConnectorAddress(addr ethcommon.Address) {
	signer.Lock()
	defer signer.Unlock()
	signer.pellConnectorAddress = addr
}

// GetPellConnectorAddress returns the OmniOperatorSharesManager address
func (signer *Signer) GetPellConnectorAddress() ethcommon.Address {
	signer.Lock()
	defer signer.Unlock()
	return signer.pellConnectorAddress
}

// Sign given data, and metadata (gas, nonce, etc)
// returns a signed transaction, sig bytes, hash bytes, and error
func (signer *Signer) Sign(
	ctx context.Context,
	data []byte,
	to ethcommon.Address,
	amount *big.Int,
	gas Gas,
	nonce uint64,
	height uint64,
	optionalPubkey string,
) (*ethtypes.Transaction, []byte, []byte, error) {
	signer.Logger().Std.Info().
		Str("tss_pub_key", signer.TSS().EVMAddress().String()).
		Msgf("Signing evm transaction")

	chainID := big.NewInt(signer.Chain().Id)

	log.Info().Msgf("Sign: ChainID: %d, Data: %s, To: %s, Amount: %s, Gas: %s, Nonce: %d", chainID, hex.EncodeToString(data), to.Hex(), amount.String(), gas.Price.String(), nonce)

	tx, err := newTx(chainID, data, to, amount, gas, nonce)
	if err != nil {
		return nil, nil, nil, err
	}

	hashBytes := signer.ethSigner.Hash(tx).Bytes()

	sig, err := signer.TSS().Sign(ctx, hashBytes, height, nonce, signer.Chain(), optionalPubkey)
	if err != nil {
		return nil, nil, nil, err
	}

	log.Info().Msgf("Sign: Signature: %s, Signer address: %s, To address: %s", hex.EncodeToString(sig[:]), signer.EvmSigner(), to)

	pubk, err := crypto.SigToPub(hashBytes, sig[:])
	if err != nil {
		signer.Logger().Std.Error().Err(err).Msgf("SigToPub error")
	}

	addr := crypto.PubkeyToAddress(*pubk)
	signer.Logger().Std.Info().Msgf("Sign: Ecrecovery of signature: %s", addr.Hex())

	signedTX, err := tx.WithSignature(signer.ethSigner, sig[:])
	if err != nil {
		return nil, nil, nil, err
	}

	return signedTX, sig[:], hashBytes[:], nil
}

func newTx(
	chainID *big.Int,
	data []byte,
	to ethcommon.Address,
	amount *big.Int,
	gas Gas,
	nonce uint64,
) (*ethtypes.Transaction, error) {
	if err := gas.validate(); err != nil {
		return nil, errors.Wrap(err, "invalid gas parameters")
	}

	if gas.isLegacy() {
		return ethtypes.NewTx(&ethtypes.LegacyTx{
			To:       &to,
			Value:    amount,
			Data:     data,
			GasPrice: gas.Price,
			Gas:      gas.Limit,
			Nonce:    nonce,
		}), nil
	}

	return ethtypes.NewTx(&ethtypes.DynamicFeeTx{
		ChainID:   chainID,
		To:        &to,
		Value:     amount,
		Data:      data,
		GasFeeCap: gas.Price,
		GasTipCap: gas.PriorityFee,
		Gas:       gas.Limit,
		Nonce:     nonce,
	}), nil
}

// Broadcast takes in signed tx, broadcast to external chain node
func (signer *Signer) Broadcast(tx *ethtypes.Transaction) error {
	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return signer.client.SendTransaction(ctxt, tx)
}

func (signer *Signer) TryProcessOutTx(
	ctx context.Context,
	xmsg *xmsgtypes.Xmsg,
	outTxMan *outtxprocessor.Processor,
	outTxID string,
	chainclient interfaces.ChainClient,
	pellBridge interfaces.PellCoreBridger,
	height uint64,
) {
	if xmsg.IsCrossChainPellTx() {
		signer.TryProcessOutTx_i(ctx, xmsg, outTxMan, outTxID, chainclient, pellBridge, height)
	} else {
		signer.Logger().Std.Error().
			Str(logs.FieldMethod, "TryProcessOutbound").
			Int64(logs.FieldChain, signer.Chain().Id).
			Str(logs.FieldXmsg, xmsg.Index).
			Msg("TryProcessOutbound: don't support xmsg content")
	}
}

// TryProcessOutTx - signer interface implementation
// This function will attempt to build and sign an evm transaction using the TSS signer.
// It will then broadcast the signed transaction to the outbound chain.
func (signer *Signer) TryProcessOutTx_i(
	ctx context.Context,
	xmsg *xmsgtypes.Xmsg,
	outTxMan *outtxprocessor.Processor,
	outTxID string,
	chainclient interfaces.ChainClient,
	pellBridge interfaces.PellCoreBridger,
	height uint64,
) {
	defer func() {
		outTxMan.EndTryProcess(outTxID)
		if r := recover(); r != nil {
			signer.Logger().Std.Error().Msgf("TryProcessOutbound: %s, caught panic error: %v", xmsg.Index, r)
		}
	}()

	var (
		params = xmsg.GetCurrentOutTxParam()
		myID   = pellBridge.GetKeys().GetOperatorAddress()
		logger = signer.Logger().Std.With().
			Str(logs.FieldMethod, "TryProcessOutbound").
			Int64(logs.FieldChain, signer.Chain().Id).
			Uint64(logs.FieldNonce, params.OutboundTxTssNonce).
			Str(logs.FieldXmsg, xmsg.Index).
			Str("xmsg.receiver", params.Receiver).
			Logger()
	)
	logger.Info().Msg("TryProcessOutbound start")

	evmClient, ok := chainclient.(*observer.ChainClient)
	if !ok {
		logger.Error().Msg("chain client is not an EVMChainClient")
		return
	}
	skipTx, err := signer.IsOutboundProcessed(ctx, xmsg, evmClient, logger)
	if err != nil {
		logger.Err(err).Msg("error setting up transaction input fields")
		return
	}
	if skipTx {
		return
	}

	// Setup Transaction input
	txData, skipTx, err := NewOutBoundTransactionData(xmsg, evmClient, signer.client, logger, height)
	if err != nil {
		logger.Err(err).Msg("error setting up transaction input fields")
		return
	}
	if skipTx {
		return
	}

	// Get destination chain for logging
	toChain, exist := chains.GetChainByChainId(txData.toChainID.Int64())
	if !exist {
		logger.Error().Err(err).Msgf("error getting toChain %d", txData.toChainID.Int64())
		return
	}

	if toChain.IsPellChain() {
		// should not happen
		logger.Error().Msgf("unable to TryProcessOutbound when toChain is pellChain (%d)", toChain.Id)
		return
	}

	// sign outbound
	tx, err := signer.SignOutboundFromXmsg(
		ctx,
		logger,
		xmsg,
		txData,
		evmClient,
		pellBridge,
		toChain.Id,
	)
	if err != nil {
		logger.Err(err).Msg(SignerErrorMsg(xmsg))
		return
	}

	logger.Info().Msgf("Key-sign success: %d => %d, nonce %d",
		xmsg.InboundTxParams.SenderChainId,
		toChain.Id,
		xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
	)

	// Broadcast Signed Tx
	signer.BroadcastOutTx(ctx, tx, xmsg, logger, myID, pellBridge, txData)
}

func (signer *Signer) IsOutboundProcessed(
	ctx context.Context,
	xmsg *xmsgtypes.Xmsg,
	evmClient *observer.ChainClient,
	logger zerolog.Logger,
) (bool, error) {
	// Get nonce, Early return if the xmsg is already processed
	included, confirmed, err := evmClient.IsOutboundProcessed(ctx, xmsg, logger)
	if err != nil {
		return true, errors.New("IsOutboundProcessed failed")
	}
	if included || confirmed {
		logger.Info().Msgf("Xmsg already processed; exit signer")
		return true, nil
	}

	// In case there is a pending transaction, make sure this keysign is a transaction replacement
	nonce := xmsg.GetCurrentOutTxParam().OutboundTxTssNonce
	gasPrice, ok := new(big.Int).SetString(xmsg.GetCurrentOutTxParam().OutboundTxGasPrice, 10)
	if !ok {
		return true, fmt.Errorf("cannot convert gas price  %s ", xmsg.GetCurrentOutTxParam().OutboundTxGasPrice)
	}
	pendingTx := evmClient.GetPendingTx(nonce)
	if pendingTx != nil {
		if gasPrice.Cmp(pendingTx.GasPrice()) > 0 {
			logger.Info().Msgf("replace pending outTx %s nonce %d using gas price %d", pendingTx.Hash().Hex(), nonce, gasPrice)
		} else {
			logger.Info().Msgf("please wait for pending outTx %s nonce %d to be included", pendingTx.Hash().Hex(), nonce)
			return true, nil
		}
	}
	return false, nil
}

// SignOutboundFromXmsg signs an outbound transaction from a given xmsg
// TODO: simplify logic with all if else
func (signer *Signer) SignOutboundFromXmsg(
	ctx context.Context,
	logger zerolog.Logger,
	xmsg *xmsgtypes.Xmsg,
	txData *OutBoundTransactionData,
	evmClient *observer.ChainClient,
	pellBridge interfaces.PellCoreBridger,
	toChain int64,
) (*ethtypes.Transaction, error) {
	// compliance check goes first
	if compliance.IsXmsgRestricted(xmsg) {
		compliance.PrintComplianceLog(logger, signer.Logger().Compliance,
			true, evmClient.Chain().Id, xmsg.Index, xmsg.InboundTxParams.Sender, txData.to.Hex(), "")
		return signer.SignCancelTx(ctx, txData) // cancel the tx
	} else if IsSenderPellChain(xmsg, pellBridge) {
		logger.Info().Msgf("SignOutboundTx: %d => %d, nonce %d, gasPrice %d",
			xmsg.InboundTxParams.SenderChainId, toChain, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, txData.gas.Price)
		return signer.SignConnectorTx(ctx, txData)
	} else if xmsg.XmsgStatus.Status == xmsgtypes.XmsgStatus_PENDING_REVERT && xmsg.OutboundTxParams[0].ReceiverChainId == pellBridge.Chain().Id {
		logger.Error().Msgf("PendingRevertTx: don't be supported. %d => %d, nonce %d, gasPrice %d",
			xmsg.InboundTxParams.SenderChainId, toChain, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, txData.gas.Price)
		return nil, errors.Errorf("PendingRevertTx: don't be supported. %d => %d, nonce %d, gasPrice %d",
			xmsg.InboundTxParams.SenderChainId, toChain, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, txData.gas.Price)
	} else if xmsg.XmsgStatus.Status == xmsgtypes.XmsgStatus_PENDING_REVERT {
		logger.Info().Msgf("SignRevertTx: don't be supported. %d => %d, nonce %d, gasPrice %d",
			xmsg.InboundTxParams.SenderChainId, toChain, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, txData.gas.Price)
		return nil, errors.Errorf("SignRevertTx: don't be supported. %d => %d, nonce %d, gasPrice %d",
			xmsg.InboundTxParams.SenderChainId, toChain, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, txData.gas.Price)
	} else if xmsg.XmsgStatus.Status == xmsgtypes.XmsgStatus_PENDING_OUTBOUND {
		logger.Info().Msgf("SignOutboundTx: %d => %d, nonce %d, gasPrice %d",
			xmsg.InboundTxParams.SenderChainId, toChain, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, txData.gas.Price)
		return signer.SignConnectorTx(ctx, txData)
	}
	return nil, fmt.Errorf("SignOutboundFromXmsg: can't determine how to sign outbound from xmsg %s", xmsg.String())
}

// BroadcastOutTx signed transaction through evm rpc client
func (signer *Signer) BroadcastOutTx(
	ctx context.Context,
	tx *ethtypes.Transaction,
	xmsg *xmsgtypes.Xmsg,
	logger zerolog.Logger,
	myID sdk.AccAddress,
	pellBridge interfaces.PellCoreBridger,
	txData *OutBoundTransactionData) {
	// Get destination chain for logging
	toChain, _ := chains.GetChainByChainId(txData.toChainID.Int64())
	if tx == nil {
		logger.Warn().Msgf("BroadcastOutTx: no tx to broadcast %s", xmsg.Index)
	} else {
		// Try to broadcast transaction
		outTxHash := tx.Hash().Hex()
		logger.Info().Msgf("BroadcastOutTx: on chain %s nonce %d, outTxHash %s signer %s", signer.Chain(), xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, outTxHash, myID)
		//if len(signers) == 0 || myid == signers[send.OutboundTxParams.Broadcaster] || myid == signers[int(send.OutboundTxParams.Broadcaster+1)%len(signers)] {
		backOff := broadcastBackoff
		// retry loop: 1s, 2s, 4s, 8s, 16s in case of RPC error
		for i := 0; i < broadcastRetries; i++ {
			logger.Info().Msgf("BroadcastOutTx: broadcasting tx %s to chain %s: nonce %d, retry %d", outTxHash, toChain.ChainName(), xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, i)
			// #nosec G404 randomness is not a security issue here
			time.Sleep(backOff)
			err := signer.Broadcast(tx)
			if err != nil {
				log.Warn().
					Err(err).
					Msgf("BroadcastOutbound: error broadcasting tx %s on chain %d nonce %d retry %d signer %s",
						outTxHash, toChain.Id, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, i, myID)

				retry, report := zbridge.HandleBroadcastError(
					err,
					strconv.FormatUint(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, 10),
					toChain.String(),
					outTxHash)
				if report {
					signer.reportToOutTxTracker(ctx, pellBridge, toChain.Id, tx.Nonce(), outTxHash, logger)
				}
				if !retry {
					break
				}
				backOff *= 2
				continue
			}
			logger.Info().Msgf("Broadcast success: nonce %d to chain %s outTxHash %s", xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, toChain.ChainName(), outTxHash)
			logger.Info().Msgf("BroadcastOutbound: broadcasted tx %s on chain %d nonce %d signer %s",
				outTxHash, toChain.Id, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce, myID)
			signer.reportToOutTxTracker(ctx, pellBridge, toChain.Id, tx.Nonce(), outTxHash, logger)
			break // successful broadcast; no need to retry
		}
	}
}

// Exported for unit tests

func (signer *Signer) EvmClient() interfaces.EVMRPCClient {
	return signer.client
}

func (signer *Signer) EvmSigner() ethtypes.Signer {
	return signer.ethSigner
}

// ________________________

// getEVMRPC is a helper function to set up the client and signer, also initializes a mock client for unit tests
func getEVMRPC(ctx context.Context, endpoint string) (interfaces.EVMRPCClient, ethtypes.Signer, error) {
	if endpoint == stub.EVMRPCEnabled {
		chainID := big.NewInt(chains.BscMainnetChain().Id)
		ethSigner := ethtypes.NewLondonSigner(chainID)
		client := &stub.MockEvmClient{}
		return client, ethSigner, nil
	}

	httpClient, err := metrics.GetInstrumentedHTTPClient(endpoint)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get instrumented HTTP client")
	}

	rpcClient, err := ethrpc.DialHTTPWithClient(endpoint, httpClient)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to dial EVM client (endpoint %q)", endpoint)
	}
	client := ethclient.NewClient(rpcClient)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get chain ID")
	}

	ethSigner := ethtypes.LatestSignerForChainID(chainID)
	return client, ethSigner, nil
}

func IsSenderPellChain(cctx *types.Xmsg, pellBridge interfaces.PellCoreBridger) bool {
	return cctx.InboundTxParams.SenderChainId == pellBridge.Chain().Id &&
		cctx.XmsgStatus.Status == types.XmsgStatus_PENDING_OUTBOUND
}

func SignerErrorMsg(xmsg *types.Xmsg) string {
	return fmt.Sprintf("signer SignOutbound error: nonce %d chain %d",
		xmsg.GetCurrentOutTxParam().OutboundTxTssNonce,
		xmsg.GetCurrentOutTxParam().ReceiverChainId)
}
