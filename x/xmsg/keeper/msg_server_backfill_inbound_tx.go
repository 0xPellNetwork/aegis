package keeper

// TODO: add api
// // Process all voting events in parallel to ensure sequential execution. Each relayer can vote out of order.
// func (k msgServer) parallelProcessInboundMsg(ctx sdk.Context, msg *types.MsgVoteOnObservedInboundTx) error {
// 	tss, tssFound := k.relayerKeeper.GetTSS(ctx)
// 	if !tssFound {
// 		return types.ErrCannotFindTSSKeys
// 	}

// 	// create a new Xmsg from the inbound message.The status of the new Xmsg is set to PendingInbound.
// 	xmsg, err := types.NewXmsg(ctx, *msg, tss.TssPubkey)
// 	if err != nil {
// 		return err
// 	}

// 	pendingblocks, err := k.findPendingBlocks(ctx, uint64(msg.SenderChainId))
// 	if err != nil {
// 		return err
// 	}

// 	k.processNewXmsg(ctx, pendingblocks, &xmsg, msg.Digest())

// 	xmsgs, err := k.findExecutableXmsgs(ctx, pendingblocks)
// 	if err != nil {
// 		return err
// 	}

// 	for i := range xmsgs {
// 		k.ProcessInbound(ctx, &xmsgs[i])
// 		k.SaveInbound(ctx, &xmsgs[i], xmsgs[i].InboundTxParams.InboundTxBlockHeight, msg.EventIndex)
// 	}
// 	return nil
// }

/* SaveInbound saves the inbound Xmsg to the store.It does the following:
    - Emits an event for the finalized inbound Xmsg.
	- Adds the inbound Xmsg to the finalized inbound Xmsg store.This is done to prevent double spending, using the same inbound tx hash and event index.
	- Updates the Xmsg with the finalized height and finalization status.
	- Removes the inbound Xmsg from the inbound transaction tracker store.This is only for inbounds created via InTx tracker suggestions
	- Sets the Xmsg and nonce to the Xmsg and inbound transaction hash to Xmsg store.
*/

// TODO: unused right now. verity vote transaction has been included in blockProof
// func (k msgServer) verifyVoteTx(ctx sdk.Context, msg *types.MsgVoteOnObservedInboundTx) error {
// 	blockProof, exist := k.GetBlockProof(ctx, uint64(msg.SenderChainId), msg.InBlockHeight)
// 	if !exist {
// 		return types.ErrBlockProofNotFound
// 	}

// 	// TODO: Binary Search
// 	for _, event := range blockProof.Events {
// 		if msg.EventIndex < event.Index {
// 			return types.ErrInvalidInboundTx
// 		}

// 	if msg.EventIndex == event.Index {
// 		if event.PellEvent.String() != msg.PellTx.String() {
// 			return types.ErrInvalidInboundTx
// 		}

// 		return nil
// 	}
// }

// 	// TODO: merkle proof verify
// 	return types.ErrInvalidInboundTx
// }

// TODO: admin backfill
// func (k msgServer) processNewXmsg(ctx sdk.Context, pendingBlocks []*types.BlockProof, newXmsg *types.Xmsg, msgDiest string) {
// 	for i := len(pendingBlocks) - 1; i >= 0; i-- {
// 		for j, event := range pendingBlocks[i].Events {
// 			if event.Status == types.EventStatus_DONE {
// 				continue
// 			}

// 			if newXmsg.InboundTxParams.InboundTxEventIndex == event.Index && newXmsg.InboundTxParams.InboundTxHash == event.TxHash && msgDiest == event.Digest {
// 				// executable
// 				k.ProcessInbound(ctx, newXmsg)
// 				k.SaveInbound(ctx, newXmsg, newXmsg.InboundTxParams.InboundTxBlockHeight, event.Index)

// 				pendingBlocks[i].Events[j].Status = types.EventStatus_DONE
// 				// TODO: store together
// 				k.SetBlockProof(ctx, pendingBlocks[i])
// 			} else {
// 				// cannot excute. store it
// 				k.SetXmsg(ctx, *newXmsg)
// 			}

// 			return
// 		}
// 	}
// }

// // executable xmsgs/ inputXmsg Exists in the executable list
// func (k msgServer) findExecutableXmsgs(ctx sdk.Context, pendingBlocks []*types.BlockProof) ([]types.Xmsg, error) {
// 	res := make([]types.Xmsg, 0)

// 	for i := len(pendingBlocks) - 1; i >= 0; i-- {
// 		writeStore := false

// 		pendingBlock := pendingBlocks[i]
// 		for i, event := range pendingBlock.Events {
// 			if event.Status == types.EventStatus_DONE {
// 				continue
// 			}

// 			xmsg, exist := k.GetXmsg(ctx, event.Digest)
// 			if !exist {
// 				if writeStore {
// 					k.SetBlockProof(ctx, pendingBlock)
// 				}

// 				return res, nil
// 			}

// 			res = append(res, xmsg)

// 			writeStore = true
// 			pendingBlock.Events[i].Status = types.EventStatus_DONE
// 		}

// 		if writeStore {
// 			k.SetBlockProof(ctx, pendingBlock)
// 		}
// 	}

// 	return res, nil
// }

// func (k msgServer) findPendingBlocks(ctx sdk.Context, chainId uint64) ([]*types.BlockProof, error) {
// 	res := make([]*types.BlockProof, 0)
// 	chainIndex, exist := k.GetChainIndex(ctx, chainId)
// 	if !exist {
// 		return nil, nil
// 	}

// 	blockHeight := chainIndex.CurrHeight
// 	for {
// 		blockProof, exist := k.GetBlockProof(ctx, chainId, blockHeight)
// 		if !exist {
// 			break
// 		}

// 		eventLen := len(blockProof.Events)
// 		if blockProof.Events[eventLen-1].Status == types.EventStatus_PENDING {
// 			res = append(res, &blockProof)
// 		} else {
// 			break
// 		}

// 		blockHeight = blockProof.PrevBlockHeight
// 	}

// 	return res, nil
// }
