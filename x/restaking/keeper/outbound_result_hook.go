package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/pkg/utils"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var _ xmsgtypes.XmsgOutboundResultHook = Hooks{}

// ProcessXmsgOutboundResult processes the outbound result of a xmsg
// The chain can only proceed with synchronization when all transactions within an epoch have been successfully synchronized.
// This mechanism ensures data consistency and completeness across the chain.
func (h Hooks) ProcessXmsgOutboundResult(ctx sdk.Context, xmsg *xmsgtypes.Xmsg, ballotStatus relayertypes.BallotStatus) {
	h.processEpochSync(ctx, xmsg, ballotStatus)
}

// processEpochSync processes the epoch sync
func (h Hooks) processEpochSync(ctx sdk.Context, xmsg *xmsgtypes.Xmsg, ballotStatus relayertypes.BallotStatus) {
	systemTxType, epoch, _, err := utils.ParseSystemTxId(xmsg.InboundTxParams.InboundTxHash)
	if err != nil || systemTxType != pevmtypes.SystemTxTypeSyncDelegationShares {
		return
	}

	h.Logger(ctx).Info("ProcessXmsgOutboundResult", "xmsg", xmsg, "ballotStatus", ballotStatus)

	syncEpochTxs, exist := h.GetEpochOperatorSharesSyncTxs(ctx, uint64(xmsg.GetCurrentOutTxParam().ReceiverChainId), epoch)
	if !exist {
		return
	}

	if len(syncEpochTxs.PendingXmsgIndexes) == 0 {
		return
	}

	h.Logger(ctx).Info("ProcessXmsgOutboundRes", "pendingTxIndexes", syncEpochTxs.PendingXmsgIndexes)

	outboundState, found := h.GetOutboundState(ctx, uint64(xmsg.GetCurrentOutTxParam().ReceiverChainId))
	if !found || outboundState == nil {
		return
	}

	if ballotStatus == relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION {
		for i, txIndex := range syncEpochTxs.PendingXmsgIndexes {
			if txIndex == xmsg.Index {
				syncEpochTxs.PendingXmsgIndexes = append(syncEpochTxs.PendingXmsgIndexes[:i], syncEpochTxs.PendingXmsgIndexes[i+1:]...)
				break
			}
		}

		h.Logger(ctx).Info("ProcessXmsgOutboundRes", "syncEpochTxs", syncEpochTxs)

		// sync epoch delegation outbound txs all done
		if len(syncEpochTxs.PendingXmsgIndexes) == 0 {
			h.DeleteEpochOperatorSharesSyncTxs(ctx, uint64(xmsg.GetCurrentOutTxParam().ReceiverChainId), epoch)

			outboundState.OutboundStatus = types.OutboundStatus_OUTBOUND_STATUS_NORMAL
			outboundState.EpochNumber = epoch

			h.SetOutboundState(ctx, outboundState)
		} else {
			h.SetEpochOperatorSharesSyncTxs(ctx, uint64(xmsg.GetCurrentOutTxParam().ReceiverChainId), epoch, syncEpochTxs.PendingXmsgIndexes)
		}
	}
}
