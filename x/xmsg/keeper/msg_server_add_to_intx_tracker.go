package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// AddToInTxTracker adds a new record to the inbound transaction tracker.
func (k msgServer) AddToInTxTracker(goCtx context.Context, msg *types.MsgAddToInTxTracker) (*types.MsgAddToInTxTrackerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	chain := k.GetRelayerKeeper().GetSupportedChainFromChainID(ctx, msg.ChainId)
	if chain == nil {
		return nil, observertypes.ErrSupportedChains
	}

	// emergency or observer group can submit tracker without proof
	isEmergencyGroup := k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_EMERGENCY)
	isObserver := k.GetRelayerKeeper().IsNonTombstonedObserver(ctx, msg.Signer)

	// only emergency group and observer can submit tracker without proof
	// if the sender is not from the emergency group or observer, the inbound proof must be provided
	if !(isEmergencyGroup || isObserver) {
		if msg.Proof == nil {
			return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, fmt.Sprintf("Creator %s", msg.Signer))
		}

		// verify the proof and tx body
		if err := verifyProofAndInTxBody(ctx, k, msg); err != nil {
			return nil, err
		}
	}

	// add the inTx tracker
	k.SetInTxTracker(ctx, types.InTxTracker{
		ChainId: msg.ChainId,
		TxHash:  msg.TxHash,
	})

	return &types.MsgAddToInTxTrackerResponse{}, nil
}

// verifyProofAndInTxBody verifies the proof and inbound tx body
func verifyProofAndInTxBody(ctx sdk.Context, k msgServer, msg *types.MsgAddToInTxTracker) error {
	txBytes, err := k.GetLightclientKeeper().VerifyProof(ctx, msg.Proof, msg.ChainId, msg.BlockHash, msg.TxIndex)
	if err != nil {
		return types.ErrProofVerificationFail.Wrapf(err.Error())
	}

	// get chain params and tss addresses to verify the inTx body
	chainParams, found := k.GetRelayerKeeper().GetChainParamsByChainID(ctx, msg.ChainId)
	if !found || chainParams == nil {
		return types.ErrUnsupportedChain.Wrapf("chain params not found for chain %d", msg.ChainId)
	}
	tss, err := k.GetRelayerKeeper().GetTssAddress(ctx, &observertypes.QueryGetTssAddressRequest{
		BitcoinChainId: msg.ChainId,
	})
	if err != nil {
		return observertypes.ErrTssNotFound.Wrapf(err.Error())
	}
	if tss == nil {
		return observertypes.ErrTssNotFound.Wrapf("tss address nil")
	}

	if err := types.VerifyInTxBody(*msg, txBytes, *chainParams, *tss); err != nil {
		return types.ErrTxBodyVerificationFail.Wrapf(err.Error())
	}

	return nil
}
