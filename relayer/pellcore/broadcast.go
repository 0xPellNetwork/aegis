package pellcore

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"github.com/0xPellNetwork/aegis/app/ante"
	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/relayer/authz"
	"github.com/0xPellNetwork/aegis/relayer/hsm"
)

var (
	// paying 50% more than the current base gas price to buffer for potential block-by-block
	// gas price increase due to EIP1559 feemarket on PellChain
	bufferMultiplier = sdkmath.LegacyMustNewDecFromStr("1.5")
)

// Broadcast Broadcasts tx to metachain. Returns txHash and error
func (b *PellCoreBridge) Broadcast(
	ctx context.Context,
	gaslimit uint64,
	authzWrappedMsgs []sdktypes.Msg,
	authzSigner authz.Signer,
) (string, error) {
	blockHeight, err := b.GetBlockHeight(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to get block height")
	}

	params, err := b.GetFeemarketParams(ctx)
	if err != nil {
		return "", errors.Wrap(err, "unable to get base gas price")
	}

	baseGasPrice := params.BaseFee.Int64()
	if baseGasPrice == 0 {
		baseGasPrice = DefaultBaseGasPrice // shouldn't happen, but just in case
	}

	reductionRate := sdkmath.LegacyMustNewDecFromStr(ante.GasPriceReductionRate)

	// TODO:
	// multiply gas price by the system tx reduction rate
	adjustedBaseGasPrice := sdkmath.LegacyNewDec(baseGasPrice).Mul(reductionRate).Mul(bufferMultiplier)
	if adjustedBaseGasPrice.LTE(params.MinGasPrice.Mul(sdkmath.LegacyMustNewDecFromStr(ante.GasPriceReductionRate))) {
		adjustedBaseGasPrice = params.MinGasPrice.Mul(sdkmath.LegacyMustNewDecFromStr(ante.GasPriceReductionRate)).Add(sdkmath.LegacyOneDec())
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if blockHeight > b.blockHeight {
		b.blockHeight = blockHeight
		accountNumber, seqNumber, err := b.GetAccountNumberAndSequenceNumber(authzSigner.KeyType)
		if err != nil {
			return "", err
		}

		b.accountNumber[authzSigner.KeyType] = accountNumber
		if b.seqNumber[authzSigner.KeyType] < seqNumber {
			b.seqNumber[authzSigner.KeyType] = seqNumber
		}
	}

	flags := flag.NewFlagSet("pellclient", 0)
	factory, err := clienttx.NewFactoryCLI(b.cosmosClientContext, flags)
	if err != nil {
		return "", err
	}

	factory = factory.WithAccountNumber(b.accountNumber[authzSigner.KeyType])
	factory = factory.WithSequence(b.seqNumber[authzSigner.KeyType])
	factory = factory.WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
	builder, err := factory.BuildUnsignedTx(authzWrappedMsgs...)
	if err != nil {
		return "", errors.Wrap(err, "unable to build unsigned tx")
	}

	builder.SetGasLimit(gaslimit)

	// #nosec G701 always in range
	fee := sdktypes.NewCoins(sdktypes.NewCoin(config.BaseDenom,
		sdkmath.NewInt(int64(gaslimit)).Mul(adjustedBaseGasPrice.Ceil().RoundInt())))
	builder.SetFeeAmount(fee)

	err = b.SignTx(factory, b.cosmosClientContext.GetFromName(), builder, true, b.cosmosClientContext.TxConfig)
	if err != nil {
		return "", errors.Wrap(err, "unable to sign tx")
	}

	txBytes, err := b.cosmosClientContext.TxConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return "", errors.Wrap(err, "unable to encode tx")
	}

	// broadcast to a Tendermint node
	commit, err := b.cosmosClientContext.BroadcastTxSync(txBytes)
	if err != nil {
		return "", errors.Wrap(err, "fail to broadcast tx sync")
	}

	// Code will be the tendermint ABICode , it start at 1 , so if it is an error , code will not be zero
	if commit.Code > 0 {
		if commit.Code == 32 {
			if err := b.handleSequenceMismatchError(commit, authzSigner); err != nil {
				return commit.TxHash, err
			}
		}
		return commit.TxHash, fmt.Errorf("fail to broadcast to pellchain,code:%d, log:%s", commit.Code, commit.RawLog)
	}

	// increment seqNum
	b.seqNumber[authzSigner.KeyType] = b.seqNumber[authzSigner.KeyType] + 1

	return commit.TxHash, nil
}

func (b *PellCoreBridge) SignTx(
	txf clienttx.Factory,
	name string,
	txBuilder client.TxBuilder,
	overwriteSig bool,
	txConfig client.TxConfig,
) error {
	if b.config.HsmMode {
		return hsm.SignWithHSM(txf, name, txBuilder, overwriteSig, txConfig)
	}
	return clienttx.Sign(context.Background(), txf, name, txBuilder, overwriteSig)
}

// QueryTxResult query the result of a tx
func (b *PellCoreBridge) QueryTxResult(hash string) (*sdktypes.TxResponse, error) {
	return authtx.QueryTx(b.cosmosClientContext, hash)
}

// HandleBroadcastError returns whether to retry in a few seconds, and whether to report via AddTxHashToOutTxTracker
// returns (bool retry, bool report)
func HandleBroadcastError(err error, nonce, toChain, outTxHash string) (bool, bool) {
	if strings.Contains(err.Error(), "nonce too low") {
		log.Warn().Err(err).Msgf("nonce too low! this might be a unnecessary key-sign. increase re-try interval and awaits outTx confirmation")
		return false, false
	}
	if strings.Contains(err.Error(), "replacement transaction underpriced") {
		log.Warn().Err(err).Msgf("Broadcast replacement: nonce %s chain %s outTxHash %s", nonce, toChain, outTxHash)
		return false, false
	} else if strings.Contains(err.Error(), "already known") { // this is error code from QuickNode
		log.Warn().Err(err).Msgf("Broadcast duplicates: nonce %s chain %s outTxHash %s", nonce, toChain, outTxHash)
		return false, true // report to tracker, because there's possibilities a successful broadcast gets this error code
	}

	log.Error().Err(err).Msgf("Broadcast error: nonce %s chain %s outTxHash %s; retrying...", nonce, toChain, outTxHash)
	return true, false
}

// handleSequenceMismatchError handles the sequence mismatch error (code 32)
// and updates the sequence number if necessary
func (b *PellCoreBridge) handleSequenceMismatchError(commit *sdktypes.TxResponse, authzSigner authz.Signer) error {
	errMsg := commit.RawLog
	re := regexp.MustCompile(`account sequence mismatch, expected ([0-9]*), got ([0-9]*)`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) != 3 {
		return fmt.Errorf("invalid sequence mismatch error format: %s", errMsg)
	}

	expectedSeq, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		b.logger.Warn().Msgf("cannot parse expected seq %s", matches[1])
		return err
	}

	gotSeq, err := strconv.Atoi(matches[2])
	if err != nil {
		b.logger.Warn().Msgf("cannot parse got seq %s", matches[2])
		return err
	}

	b.seqNumber[authzSigner.KeyType] = expectedSeq
	b.logger.Warn().Msgf("Reset seq number to %d (from err msg) from %d", b.seqNumber[authzSigner.KeyType], gotSeq)
	return nil
}
