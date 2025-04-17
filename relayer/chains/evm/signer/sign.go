package signer

import (
	"context"
	"encoding/base64"
	"fmt"

	"cosmossdk.io/errors"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pell-chain/pellcore/relayer/chains/evm"
	"github.com/pell-chain/pellcore/x/pevm/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func (signer *Signer) SignConnectorTx(ctx context.Context, txData *OutBoundTransactionData) (*ethtypes.Transaction, error) {
	pellSent := txData.pellTxData.GetPellSent()
	if pellSent == nil {
		return nil, fmt.Errorf("no pell data to be sent")
	}

	paramType, err := types.PellSentParamTypeFromString(pellSent.PellParams)
	if err != nil {
		// shouldn't return err because the following switch case has a fallback
		//return nil, errors.Wrap(err, "parse pellSent.PellParams error")
	}

	switch paramType {
	case types.ReceiveCall:
		return signer.signConnectorTxReceiveCall(ctx, pellSent, txData, paramType)
	case types.RevertableCall:
		return signer.signConnectorTxOnReceive(ctx, pellSent, txData, paramType)
	case types.Transfer:
		return signer.signTransferTx(ctx, pellSent, txData)
	default:
		// TODO: Remove this block in the next release
		// This block is maintained for backward compatibility.
		// Previously, Pell parameters were Base64-encoded.
		// Decode the Base64-encoded PellParams to handle legacy data.
		decodedData, err := base64.StdEncoding.DecodeString(pellSent.PellParams)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode pellSent.PellParams")
		}
		// Ensure the decoded data is not empty.
		if len(decodedData) == 0 {
			return nil, fmt.Errorf("invalid pellSent.PellParams: %s", pellSent.PellParams)
		}
		// The last byte of decoded data determines the type of transaction to process.
		lastByte := decodedData[len(decodedData)-1]
		switch lastByte {
		case 0:
			return signer.signConnectorTxReceiveCall(ctx, pellSent, txData, paramType)
		case 1:
			return signer.signConnectorTxOnReceive(ctx, pellSent, txData, paramType)
		case 2:
			return signer.signTransferTx(ctx, pellSent, txData)
		}

		return nil, fmt.Errorf("invalid pellSent.PellParams: %s", pellSent.PellParams)
	}
}

// function receiveCall(
//
//	  bytes calldata pellTxSenderAddress,
//	  uint256 sourceChainId,
//	  address destinationAddress,
//	  bytes calldata message,
//	  bytes32 internalSendHash
//	)
func (signer *Signer) signConnectorTxReceiveCall(
	ctx context.Context,
	pellSent *xmsgtypes.PellSent,
	txData *OutBoundTransactionData,
	paramType types.PellSentParamType,
) (*ethtypes.Transaction, error) {
	receiver := ethcommon.HexToAddress(pellSent.Receiver)
	message, err := base64.StdEncoding.DecodeString(pellSent.Message)
	if err != nil {
		return nil, errors.Wrapf(err, "decode pellSent.Message %s error", pellSent.Message)
	}

	methodName, err := paramType.MethodName()
	if err != nil {
		return nil, errors.Wrap(err, "get method name error")
	}

	data, err := connectorABI.Pack(
		methodName,
		ethcommon.HexToAddress(pellSent.Sender).Bytes(),
		txData.srcChainID,
		receiver,
		message,
		txData.xmsgIndex,
	)
	if err != nil {
		return nil, errors.Wrap(err, "connector pack error")
	}

	tx, _, _, err := signer.Sign(
		ctx,
		data,
		signer.pellConnectorAddress,
		zeroValue,
		txData.gas,
		txData.nonce,
		txData.height,
		txData.outboundParams.TssPubkey,
	)
	if err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}
	return tx, nil
}

// function onReceive(
//
//	bytes calldata pellTxSenderAddress,
//	uint256 sourceChainId,
//	address destinationAddress,
//	uint256 pellValue,
//	bytes calldata message,
//	bytes32 internalSendHash
//
// )
func (signer *Signer) signConnectorTxOnReceive(
	ctx context.Context,
	pellSent *xmsgtypes.PellSent,
	txData *OutBoundTransactionData,
	paramType types.PellSentParamType,
) (*ethtypes.Transaction, error) {
	receiver := ethcommon.HexToAddress(pellSent.Receiver)
	message, err := base64.StdEncoding.DecodeString(pellSent.Message)
	if err != nil {
		return nil, errors.Wrapf(err, "decode pellSent.Message %s error", pellSent.Message)
	}

	methodName, err := paramType.MethodName()
	if err != nil {
		return nil, errors.Wrap(err, "get method name error")
	}

	data, err := connectorABI.Pack(
		methodName,
		ethcommon.HexToAddress(pellSent.Sender).Bytes(),
		txData.srcChainID,
		receiver,
		pellSent.PellValue.BigInt(),
		message,
		txData.xmsgIndex,
	)
	if err != nil {
		return nil, errors.Wrap(err, "connector pack error")
	}

	tx, _, _, err := signer.Sign(
		ctx,
		data,
		signer.pellConnectorAddress,
		zeroValue,
		txData.gas,
		txData.nonce,
		txData.height,
		txData.outboundParams.TssPubkey,
	)
	if err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}
	return tx, nil
}

// SignCancelTx signs a transaction from TSS address to itself with a zero amount in order to increment the nonce
func (signer *Signer) SignCancelTx(ctx context.Context, txData *OutBoundTransactionData) (*ethtypes.Transaction, error) {
	txData.gas.Limit = evm.EthTransferGasLimit
	tx, _, _, err := signer.Sign(
		ctx,
		nil,
		signer.TSS().EVMAddress(),
		zeroValue, // zero out the amount to cancel the tx
		txData.gas,
		txData.nonce,
		txData.height,
		txData.outboundParams.TssPubkey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "SignCancel error")
	}

	return tx, nil
}

func (signer *Signer) signTransferTx(
	ctx context.Context,
	pellSent *xmsgtypes.PellSent,
	txData *OutBoundTransactionData,
) (*ethtypes.Transaction, error) {
	receiver := ethcommon.HexToAddress(pellSent.Receiver)

	tx, _, _, err := signer.Sign(
		ctx,
		nil,
		receiver,
		pellSent.PellValue.BigInt(),
		txData.gas,
		txData.nonce,
		txData.height,
		txData.outboundParams.TssPubkey,
	)
	if err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}

	return tx, nil
}
