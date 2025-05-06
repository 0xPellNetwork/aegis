package keeper

import (
	"fmt"
	"sort"

	cosmoserrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	pellrelayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// UpdateNonce sets the Xmsg outbound nonce to the next nonce, and updates the nonce of blockchain state.
// It also updates the PendingNonces that is used to track the unfulfilled outbound txs.
func (k Keeper) UpdateNonce(ctx sdk.Context, receiveChainID int64, xmsg *types.Xmsg) error {
	chain := k.GetRelayerKeeper().GetSupportedChainFromChainID(ctx, receiveChainID)
	if chain == nil {
		return pellrelayertypes.ErrSupportedChains
	}

	nonce, found := k.GetRelayerKeeper().GetChainNonces(ctx, chain.ChainName())
	if !found {
		return cosmoserrors.Wrap(types.ErrCannotFindReceiverNonce, fmt.Sprintf("Chain(%s) | Identifiers : %s ", chain.ChainName(), xmsg.LogIdentifierForXmsg()))
	}

	// SET nonce
	xmsg.GetCurrentOutTxParam().OutboundTxTssNonce = nonce.Nonce
	tss, found := k.GetRelayerKeeper().GetTSS(ctx)
	if !found {
		return cosmoserrors.Wrap(types.ErrCannotFindTSSKeys, fmt.Sprintf("Chain(%s) | Identifiers : %s ", chain.ChainName(), xmsg.LogIdentifierForXmsg()))
	}

	p, found := k.GetRelayerKeeper().GetPendingNonces(ctx, tss.TssPubkey, receiveChainID)
	if !found {
		return cosmoserrors.Wrap(types.ErrCannotFindPendingNonces, fmt.Sprintf("chain_id %d, nonce %d", receiveChainID, nonce.Nonce))
	}

	// #nosec G701 always in range
	if p.NonceHigh != int64(nonce.Nonce) {
		return cosmoserrors.Wrap(types.ErrNonceMismatch, fmt.Sprintf("chain_id %d, high nonce %d, current nonce %d", receiveChainID, p.NonceHigh, nonce.Nonce))
	}

	ctx.Logger().Info("UpdateNonce before update", "chain", chain.ChainName(), "nonce", nonce.Nonce, "xmsg", xmsg.Index, "identifier", xmsg.LogIdentifierForXmsg())

	nonce.Nonce++
	p.NonceHigh++
	k.GetRelayerKeeper().SetChainNonces(ctx, nonce)
	k.GetRelayerKeeper().SetPendingNonces(ctx, p)

	k.Logger(ctx).Info("UpdateNonce after update", "chain", chain.ChainName(), "nonce", nonce.Nonce, "tss", tss.TssPubkey, "xmsgIndex", xmsg.Index)
	return nil
}

// GetRevertGasLimit returns the gas limit for the revert transaction in a Xmsg
// It returns 0 if there is no error but the gas limit can't be determined from the Xmsg data
func (k Keeper) GetRevertGasLimit(ctx sdk.Context, xmsg types.Xmsg) (uint64, error) {
	return 0, nil
}

func IsPending(xmsg *types.Xmsg) bool {
	// pending inbound is not considered a "pending" state because it has not reached consensus yet
	return xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_OUTBOUND || xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_REVERT
}

// GetAbortedAmount returns the amount to refund for a given Xmsg .
// If the Xmsg has an outbound transaction, it returns the amount of the outbound transaction.
// If OutTxParams is nil or the amount is zero, it returns the amount of the inbound transaction.
// This is because there might be a case where the transaction is set to be aborted before paying gas or creating an outbound transaction.In such a situation we can refund the entire amount that has been locked in connector or TSS
func GetAbortedAmount(xmsg types.Xmsg) sdkmath.Uint {
	return sdkmath.ZeroUint()
}

// SortXmsgsByHeightAndChainID sorts the cctxs by height (first come first serve), the chain ID doesn't really matter
func SortXmsgsByHeightAndChainID(cctxs []*types.Xmsg) []*types.Xmsg {
	sort.SliceStable(cctxs, func(i, j int) bool {
		if cctxs[i].InboundTxParams.InboundTxBlockHeight == cctxs[j].InboundTxParams.InboundTxBlockHeight {
			return cctxs[i].GetCurrentOutTxParam().ReceiverChainId < cctxs[j].GetCurrentOutTxParam().ReceiverChainId
		}
		return cctxs[i].InboundTxParams.InboundTxBlockHeight < cctxs[j].InboundTxParams.InboundTxBlockHeight
	})
	return cctxs
}
