package pellcore

import (
	"context"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/aegis/pkg/retry"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// MonitorVoteInboundResult monitors the result of a vote inbound tx
// retryGasLimit is the gas limit used to resend the tx if it fails because of insufficient gas
// if retryGasLimit is 0, the tx is not resent
func (c *PellCoreBridge) MonitorVoteInboundTxResult(
	ctx context.Context,
	pellTxHash string,
	retryGasLimit uint64,
	msg []*types.MsgVoteOnObservedInboundTx,
) error {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error().
				Interface("panic", r).
				Str("inbound.hash", pellTxHash).
				Msg("monitorVoteInboundResult: recovered from panic")
		}
	}()

	call := func() error {
		return retry.Retry(c.monitorVoteInboundResult(ctx, pellTxHash, retryGasLimit, msg))
	}

	err := retryWithBackoff(call, monitorRetryCount, monitorInterval/2, monitorInterval)
	if err != nil {
		c.logger.Error().Err(err).
			Str("inbound.hash", pellTxHash).
			Msg("monitorVoteInboundResult: unable to query tx result")

		return err
	}

	return nil
}

func (c *PellCoreBridge) monitorVoteInboundResult(
	ctx context.Context,
	pellTxHash string,
	retryGasLimit uint64,
	msg []*types.MsgVoteOnObservedInboundTx,
) error {
	// query tx result from PellChain
	txResult, err := c.QueryTxResult(pellTxHash)
	if err != nil {
		return errors.Wrap(err, "failed to query tx result")
	}

	logFields := map[string]any{
		"inbound.hash":    pellTxHash,
		"inbound.raw_log": txResult.RawLog,
	}

	switch {
	// Handle retry cases:
	// 1. "out of gas": transaction failed due to insufficient gas
	// 2. "invalid inbound tx": block proof for the inbound tx hasn't been balloted yet, needs retry
	case strings.Contains(txResult.RawLog, "out of gas") || strings.Contains(txResult.RawLog, "sequential error"):
		// if the tx fails with an out of gas error, resend the tx with more gas if retryGasLimit > 0
		c.logger.Debug().Fields(logFields).Any("inbound.raw_log", txResult.RawLog).Msg("monitorVoteInboundResult: error")

		if retryGasLimit > 0 {
			// new retryGasLimit set to 0 to prevent reentering this function
			if resentTxHash, _, err := c.PostVoteInboundEvents(ctx, retryGasLimit, 0, msg); err != nil {
				c.logger.Error().Err(err).Fields(logFields).Msg("monitorVoteInboundResult: failed to resend tx")
			} else {
				logFields["inbound.resent_hash"] = resentTxHash
				c.logger.Info().Fields(logFields).Msgf("monitorVoteInboundResult: successfully resent tx")
			}
		}

	case strings.Contains(txResult.RawLog, "failed to execute message"):
		// the inbound vote tx shouldn't fail to execute
		// this shouldn't happen
		c.logger.Error().Fields(logFields).Msg("monitorVoteInboundResult: failed to execute vote")

	default:
		c.logger.Debug().Fields(logFields).Msgf("monitorVoteInboundResult: successful")
	}

	return nil
}

// MonitorVoteInboundBlockResult monitors the result of a vote inbound tx
// retryGasLimit is the gas limit used to resend the tx if it fails because of insufficient gas
// if retryGasLimit is 0, the tx is not resent
func (c *PellCoreBridge) MonitorVoteInboundBlockResult(
	ctx context.Context,
	pellTxHash string,
	retryGasLimit uint64,
	msg *types.MsgVoteInboundBlock,
	events []*types.MsgVoteOnObservedInboundTx,
) error {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error().
				Interface("panic", r).
				Str("inboundblock.hash", pellTxHash).
				Msg("monitorVoteInboundBlockResult: recovered from panic")
		}
	}()

	call := func() error {
		return retry.Retry(c.monitorVoteInboundBlockResult(ctx, pellTxHash, retryGasLimit, msg, events))
	}

	err := retryWithBackoff(call, monitorRetryCount, monitorInterval/2, monitorInterval)
	if err != nil {
		c.logger.Error().Err(err).
			Str("inboundBlock.hash", pellTxHash).
			Msg("monitorVoteInboundBlockResult: unable to query tx result")

		return err
	}

	return nil
}

