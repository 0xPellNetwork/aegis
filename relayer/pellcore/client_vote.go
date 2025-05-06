package pellcore

import (
	"context"
	"math/big"
	"strings"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/tss/go-tss/blame"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	"github.com/0xPellNetwork/aegis/pkg/retry"
	pctx "github.com/0xPellNetwork/aegis/relayer/context"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// PostVoteBlockHeader posts a vote on an observed block header
func (c *PellCoreBridge) PostVoteBlockHeader(
	ctx context.Context,
	chainID int64,
	blockHash []byte,
	height int64,
	header proofs.HeaderData,
) (string, error) {
	signerAddress := c.keys.GetOperatorAddress().String()

	msg := relayertypes.NewMsgVoteBlockHeader(signerAddress, chainID, blockHash, height, header)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	pellTxHash, err := retry.DoTypedWithRetry(func() (string, error) {
		return c.Broadcast(ctx, DefaultGasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})

	if err != nil {
		return "", errors.Wrap(err, "unable to broadcast vote block header")
	}

	return pellTxHash, nil
}

// PostVoteGasPrice posts a gas price vote. Returns txHash and error.
func (b *PellCoreBridge) PostGasPrice(
	ctx context.Context,
	chain chains.Chain,
	gasPrice uint64,
	supply string,
	blockNum uint64,
) (string, error) {
	// apply gas price multiplier for the chain
	multiplier := GasPriceMultiplier(chain.Id)

	// #nosec G701 always in range
	gasPrice = uint64(float64(gasPrice) * multiplier)
	signerAddress := b.keys.GetOperatorAddress().String()
	msg := types.NewMsgVoteGasPrice(signerAddress, chain.Id, gasPrice, supply, blockNum)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	hash, err := retry.DoTypedWithRetry(func() (string, error) {
		return b.Broadcast(ctx, PostGasPriceGasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})

	if err != nil {
		return "", errors.Wrap(err, "unable to broadcast vote gas price")
	}

	return hash, nil
}

// PostVoteTSS sends message to vote TSS. Returns txHash and error.
func (c *PellCoreBridge) PostVoteTSS(
	ctx context.Context,
	tssPubKey string,
	keyGenPellHeight int64,
	status chains.ReceiveStatus,
) (string, error) {
	signerAddress := c.keys.GetOperatorAddress().String()
	msg := relayertypes.NewMsgVoteTSS(signerAddress, tssPubKey, keyGenPellHeight, status)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	pellTxHash, err := retry.DoTypedWithRetry(func() (string, error) {
		return c.Broadcast(ctx, PostTSSGasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})

	if err != nil {
		return "", errors.Wrap(err, "unable to broadcast vote for setting tss")
	}

	return pellTxHash, nil
}

// PostVoteBlameData posts blame data message to pellcore. Returns txHash and error.
func (b *PellCoreBridge) PostBlameData(
	ctx context.Context,
	blame *blame.Blame,
	chainID int64,
	index string,
) (string, error) {
	signerAddress := b.keys.GetOperatorAddress().String()
	pellBlame := relayertypes.Blame{
		Index:         index,
		FailureReason: blame.FailReason,
		Nodes:         relayertypes.ConvertNodes(blame.BlameNodes),
	}
	msg := relayertypes.NewMsgAddBlameVoteMsg(signerAddress, chainID, pellBlame)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	var gasLimit uint64 = PostBlameDataGasLimit

	pellTxHash, err := retry.DoTypedWithRetry(func() (string, error) {
		return b.Broadcast(ctx, gasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})

	if err != nil {
		return "", errors.Wrap(err, "unable to broadcast blame data")
	}

	return pellTxHash, nil
}

// PostVoteInbound posts a vote on an observed inbound tx
// retryGasLimit is the gas limit used to resend the tx if it fails because of insufficient gas
// it is used when the ballot is finalized and the inbound tx needs to be processed
func (c *PellCoreBridge) PostVoteInboundEvents(
	ctx context.Context,
	gasLimit, retryGasLimit uint64,
	msg []*types.MsgVoteOnObservedInboundTx,
) (string, string, error) {
	if len(msg) == 0 {
		return "", "", nil
	}

	authzMsgs := make([]sdktypes.Msg, 0)
	_, authzSigner, err := WrapMessageWithAuthz(msg[0])
	if err != nil {
		return "", "", err
	}

	for _, m := range msg {
		authzMsg, _, err := WrapMessageWithAuthz(m)
		if err != nil {
			return "", "", err
		}

		authzMsgs = append(authzMsgs, authzMsg)
	}

	// don't post send if has already voted before
	ballotIndex := msg[0].Digest()
	hasVoted, err := c.HasVoted(ctx, ballotIndex, msg[0].Signer)
	if err != nil {
		return "", ballotIndex, errors.Wrapf(err,
			"PostVoteInbound: unable to check if already voted for ballot %s voter %s",
			ballotIndex,
			msg[0].Signer,
		)
	}

	if hasVoted {
		return "", ballotIndex, nil
	}

	pellTxHash, err := retry.DoTypedWithRetry(func() (string, error) {
		return c.Broadcast(ctx, gasLimit, authzMsgs, authzSigner)
	})

	c.logger.Info().Msgf("PostVoteInboundEvents: pellTxHash: %s, ballotIndex: %s, blockheight: %d", pellTxHash, ballotIndex, msg[0].InBlockHeight)

	if err != nil {
		return "", ballotIndex, errors.Wrap(err, "unable to broadcast vote inbound")
	}

	go func() {
		ctxForWorker := pctx.Copy(ctx, context.Background())

		errMonitor := c.MonitorVoteInboundTxResult(ctxForWorker, pellTxHash, retryGasLimit, msg)
		if errMonitor != nil {
			c.logger.Error().Err(err).Msg("PostVoteInbound: failed to monitor vote inbound result")
		}
	}()

	return pellTxHash, ballotIndex, nil
}

// PostVoteOutbound posts a vote on an observed outbound tx from a MsgVoteOutbound.
// Returns tx hash, ballotIndex, and error.
func (b *PellCoreBridge) PostVoteOutbound(
	ctx context.Context,
	xmsgIndex string,
	outTxHash string,
	outBlockHeight uint64,
	outTxGasUsed uint64,
	outTxEffectiveGasPrice *big.Int,
	outTxEffectiveGasLimit uint64,
	status chains.ReceiveStatus,
	failedReasonMsg string,
	chain chains.Chain,
	nonce uint64,
) (string, string, error) {
	signerAddress := b.keys.GetOperatorAddress().String()
	msg := types.NewMsgVoteOnObservedOutboundTx(
		signerAddress,
		xmsgIndex,
		outTxHash,
		outBlockHeight,
		outTxGasUsed,
		math.NewIntFromBigInt(outTxEffectiveGasPrice),
		outTxEffectiveGasLimit,
		status,
		failedReasonMsg,
		chain.Id,
		nonce,
	)

	// when an outbound fails and a revert is required, the gas limit needs to be higher
	// this is because the revert tx needs to interact with the EVM to perform swaps for the gas token
	// the higher gas limit is only necessary when the vote is finalized and the outbound is processed
	// therefore we use a retryGasLimit with a higher value to resend the tx if it fails (when the vote is finalized)
	retryGasLimit := uint64(0)
	if msg.Status == chains.ReceiveStatus_FAILED {
		retryGasLimit = PostVoteOutboundRevertGasLimit
	}

	return b.PostVoteOutboundFromMsg(ctx, PostVoteOutboundGasLimit, retryGasLimit, msg)
}

// PostVoteOutboundFromMsg posts a vote on an observed outbound tx from a MsgVoteOnObservedOutboundTx
func (b *PellCoreBridge) PostVoteOutboundFromMsg(
	ctx context.Context,
	gasLimit, retryGasLimit uint64,
	msg *types.MsgVoteOnObservedOutboundTx,
) (string, string, error) {
	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to wrap message with authz")
	}

	// don't post confirmation if it  already voted before
	ballotIndex := msg.Digest()
	hasVoted, err := b.HasVoted(ctx, ballotIndex, msg.Signer)
	if err != nil {
		return "", ballotIndex, errors.Wrapf(
			err,
			"PostVoteOutbound: unable to check if already voted for ballot %s voter %s",
			ballotIndex,
			msg.Signer,
		)
	}

	if hasVoted {
		return "", ballotIndex, nil
	}

	pellTxHash, err := retry.DoTypedWithRetry(func() (string, error) {
		return b.Broadcast(ctx, gasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})

	if err != nil {
		return "", ballotIndex, errors.Wrap(err, "unable to broadcast vote outbound")
	}

	go func() {
		ctxForWorker := pctx.Copy(ctx, context.Background())

		errMonitor := b.MonitorVoteOutboundTxResult(ctxForWorker, pellTxHash, retryGasLimit, msg)
		if errMonitor != nil {
			b.logger.Error().Err(err).Msg("PostVoteOutbound: failed to monitor vote outbound result")
		}
	}()
	return pellTxHash, ballotIndex, nil
}

func (b *PellCoreBridge) PostAddTxHashToOutTxTracker(
	ctx context.Context,
	chainID int64,
	nonce uint64,
	txHash string,
	proof *proofs.Proof,
	blockHash string,
	txIndex int64,
) (string, error) {
	// don't report if the tracker already contains the txHash
	tracker, err := b.GetOutTxTracker(ctx, chains.Chain{Id: chainID}, nonce)
	if err == nil {
		for _, hash := range tracker.HashLists {
			if strings.EqualFold(hash.TxHash, txHash) {
				return "", nil
			}
		}
	}

	signerAddress := b.keys.GetOperatorAddress().String()
	msg := types.NewMsgAddToOutTxTracker(signerAddress, chainID, nonce, txHash, proof, blockHash, txIndex)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	pellTxHash, err := b.Broadcast(ctx, AddTxHashToOutTxTrackerGasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	if err != nil {
		return "", err
	}

	return pellTxHash, nil
}

// PostVoteInboundBlock posts a vote on an observed inbound block
// When there are more than 14 events (since one message slot is reserved for the block message,
// leaving 14 slots out of the MAX_MSG_LENGTH=15), the events are processed in batches:
// 1. First batch includes the block message and up to 14 events
// 2. Subsequent batches contain up to 15 events each until all events are processed
func (b *PellCoreBridge) PostVoteInboundBlock(ctx context.Context, gasLimit, retryLimit uint64, block *types.MsgVoteInboundBlock, events []*types.MsgVoteOnObservedInboundTx) ([]string, []string, error) {
	resHash := make([]string, 0)
	resBallotIndex := make([]string, 0)
	if len(events) == 0 {
		return resHash, resBallotIndex, nil
	}

	ballotIndex := block.Digest()
	hasVoted, err := b.HasVoted(ctx, ballotIndex, block.Signer)
	if err != nil {
		return resHash, resBallotIndex, errors.Wrapf(err,
			"PostVoteInboundBlock: unable to check if already voted for ballot %s voter %s",
			ballotIndex,
			block.Signer,
		)
	}

	if hasVoted {
		return resHash, resBallotIndex, nil
	}

	firstBatchSize := min(len(events), int(b.pellTxMsgLength)-1)

	if !hasVoted {
		firstMsg, ballotIndex, err := b.PostVoteInboundFirstPart(ctx, gasLimit, retryLimit, block, events[:firstBatchSize])
		if err != nil {
			b.logger.Error().Err(err).Msg("PostVoteInboundBlock: failed to post vote inbound first part")
		}

		resHash = append(resHash, firstMsg)
		resBallotIndex = append(resBallotIndex, ballotIndex)
	}

	infoLog := b.logger.Info().
		Str("ballotIndex", ballotIndex).
		Uint64("blockHeight", block.GetBlockProof().BlockHeight).
		Uint64("chain", block.GetBlockProof().ChainId)

	infoLog.Msgf("PostVoteInboundBlock: first part processed with %d events left in block %d on chain %d", len(events)-firstBatchSize, block.GetBlockProof().BlockHeight, block.GetBlockProof().ChainId)

	batchSize := int(b.pellTxMsgLength) - 1
	for i := firstBatchSize; i < len(events); i += batchSize {
		end := min(i+batchSize, len(events))

		voteTxHash, ballotIndex, err := b.PostVoteInboundEvents(ctx, gasLimit, retryLimit, events[i:end])
		if err != nil {
			return resHash, resBallotIndex, err
		}

		infoLog.Msgf("PostVoteInboundBlock: subsequent part processed with %d events left in block %d on chain %d", len(events)-end, block.GetBlockProof().BlockHeight, block.GetBlockProof().ChainId)

		resHash = append(resHash, voteTxHash)
		resBallotIndex = append(resBallotIndex, ballotIndex)
	}

	return resHash, resBallotIndex, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b *PellCoreBridge) PostVoteInboundFirstPart(ctx context.Context, gasLimit, retryLimit uint64, block *types.MsgVoteInboundBlock, events []*types.MsgVoteOnObservedInboundTx) (string, string, error) {
	if len(events) == 0 {
		return "", "", nil
	}
	msgs := make([]sdktypes.Msg, 0)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(block)
	if err != nil {
		return "", "", err
	}

	msgs = append(msgs, authzMsg)

	ballotIndex := block.Digest()
	hasVoted, err := b.HasVoted(ctx, ballotIndex, block.Signer)
	if err != nil {
		return "", ballotIndex, errors.Wrapf(err,
			"PostVoteInboundEvents: unable to check if already voted for ballot %s voter %s",
			ballotIndex,
			block.Signer,
		)
	}

	if hasVoted {
		b.logger.Info().Msgf("PostVoteInboundFirstPart: already voted for ballot %s", ballotIndex)
		return "", ballotIndex, nil
	}

	for _, v := range events {
		authzMsg, _, err := WrapMessageWithAuthz(v)
		if err != nil {
			return "", ballotIndex, err
		}

		msgs = append(msgs, authzMsg)
	}

	pellTxHash, err := retry.DoTypedWithRetry(func() (string, error) {
		return b.Broadcast(ctx, gasLimit, msgs, authzSigner)
	})

	b.logger.Info().Msgf("PostVoteInboundFirstPart: pellTxHash: %s, ballotIndex: %s, blockheight: %d", pellTxHash, ballotIndex, block.BlockProof.BlockHeight)

	if err != nil {
		return "", ballotIndex, errors.Wrap(err, "unable to broadcast vote inbound")
	}

	go func() {
		ctxForWorker := pctx.Copy(ctx, context.Background())

		errMonitor := b.MonitorVoteInboundBlockResult(ctxForWorker, pellTxHash, retryLimit, block, events)
		if errMonitor != nil {
			b.logger.Error().Err(err).Msg("PostVoteInboundBlock: failed to monitor vote inbound result")
		}
	}()

	return pellTxHash, ballotIndex, nil
}

// PostVoteAddPellToken posts a vote to add pell token to a chain
func (b *PellCoreBridge) PostVoteOnPellRecharge(ctx context.Context, chain chains.Chain, voteIndex uint64) (string, error) {
	signerAddress := b.keys.GetOperatorAddress().String()
	msg := types.NewMsgVoteOnPellRecharge(signerAddress, chain.Id, voteIndex)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	hash, err := retry.DoTypedWithRetry(func() (string, error) {
		return b.Broadcast(ctx, PostAddPellTokenGasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to broadcast vote add pell token")
	}

	return hash, nil
}

// PostVoteAddGasToken posts a vote to add gas token to a chain
func (b *PellCoreBridge) PostVoteOnGasRecharge(ctx context.Context, chain chains.Chain, voteIndex uint64) (string, error) {
	signerAddress := b.keys.GetOperatorAddress().String()
	msg := types.NewMsgVoteOnGasRecharge(signerAddress, chain.Id, voteIndex)

	authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
	if err != nil {
		return "", err
	}

	hash, err := retry.DoTypedWithRetry(func() (string, error) {
		return b.Broadcast(ctx, PostAddGasTokenGasLimit, []sdktypes.Msg{authzMsg}, authzSigner)
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to broadcast vote add gas token")
	}

	return hash, nil
}
