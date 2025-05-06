package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// VoteOnObservedInboundBlock handles voting for a new inbound block.
// If the current block gets finalized before its previous block,
// the vote will fail and the voting status will be rolled back.
// This ensures the proper sequential finalization of blocks.
func (k msgServer) VoteOnObservedInboundBlock(goctx context.Context, msg *types.MsgVoteInboundBlock) (*types.MsgVoteOnObservedInboundBlockResponse, error) {
	ctx := sdk.UnwrapSDKContext(goctx)

	finalized, _, err := k.relayerKeeper.VoteOnInboundBlockBallot(
		ctx,
		int64(msg.BlockProof.ChainId),
		msg.Signer,
		msg.Digest(),
		msg.BlockProof.BlockHash,
	)

	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("VoteOnObservedInboundBlock", "block height",
		msg.BlockProof.BlockHeight, "vote signer", msg.Signer, "finalized", finalized)

	if finalized {
		if err := k.processFinalizedBlock(ctx, msg); err != nil {
			return nil, err
		}
	}

	return &types.MsgVoteOnObservedInboundBlockResponse{}, nil
}

func (k msgServer) processFinalizedBlock(ctx sdk.Context, msg *types.MsgVoteInboundBlock) error {
	var prevBlock *types.BlockProof = nil

	if msg.BlockProof.PrevBlockHeight != 0 {
		block, exist := k.GetBlockProof(ctx, msg.BlockProof.ChainId, msg.BlockProof.PrevBlockHeight)
		if !exist {
			return types.ErrInboundPrevBlockNotFound
		}

		prevBlock = &block
	}

	if _, exist := k.GetBlockProof(ctx, msg.BlockProof.ChainId, msg.BlockProof.BlockHeight); exist {
		return types.ErrBlockProofAlreadyFinalized
	}

	k.SetBlockProof(ctx, msg.BlockProof)
	k.SetChainIndex(ctx, msg.BlockProof.ChainId, msg.BlockProof.BlockHeight)

	EmitEventChainIndex(ctx, msg.BlockProof.ChainId, msg.BlockProof.BlockHeight)

	return k.SetBlockEvents(ctx, msg.BlockProof, prevBlock)
}

// SetBlockEvents creates a doubly-linked list of events from the BlockProof.
// Each event is connected to its previous and next events through their digests,
// forming a chain of events across multiple blocks:
// - For non-first blocks, it links the last event of the previous block to the first event of the current block
// - Within the current block, it links all events sequentially
func (k msgServer) SetBlockEvents(ctx sdk.Context, blockProof *types.BlockProof, prevBlock *types.BlockProof) error {
	prevEventDigest := ""
	if blockProof.PrevBlockHeight != 0 && prevBlock != nil {
		preBlockLastEvent := prevBlock.Events[len(prevBlock.Events)-1]

		prevEventDigest = preBlockLastEvent.Digest

		preBlockLastEventNode, exist := k.GetEventStatusNode(ctx, preBlockLastEvent.Digest)
		if !exist {
			return types.ErrInboundPrevBlockNotFound
		}

		preBlockLastEventNode.NextEventIndex = blockProof.Events[0].Digest
		k.SetEventStatusNode(ctx, preBlockLastEvent.Digest, preBlockLastEventNode)
	}

	proofLen := len(blockProof.Events)

	for i, event := range blockProof.Events {
		nextEventIndex := ""
		if i != proofLen-1 {
			nextEventIndex = blockProof.Events[i+1].Digest
		}

		k.SetEventStatusNode(ctx, event.Digest, types.EventStatusNode{
			PrevEventIndex:    prevEventDigest,
			NextEventIndex:    nextEventIndex,
			EventIndexInBlock: event.Index,
			Status:            types.EventStatus_PENDING,
		})

		prevEventDigest = event.Digest
	}

	return nil
}

// SetXmsg set a specific send in the store from its index
func (k Keeper) SetEventStatusNode(ctx sdk.Context, index string, event types.EventStatusNode) {
	p := types.KeyPrefix(types.InboundEventKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)
	b := k.cdc.MustMarshal(&event)
	store.Set(types.KeyPrefix(index), b)
}

// GetXmsg returns a send from its index
func (k Keeper) GetEventStatusNode(ctx sdk.Context, index string) (val types.EventStatusNode, found bool) {
	p := types.KeyPrefix(types.InboundEventKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)

	b := store.Get(types.KeyPrefix(index))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
