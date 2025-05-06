package signer

import (
	"encoding/hex"
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/0xPellNetwork/aegis/relayer/chains/evm/observer"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	MinGasLimit = 100_000
	MaxGasLimit = 1_000_000
)

// OutBoundTransactionData is a data structure containing input fields used to construct each type of transaction.
// This is populated using xmsg and other input parameters passed to TryProcessOutTx
type OutBoundTransactionData struct {
	srcChainID *big.Int
	toChainID  *big.Int
	sender     ethcommon.Address
	to         ethcommon.Address

	gas Gas

	nonce  uint64
	height uint64

	// sendHash field is the inbound message digest that is sent to the destination contract
	xmsgIndex [32]byte

	// outboundParams field contains data detailing the receiver chain and outbound transaction
	outboundParams *types.OutboundTxParams

	pellTxData *types.InboundPellEvent
}

// NewOutBoundTransactionData populates transaction input fields parsed from the xmsg and other parameters
// returns
//  1. New OutBoundTransaction Data struct or nil if an error occurred.
//  2. bool (skipTx) - if the transaction doesn't qualify to be processed the function will return true, meaning that this
//     xmsg will be skipped and false otherwise.
//  3. error
func NewOutBoundTransactionData(
	xmsg *xmsgtypes.Xmsg,
	evmClient *observer.ChainClient,
	evmRPC interfaces.EVMRPCClient,
	logger zerolog.Logger,
	height uint64,
) (*OutBoundTransactionData, bool, error) {
	if xmsg == nil {
		return nil, false, errors.New("xmsg is nil")
	}

	outboundParams := xmsg.GetCurrentOutTxParam()
	if err := validateParams(outboundParams); err != nil {
		return nil, false, errors.Wrap(err, "invalid outboundParams")
	}

	to, toChainID, skip := getDestination(xmsg, logger)
	if skip {
		return nil, true, nil
	}

	gas, err := gasFromXmsg(logger, xmsg)
	if err != nil {
		return nil, false, errors.Wrap(err, "unable to make gas from Xmsg")
	}

	xmsgIndex, err := getXmsgIndices(xmsg)
	if err != nil {
		return nil, false, errors.Wrap(err, "unable to get xmsg index")
	}

	return &OutBoundTransactionData{
		srcChainID: big.NewInt(xmsg.InboundTxParams.SenderChainId),
		sender:     ethcommon.HexToAddress(xmsg.InboundTxParams.Sender),

		toChainID: toChainID,
		to:        to,

		gas: gas,

		nonce:  outboundParams.OutboundTxTssNonce,
		height: height,

		xmsgIndex: xmsgIndex,

		outboundParams: outboundParams,

		pellTxData: xmsg.InboundTxParams.InboundPellTx,
	}, false, nil
}

// getDestination picks the destination address and Chain ID based on the status of the cross chain tx.
// returns true if transaction should be skipped.
func getDestination(xmsg *xmsgtypes.Xmsg, logger zerolog.Logger) (ethcommon.Address, *big.Int, bool) {
	switch xmsg.XmsgStatus.Status {
	case xmsgtypes.XmsgStatus_PENDING_REVERT:
		to := ethcommon.HexToAddress(xmsg.InboundTxParams.Sender)
		chainID := big.NewInt(xmsg.InboundTxParams.SenderChainId)

		logger.Info().
			Str("xmsg.index", xmsg.Index).
			Int64("xmsg.chain_id", chainID.Int64()).
			Msgf("Abort: reverting inbound")

		return to, chainID, false
	case xmsgtypes.XmsgStatus_PENDING_OUTBOUND:
		to := ethcommon.HexToAddress(xmsg.GetCurrentOutTxParam().Receiver)
		chainID := big.NewInt(xmsg.GetCurrentOutTxParam().ReceiverChainId)

		return to, chainID, false
	}

	logger.Info().
		Str("xmsg.index", xmsg.Index).
		Str("xmsg.status", xmsg.XmsgStatus.String()).
		Msgf("Xmsg doesn't need to be processed")

	return ethcommon.Address{}, nil, true
}

func getXmsgIndices(xmsg *types.Xmsg) ([32]byte, error) {
	// `0x` + `64 chars`. Two chars ranging `00...FF` represent one byte (64 chars = 32 bytes)
	if len(xmsg.Index) != (2 + 64) {
		return [32]byte{}, fmt.Errorf("xmsg index %q is invalid", xmsg.Index)
	}

	// remove the leading `0x`
	xmsgIndexSlice, err := hex.DecodeString(xmsg.Index[2:])
	if err != nil || len(xmsgIndexSlice) != 32 {
		return [32]byte{}, errors.Wrapf(err, "unable to decode xmsg index %s", xmsg.Index)
	}

	var xmsgIndex [32]byte
	copy(xmsgIndex[:32], xmsgIndexSlice[:32])

	return xmsgIndex, nil
}

func validateParams(params *xmsgtypes.OutboundTxParams) error {
	if params == nil || params.OutboundTxGasLimit == 0 {
		return errors.New("outboundParams is empty")
	}

	return nil
}
