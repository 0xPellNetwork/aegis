package runner

import (
	"fmt"
	"sync"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	loggerSeparator = " | "
	padding         = 10
)

// Logger is a wrapper around log.Logger that adds verbosity
type Logger struct {
	verbose bool
	logger  *color.Color
	prefix  string
	mu      sync.Mutex
}

// NewLogger creates a new Logger
func NewLogger(verbose bool, printColor color.Attribute, prefix string) *Logger {
	// trim prefix to padding
	if len(prefix) > padding {
		prefix = prefix[:padding]
	}

	return &Logger{
		verbose: verbose,
		logger:  color.New(printColor),
		prefix:  prefix,
	}
}

// SetColor sets the color of the logger
func (l *Logger) SetColor(printColor color.Attribute) {
	l.logger = color.New(printColor)
}

// Prefix returns the prefix of the logger
func (l *Logger) Prefix() string {
	return l.getPrefixWithPadding() + loggerSeparator
}

// Print prints a message to the logger
func (l *Logger) Print(message string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	text := fmt.Sprintf(message, args...)
	// #nosec G104 - we are not using user input
	l.logger.Printf(l.getPrefixWithPadding() + loggerSeparator + text + "\n")
}

// PrintNoPrefix prints a message to the logger without the prefix
func (l *Logger) PrintNoPrefix(message string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	text := fmt.Sprintf(message, args...)
	// #nosec G104 - we are not using user input
	l.logger.Printf(text + "\n")
}

// Info prints a message to the logger if verbose is true
func (l *Logger) Info(message string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.verbose {
		text := fmt.Sprintf(message, args...)
		// #nosec G104 - we are not using user input
		l.logger.Printf(l.getPrefixWithPadding() + loggerSeparator + "[INFO]" + text + "\n")
	}
}

// InfoLoud prints a message to the logger if verbose is true
func (l *Logger) InfoLoud(message string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.verbose {
		text := fmt.Sprintf(message, args...)
		// #nosec G104 - we are not using user input
		l.logger.Printf(l.getPrefixWithPadding() + loggerSeparator + "[INFO] =======================================")
		// #nosec G104 - we are not using user input
		l.logger.Printf(l.getPrefixWithPadding() + loggerSeparator + "[INFO]" + text + "\n")
		// #nosec G104 - we are not using user input
		l.logger.Printf(l.getPrefixWithPadding() + loggerSeparator + "[INFO] =======================================")
	}
}

// Error prints an error message to the logger
func (l *Logger) Error(message string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	text := fmt.Sprintf(message, args...)
	// #nosec G104 - we are not using user input
	l.logger.Printf(l.getPrefixWithPadding() + loggerSeparator + "[ERROR]" + text + "\n")
}

// Xmsg prints a Xmsg
func (l *Logger) Xmsg(xmsg xmsgtypes.Xmsg, name string) {
	l.Info(" %s cross-chain transaction: %s", name, xmsg.Index)
	if xmsg.XmsgStatus != nil {
		l.Info(" XmsgStatus:")
		l.Info("  Status: %s", xmsg.XmsgStatus.Status.String())
		if xmsg.XmsgStatus.StatusMessage != "" {
			l.Info("  StatusMessage: %s", xmsg.XmsgStatus.StatusMessage)
		}
	}
	if xmsg.InboundTxParams != nil {
		l.Info(" InboundTxParams:")
		l.Info("  TxHash: %s", xmsg.InboundTxParams.InboundTxHash)
		l.Info("  TxHeight: %d", xmsg.InboundTxParams.InboundTxBlockHeight)
		l.Info("  BallotIndex: %s", xmsg.InboundTxParams.InboundTxBallotIndex)
		l.Info("  SenderChainId: %d", xmsg.InboundTxParams.SenderChainId)
		l.Info("  Origin: %s", xmsg.InboundTxParams.TxOrigin)
		if xmsg.InboundTxParams.Sender != "" {
			l.Info("  Sender: %s", xmsg.InboundTxParams.Sender)
		}

	}

	for i, outTxParam := range xmsg.OutboundTxParams {
		if i == 0 {
			l.Info(" OutboundTxParams:")
		} else {
			l.Info(" RevertTxParams:")
		}
		l.Info("  TxHash: %s", outTxParam.OutboundTxHash)
		l.Info("  TxHeight: %d", outTxParam.OutboundTxExternalHeight)
		l.Info("  BallotIndex: %s", outTxParam.OutboundTxBallotIndex)
		l.Info("  TSSNonce: %d", outTxParam.OutboundTxTssNonce)
		l.Info("  GasLimit: %d", outTxParam.OutboundTxGasLimit)
		l.Info("  GasPrice: %s", outTxParam.OutboundTxGasPrice)
		l.Info("  GasUsed: %d", outTxParam.OutboundTxGasUsed)
		l.Info("  EffectiveGasPrice: %s", outTxParam.OutboundTxEffectiveGasPrice.String())
		l.Info("  EffectiveGasLimit: %d", outTxParam.OutboundTxEffectiveGasLimit)
		l.Info("  Receiver: %s", outTxParam.Receiver)
		l.Info("  ReceiverChainId: %d", outTxParam.ReceiverChainId)
	}
}

// EVMTransaction prints a transaction
func (l *Logger) EVMTransaction(tx ethtypes.Transaction, name string) {
	l.Info(" %s EVM transaction: %s", name, tx.Hash().Hex())
	l.Info("  To: %s", tx.To().Hex())
	l.Info("  Value: %d", tx.Value())
	l.Info("  Gas: %d", tx.Gas())
	l.Info("  GasPrice: %d", tx.GasPrice())
}

// EVMReceipt prints a receipt
func (l *Logger) EVMReceipt(receipt ethtypes.Receipt, name string) {
	l.Info(" %s EVM receipt: %s", name, receipt.TxHash.Hex())
	l.Info("  BlockNumber: %d", receipt.BlockNumber)
	l.Info("  GasUsed: %d", receipt.GasUsed)
	l.Info("  ContractAddress: %s", receipt.ContractAddress.Hex())
	l.Info("  Status: %d", receipt.Status)
}

func (l *Logger) getPrefixWithPadding() string {
	// add padding to prefix
	prefix := l.prefix
	for i := len(l.prefix); i < padding; i++ {
		prefix += " "
	}
	return prefix
}
