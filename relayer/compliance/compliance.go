package compliance

import (
	"github.com/rs/zerolog"

	"github.com/0xPellNetwork/aegis/relayer/config"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// IsXmsgRestricted returns true if the xmsg involves restricted addresses
func IsXmsgRestricted(xmsg *xmsgtypes.Xmsg) bool {
	sender := xmsg.InboundTxParams.Sender
	receiver := xmsg.GetCurrentOutTxParam().Receiver
	return config.ContainRestrictedAddress(sender, receiver)
}

// PrintComplianceLog prints compliance log with fields [chain, xmsg/intx, chain, sender, receiver, token]
func PrintComplianceLog(
	logger1 zerolog.Logger,
	logger2 zerolog.Logger,
	outbound bool,
	chainID int64,
	identifier, sender, receiver, token string) {
	var logMsg string
	var logWithFields1 zerolog.Logger
	var logWithFields2 zerolog.Logger
	if outbound {
		// we print xmsg for outbound tx
		logMsg = "Restricted address detected in xmsg"
		logWithFields1 = logger1.With().Int64("chain", chainID).Str("xmsg", identifier).Str("sender", sender).Str("receiver", receiver).Str("token", token).Logger()
		logWithFields2 = logger2.With().Int64("chain", chainID).Str("xmsg", identifier).Str("sender", sender).Str("receiver", receiver).Str("token", token).Logger()
	} else {
		// we print intx for inbound tx
		logMsg = "Restricted address detected in intx"
		logWithFields1 = logger1.With().Int64("chain", chainID).Str("intx", identifier).Str("sender", sender).Str("receiver", receiver).Str("token", token).Logger()
		logWithFields2 = logger2.With().Int64("chain", chainID).Str("intx", identifier).Str("sender", sender).Str("receiver", receiver).Str("token", token).Logger()
	}
	logWithFields1.Warn().Msg(logMsg)
	logWithFields2.Warn().Msg(logMsg)
}
