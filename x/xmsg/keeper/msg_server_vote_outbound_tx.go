package keeper

import (
	"context"
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	relayerkeeper "github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// VoteOnObservedOutboundTx casts a vote on an outbound transaction observed on a connected chain (after
// it has been broadcasted to and finalized on a connected chain). If this is
// the first vote, a new ballot is created. When a threshold of votes is
// reached, the ballot is finalized. When a ballot is finalized, the outbound
// transaction is processed.
//
// If the observation is successful, the difference between pell burned
// and minted is minted by the bank module and deposited into the module
// account.
//
// If the observation is unsuccessful, the logic depends on the previous
// status.
//
// If the previous status was `PendingOutbound`, a new revert transaction is
// created. To cover the revert transaction fee, the required amount of tokens
// submitted with the Xmsg are swapped using a Uniswap V2 contract instance on
// PellChain for the PRC20 of the gas token of the receiver chain. The PRC20
// tokens are then
// burned. The nonce is updated. If everything is successful, the Xmsg status is
// changed to `PendingRevert`.
//
// If the previous status was `PendingRevert`, the Xmsg is aborted.
//
// ```mermaid
// stateDiagram-v2
//
//	state observation <<choice>>
//	state success_old_status <<choice>>
//	state fail_old_status <<choice>>
//	PendingOutbound --> observation: Finalize outbound
//	observation --> success_old_status: Observation succeeded
//	success_old_status --> Reverted: Old status is PendingRevert
//	success_old_status --> OutboundMined: Old status is PendingOutbound
//	observation --> fail_old_status: Observation failed
//	fail_old_status --> PendingRevert: Old status is PendingOutbound
//	fail_old_status --> Aborted: Old status is PendingRevert
//	PendingOutbound --> Aborted: Finalize outbound error
//
// ```
//
// Only observer validators are authorized to broadcast this message.
func (k msgServer) VoteOnObservedOutboundTx(goCtx context.Context, msg *types.MsgVoteOnObservedOutboundTx) (*types.MsgVoteOnObservedOutboundTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the message params to verify it against an existing xmsg
	xmsg, err := k.ValidateOutboundMessage(ctx, *msg)
	if err != nil {
		return nil, err
	}
	// get ballot index
	ballotIndex := msg.Digest()
	// vote on outbound ballot
	isFinalizingVote, isNew, ballot, observationChain, err := k.relayerKeeper.VoteOnOutboundBallot(
		ctx,
		ballotIndex,
		msg.OutTxChain,
		msg.Status,
		msg.Signer)
	if err != nil {
		return nil, err
	}
	// if the ballot is new, set the index to the Xmsg
	if isNew {
		relayerkeeper.EmitEventBallotCreated(ctx, ballot, msg.ObservedOutTxHash, observationChain, relayerkeeper.EventTypeMsgServerVoteOutboundTx)
	}
	// if not finalized commit state here
	if !isFinalizingVote {
		return &types.MsgVoteOnObservedOutboundTxResponse{}, nil
	}

	// if ballot successful, the value received should be the out tx amount
	err = xmsg.AddOutbound(ctx, *msg, ballot.BallotStatus)
	if err != nil {
		return nil, err
	}

	ctx.Logger().Info(fmt.Sprintf("VoteOnObservedOutboundTx xmsgIndex %s processed successfully", xmsg.Index))

	k.processXmsgOutboundResult(ctx, &xmsg, ballot.BallotStatus)
	err = k.ProcessOutbound(ctx, &xmsg, ballot.BallotStatus)
	if err != nil {
		k.SaveFailedOutbound(ctx, &xmsg, err.Error(), ballotIndex)
		return &types.MsgVoteOnObservedOutboundTxResponse{}, nil
	}

	k.SaveSuccessfulOutbound(ctx, &xmsg, ballotIndex)
	return &types.MsgVoteOnObservedOutboundTxResponse{}, nil
}

/*
SaveFailedOutbound saves a failed outbound transaction.It does the following things in one function:

 1. Change the status of the Xmsg to Aborted

 2. Save the outbound
*/

func (k Keeper) SaveFailedOutbound(ctx sdk.Context, xmsg *types.Xmsg, errMessage string, ballotIndex string) {
	xmsg.SetAbort(errMessage)
	ctx.Logger().Error(errMessage)

	k.SaveOutbound(ctx, xmsg, ballotIndex)
}

// SaveSuccessfulOutbound saves a successful outbound transaction.
// This function does not set the Xmsg status, therefore all successful outbound transactions need
// to have their status set during processing
func (k Keeper) SaveSuccessfulOutbound(ctx sdk.Context, xmsg *types.Xmsg, ballotIndex string) {
	k.SaveOutbound(ctx, xmsg, ballotIndex)
}

/*
SaveOutbound saves the outbound transaction.It does the following things in one function:

 1. Set the ballot index for the outbound vote to the xmsg

 2. Remove the nonce from the pending nonces

 3. Remove the outbound tx tracker

 4. Set the xmsg and nonce to xmsg and inTxHash to xmsg
*/
func (k Keeper) SaveOutbound(ctx sdk.Context, xmsg *types.Xmsg, ballotIndex string) {
	receiverChain := xmsg.GetCurrentOutTxParam().ReceiverChainId
	outTxTssNonce := xmsg.GetCurrentOutTxParam().OutboundTxTssNonce

	xmsg.GetCurrentOutTxParam().OutboundTxBallotIndex = ballotIndex
	// #nosec G115 always in range
	for _, outboundParams := range xmsg.GetOutboundTxParams() {
		k.GetRelayerKeeper().RemoveFromPendingNonces(ctx, outboundParams.TssPubkey, outboundParams.ReceiverChainId, int64(outboundParams.OutboundTxTssNonce))
		k.RemoveOutTxTracker(ctx, outboundParams.ReceiverChainId, outboundParams.OutboundTxTssNonce)
		ctx.Logger().Info(fmt.Sprintf("Remove tracker %s: , Block Height: %d ", getOutTrackerIndex(receiverChain, outTxTssNonce), ctx.BlockHeight()))
	}
	// This should set nonce to xmsg only if a new revert is created.
	k.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, *xmsg)
}

func (k Keeper) ValidateOutboundMessage(ctx sdk.Context, msg types.MsgVoteOnObservedOutboundTx) (types.Xmsg, error) {
	// check if Xmsg exists and if the nonce matches
	xmsg, found := k.GetXmsg(ctx, msg.XmsgHash)
	if !found {
		return types.Xmsg{}, cosmoserrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("Xmsg %s does not exist", msg.XmsgHash))
	}
	if xmsg.GetCurrentOutTxParam().OutboundTxTssNonce != msg.OutTxTssNonce {
		return types.Xmsg{}, cosmoserrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("OutTxTssNonce %d does not match Xmsg OutTxTssNonce %d", msg.OutTxTssNonce, xmsg.GetCurrentOutTxParam().OutboundTxTssNonce))
	}
	// do not process an outbound vote if TSS is not found
	_, found = k.relayerKeeper.GetTSS(ctx)
	if !found {
		return types.Xmsg{}, types.ErrCannotFindTSSKeys
	}
	if xmsg.GetCurrentOutTxParam().ReceiverChainId != msg.OutTxChain {
		return types.Xmsg{}, cosmoserrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("OutTxChain %d does not match Xmsg OutTxChain %d", msg.OutTxChain, xmsg.GetCurrentOutTxParam().ReceiverChainId))
	}
	return xmsg, nil
}
