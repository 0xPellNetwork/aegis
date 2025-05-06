package keeper

import (
	"context"
	"fmt"
	"strings"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// MaxOutTxTrackerHashes is the maximum number of hashes that can be stored in the outbound transaction tracker
const MaxOutTxTrackerHashes = 2

// AddToOutTxTracker adds a new record to the outbound transaction tracker.
// only the admin policy account and the observer validators are authorized to broadcast this message without proof.
// If no pending xmsg is found, the tracker is removed, if there is an existed tracker with the nonce & chainID.
func (k msgServer) AddToOutTxTracker(goCtx context.Context, msg *types.MsgAddToOutTxTracker) (*types.MsgAddToOutTxTrackerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check the chain is supported
	chain := k.GetRelayerKeeper().GetSupportedChainFromChainID(ctx, msg.ChainId)
	if chain == nil {
		return nil, relayertypes.ErrSupportedChains
	}

	// the xmsg must exist
	xmsg, err := k.XmsgByNonce(ctx, &types.QueryGetXmsgByNonceRequest{
		ChainId: msg.ChainId,
		Nonce:   msg.Nonce,
	})
	if err != nil {
		return nil, cosmoserrors.Wrap(types.ErrCannotFindXmsg, err.Error())
	}
	if xmsg == nil || xmsg.Xmsg == nil {
		return nil, cosmoserrors.Wrapf(types.ErrCannotFindXmsg, "no corresponding xmsg found for chain %d, nonce %d", msg.ChainId, msg.Nonce)
	}

	// tracker submission is only allowed when the xmsg is pending
	if !IsPending(xmsg.Xmsg) {
		// garbage tracker (for any reason) is harmful to outTx observation and should be removed if it exists
		// it if does not exist, RemoveOutTxTracker is a no-op
		k.RemoveOutTxTracker(ctx, msg.ChainId, msg.Nonce)
		return &types.MsgAddToOutTxTrackerResponse{IsRemoved: true}, nil
	}

	isEmergencyGroup := k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_EMERGENCY)
	isObserver := k.GetRelayerKeeper().IsNonTombstonedObserver(ctx, msg.Signer)
	isProven := false

	// only emergency group and observer can submit tracker without proof
	// if the sender is not from the emergency group or observer, the outbound proof must be provided
	if !(isEmergencyGroup || isObserver) {
		if msg.Proof == nil {
			return nil, cosmoserrors.Wrap(authoritytypes.ErrUnauthorized, fmt.Sprintf("Creator %s", msg.Signer))
		}
		// verify proof when it is provided
		if err := verifyProofAndOutTxBody(ctx, k, msg); err != nil {
			return nil, err
		}

		isProven = true
	}

	// fetch the tracker
	// if the tracker does not exist, initialize a new one
	tracker, found := k.GetOutTxTracker(ctx, msg.ChainId, msg.Nonce)
	hash := types.TxHashList{
		TxHash:   msg.TxHash,
		TxSigner: msg.Signer,
		Proved:   isProven,
	}
	if !found {
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			Index:     "",
			ChainId:   msg.ChainId,
			Nonce:     msg.Nonce,
			HashLists: []*types.TxHashList{&hash},
		})
		return &types.MsgAddToOutTxTrackerResponse{}, nil
	}

	// check if the hash is already in the tracker
	for i, hash := range tracker.HashLists {
		hash := hash
		if strings.EqualFold(hash.TxHash, msg.TxHash) {
			// if the hash is already in the tracker but we have a proof, mark it as proven and only keep this one in the list
			if isProven {
				tracker.HashLists[i].Proved = true
				k.SetOutTxTracker(ctx, tracker)
			}
			return &types.MsgAddToOutTxTrackerResponse{}, nil
		}
	}

	// check if max hashes are reached
	if len(tracker.HashLists) >= MaxOutTxTrackerHashes {
		return nil, types.ErrMaxTxOutTrackerHashesReached.Wrapf(
			"max hashes reached for chain %d, nonce %d, hash number: %d",
			msg.ChainId,
			msg.Nonce,
			len(tracker.HashLists),
		)
	}

	// add the tracker to the list
	tracker.HashLists = append(tracker.HashLists, &hash)
	k.SetOutTxTracker(ctx, tracker)
	return &types.MsgAddToOutTxTrackerResponse{}, nil
}

// verifyProofAndOutTxBody verifies the proof and outbound tx body
// Precondition: the proof must be non-nil
func verifyProofAndOutTxBody(ctx sdk.Context, k msgServer, msg *types.MsgAddToOutTxTracker) error {
	txBytes, err := k.lightclientKeeper.VerifyProof(ctx, msg.Proof, msg.ChainId, msg.BlockHash, msg.TxIndex)
	if err != nil {
		return types.ErrProofVerificationFail.Wrapf(err.Error())
	}

	// get tss address
	var bitcoinChainID int64

	tss, err := k.GetRelayerKeeper().GetTssAddress(ctx, &relayertypes.QueryGetTssAddressRequest{
		BitcoinChainId: bitcoinChainID,
	})
	if err != nil {
		return relayertypes.ErrTssNotFound.Wrapf(err.Error())
	}
	if tss == nil {
		return relayertypes.ErrTssNotFound.Wrapf("tss address nil")
	}

	if err := types.VerifyOutTxBody(*msg, txBytes, *tss); err != nil {
		return types.ErrTxBodyVerificationFail.Wrapf(err.Error())
	}

	return nil
}