func (c *PellCoreBridge) monitorVoteInboundBlockResult(
	ctx context.Context,
	pellTxHash string,
	retryGasLimit uint64,
	msg *types.MsgVoteInboundBlock,
	events []*types.MsgVoteOnObservedInboundTx,
) error {
	// query tx result from PellChain
	txResult, err := c.QueryTxResult(pellTxHash)
	if err != nil {
		return errors.Wrap(err, "failed to query tx result")
	}

	logFields := map[string]any{
		"inbound.hash":    pellTxHash,
		"inbound.raw_log": txResult.RawLog,
	}

	switch {
	// Handle retry cases:
	// 1. "out of gas": transaction failed due to insufficient gas
	// 2. "invalid inbound tx": block proof for the inbound tx is not in sequence, needs retry
	case strings.Contains(txResult.RawLog, "out of gas") || strings.Contains(txResult.RawLog, "sequential error"):
		// if the tx fails with an out of gas error, resend the tx with more gas if retryGasLimit > 0
		c.logger.Debug().Fields(logFields).Any("inbound.raw_log", txResult.RawLog).Msg("monitorVoteInboundBlockResult: error")

		if retryGasLimit > 0 {
			// new retryGasLimit set to 0 to prevent reentering this function
			if resentTxHash, _, err := c.PostVoteInboundBlock(ctx, retryGasLimit, 0, msg, events); err != nil {
				c.logger.Error().Err(err).Fields(logFields).Msg("monitorVoteInboundBlockResult: failed to resend tx")
			} else {
				logFields["inbound.resent_hash"] = resentTxHash
				c.logger.Info().Fields(logFields).Msgf("monitorVoteInboundBlockResult: successfully resent tx")
			}
		}

	case strings.Contains(txResult.RawLog, "failed to execute message"):
		// the inbound vote tx shouldn't fail to execute
		// this shouldn't happen
		c.logger.Error().Fields(logFields).Msg("monitorVoteInboundBlockResult: failed to execute vote")
	default:
		c.logger.Debug().Fields(logFields).Msgf("monitorVoteInboundBlockResult: successful")
	}

	return nil
}

// MonitorVoteOutboundResult monitors the result of a vote outbound tx
// retryGasLimit is the gas limit used to resend the tx if it fails because of insufficient gas
// if retryGasLimit is 0, the tx is not resent
func (c *PellCoreBridge) MonitorVoteOutboundTxResult(
	ctx context.Context,
	pellTxHash string,
	retryGasLimit uint64,
	msg *types.MsgVoteOnObservedOutboundTx,
) error {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error().
				Interface("panic", r).
				Str("outbound.hash", pellTxHash).
				Msg("monitorVoteOutboundResult: recovered from panic")
		}
	}()

	call := func() error {
		return retry.Retry(c.monitorVoteOutboundResult(ctx, pellTxHash, retryGasLimit, msg))
	}

	err := retryWithBackoff(call, monitorRetryCount, monitorInterval/2, monitorInterval)
	if err != nil {
		c.logger.Error().Err(err).
			Str("outbound.hash", pellTxHash).
			Msg("monitorVoteOutboundResult: unable to query tx result")

		return err
	}

	return nil
}

func (c *PellCoreBridge) monitorVoteOutboundResult(
	ctx context.Context,
	pellTxHash string,
	retryGasLimit uint64,
	msg *types.MsgVoteOnObservedOutboundTx,
) error {
	// query tx result from PellChain
	txResult, err := c.QueryTxResult(pellTxHash)
	if err != nil {
		return errors.Wrap(err, "failed to query tx result")
	}

	logFields := map[string]any{
		"outbound.hash":    pellTxHash,
		"outbound.raw_log": txResult.RawLog,
	}

	switch {
	case strings.Contains(txResult.RawLog, "failed to execute message"):
		// the inbound vote tx shouldn't fail to execute
		// this shouldn't happen
		c.logger.Error().Fields(logFields).Msg("monitorVoteOutboundResult: failed to execute vote")
	case strings.Contains(txResult.RawLog, "out of gas"):
		// if the tx fails with an out of gas error, resend the tx with more gas if retryGasLimit > 0
		c.logger.Debug().Fields(logFields).Msg("monitorVoteOutboundResult: out of gas")

		if retryGasLimit > 0 {
			// new retryGasLimit set to 0 to prevent reentering this function
			if _, _, err := c.PostVoteOutboundFromMsg(ctx, retryGasLimit, 0, msg); err != nil {
				c.logger.Error().Err(err).Fields(logFields).Msg("monitorVoteOutboundResult: failed to resend tx")
			} else {
				c.logger.Info().Fields(logFields).Msg("monitorVoteOutboundResult: successfully resent tx")
			}
		}
	default:
		c.logger.Debug().Fields(logFields).Msg("monitorVoteOutboundResult: successful")
	}

	return nil
}

func retryWithBackoff(call func() error, attempts int, minInternal, maxInterval time.Duration) error {
	if attempts < 1 {
		return errors.New("attempts must be positive")
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = minInternal
	bo.MaxInterval = maxInterval

	backoffWithRetry := backoff.WithMaxRetries(bo, uint64(attempts))

	return retry.DoWithBackoff(call, backoffWithRetry)
}
