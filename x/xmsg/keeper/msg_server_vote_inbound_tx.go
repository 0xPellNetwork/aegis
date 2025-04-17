package keeper

import (
	"context"
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const XMSG_EVENT_INDEX = "%d-%d"

// FIXME: use more specific error types & codes

// VoteOnObservedInboundTx casts a vote on an inbound transaction observed on a connected chain. If this
// is the first vote, a new ballot is created. When a threshold of votes is
// reached, the ballot is finalized. When a ballot is finalized, a new Xmsg is
// created.
//
// If the receiver chain is PellChain, `HandleEVMDeposit` is called. If the
// tokens being deposited are PELL, `MintPellToEVMAccount` is called and the
// tokens are minted to the receiver account on PellChain. If the tokens being
// deposited are gas tokens or ERC20 of a connected chain, ZRC20's `deposit`
// method is called and the tokens are deposited to the receiver account on
// PellChain. If the message is not empty, system contract's `depositAndCall`
// method is also called and an omnichain contract on PellChain is executed.
// Omnichain contract address and arguments are passed as part of the message.
// If everything is successful, the Xmsg status is changed to `OutboundMined`.
//
// If the receiver chain is a connected chain, the `FinalizeInbound` method is
// called to prepare the Xmsg to be processed as an outbound transaction. To
// cover the outbound transaction fee, the required amount of tokens submitted
// with the Xmsg are swapped using a Uniswap V2 contract instance on PellChain
// for the ZRC20 of the gas token of the receiver chain. The ZRC20 tokens are
// then burned. The nonce is updated. If everything is successful, the Xmsg
// status is changed to `PendingOutbound`.
//
// ```mermaid
// stateDiagram-v2
//
//	state evm_deposit_success <<choice>>
//	state finalize_inbound <<choice>>
//	state evm_deposit_error <<choice>>
//	PendingInbound --> evm_deposit_success: Receiver is PellChain
//	evm_deposit_success --> OutboundMined: EVM deposit success
//	evm_deposit_success --> evm_deposit_error: EVM deposit error
//	evm_deposit_error --> PendingRevert: Contract error
//	evm_deposit_error --> Aborted: Internal error, invalid chain, gas, nonce
//	PendingInbound --> finalize_inbound: Receiver is connected chain
//	finalize_inbound --> Aborted: Finalize inbound error
//	finalize_inbound --> PendingOutbound: Finalize inbound success
//
// ```
//
// Only observer validators are authorized to broadcast this message.
func (k msgServer) VoteOnObservedInboundTx(goCtx context.Context, msg *types.MsgVoteOnObservedInboundTx) (*types.MsgVoteOnObservedInboundTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// verify vote tx in the block proof
	index := msg.Digest()

	ballotFinalized, err := k.processInboundTxBallot(ctx, msg, index)
	if err != nil || !ballotFinalized {
		return &types.MsgVoteOnObservedInboundTxResponse{}, err
	}

	if err := k.processInboundEvent(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgVoteOnObservedInboundTxResponse{}, nil
}

// vote on inbound ballot
// use a temporary context to not commit any ballot state change in case of error
// If it is a new ballot, check if an inbound with the same hash, sender chain and event index has already been finalized
// This may happen if the same inbound is observed twice where msg.Digest gives a different index
// This check prevents double spending
func (k msgServer) processInboundTxBallot(ctx sdk.Context, msg *types.MsgVoteOnObservedInboundTx, index string) (bool, error) {
	tmpCtx, commit := ctx.CacheContext()
	finalized, isNew, err := k.relayerKeeper.VoteOnInboundBallot(
		tmpCtx,
		msg.SenderChainId,
		msg.ReceiverChain,
		1, // todo: remove
		msg.Signer,
		index,
		msg.InTxHash,
	)
	if err != nil {
		return false, err
	}

	if isNew && k.IsFinalizedInbound(tmpCtx, msg.InTxHash, msg.SenderChainId, msg.EventIndex) {
		return false, cosmoserrors.Wrap(
			types.ErrObservedTxAlreadyFinalized,
			fmt.Sprintf("InTxHash:%s, SenderChainId:%d, EventIndex:%d", msg.InTxHash, msg.SenderChainId, msg.EventIndex),
		)
	}

	commit()

	return finalized, nil
}

// findExecutableEvents sequentially traverses event nodes starting from the given node.
// For each subsequent event in the waiting queue (stored in XmsgByEventIndex),
// it executes the corresponding xmsg. The traversal stops when either:
// - the next event doesn't exist
// - the next event's status is not PENDING
// - the xmsg for the next event index cannot be found
// Returns the list of executable xmsgs and their corresponding event indices in the block.
func (k msgServer) findExecutableEvents(ctx sdk.Context, msg *types.EventStatusNode) ([]types.Xmsg, []uint64, error) {
	res := []types.Xmsg{}
	eventIndexInBlock := []uint64{}
	nextEventIndex := msg.NextEventIndex

	for nextEventIndex != "" {
		nextEventNode, exist := k.GetEventStatusNode(ctx, nextEventIndex)
		if !exist || nextEventNode.Status != types.EventStatus_PENDING {
			break
		}

		// find xmsg by pending queue
		xmsg, exist := k.GetXmsgByEventIndex(ctx, nextEventIndex)
		if !exist {
			break
		}

		k.SetEventStatusNode(ctx, nextEventIndex, types.EventStatusNode{
			PrevEventIndex:    nextEventNode.PrevEventIndex,
			NextEventIndex:    nextEventNode.NextEventIndex,
			EventIndexInBlock: nextEventNode.EventIndexInBlock,
			Status:            types.EventStatus_DONE,
		})

		k.DeleteXmsgByEventIndex(ctx, nextEventIndex)

		res = append(res, xmsg)

		eventIndexInBlock = append(eventIndexInBlock, nextEventNode.EventIndexInBlock)

		nextEventIndex = nextEventNode.NextEventIndex
	}

	return res, eventIndexInBlock, nil
}

// Process inbound events sequentially. First checks if the current inbound event can be executed
// (previous event must be successfully executed). Then finds and processes all subsequent executable
// events in the queue. An event is executable only if its previous event is confirmed as DONE.
func (k msgServer) processInboundEvent(ctx sdk.Context, msg *types.MsgVoteOnObservedInboundTx) error {
	eventNode, exist := k.GetEventStatusNode(ctx, msg.Digest())
	// block proof for the inbound tx hasn't been ballot
	if !exist {
		k.Logger(ctx).Warn("block proof for the inbound tx hasn't been ballot.", "event_index", msg.Digest(), "block height", msg.InBlockHeight)

		return cosmoserrors.Wrap(
			types.ErrInboundPrevEventNotFound,
			fmt.Sprintf("InTxHash:%s, SenderChainId:%d, EventIndex:%d", msg.InTxHash, msg.SenderChainId, msg.EventIndex),
		)
	}

	tss, tssFound := k.relayerKeeper.GetTSS(ctx)
	if !tssFound {
		return types.ErrCannotFindTSSKeys
	}

	// create a new Xmsg from the inbound message.The status of the new Xmsg is set to PendingInbound.
	xmsg, err := types.NewXmsg(ctx, *msg, tss.TssPubkey)
	if err != nil {
		return err
	}

	if eventNode.PrevEventIndex != "" {
		prevEventNode, exist := k.GetEventStatusNode(ctx, eventNode.PrevEventIndex)
		if !exist || prevEventNode.Status != types.EventStatus_DONE {
			// for backfill. store eventIndex -> xmsg
			k.SetXmsgByEventIndex(ctx, msg.Digest(), xmsg)

			return nil
		}
	}

	k.Logger(ctx).Info("process inbound event", "event_index", msg.EventIndex)

	k.ProcessInbound(ctx, &xmsg)
	k.SaveInbound(ctx, &xmsg, xmsg.InboundTxParams.InboundTxBlockHeight, msg.EventIndex)

	eventNode.Status = types.EventStatus_DONE
	k.SetEventStatusNode(ctx, msg.Digest(), eventNode)

	EmitEventStatusNode(ctx, &eventNode)
	// find available execute event
	availableXmsgs, eventIndexInBlock, err := k.findExecutableEvents(ctx, &eventNode)
	if err != nil {
		return err
	}

	for i := range availableXmsgs {
		k.ProcessInbound(ctx, &availableXmsgs[i])
		k.SaveInbound(ctx, &availableXmsgs[i], availableXmsgs[i].InboundTxParams.InboundTxBlockHeight, eventIndexInBlock[i])
	}

	return nil
}

func (k Keeper) SaveInbound(ctx sdk.Context, xmsg *types.Xmsg, blockHeight uint64, eventIndex uint64) {
	EmitEventInboundFinalized(ctx, xmsg)
	k.AddFinalizedInbound(ctx,
		xmsg.GetInboundTxParams().InboundTxHash,
		xmsg.GetInboundTxParams().SenderChainId,
		eventIndex)
	// #nosec G701 always positive
	xmsg.InboundTxParams.InboundTxFinalizedPellHeight = uint64(ctx.BlockHeight())
	xmsg.InboundTxParams.TxFinalizationStatus = types.TxFinalizationStatus_EXECUTED
	k.RemoveInTxTrackerIfExists(ctx, xmsg.InboundTxParams.SenderChainId, xmsg.InboundTxParams.InboundTxHash)
	k.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, *xmsg)
}
