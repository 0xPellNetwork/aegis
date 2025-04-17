package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// ProcessEpochs processes the epochs
func (k Keeper) ProcessEpochs(ctx sdk.Context) {
	blocksPerEpoch := k.GetBlocksPerEpoch(ctx)
	blockHeight := ctx.BlockHeight()

	if blockHeight%int64(blocksPerEpoch) != 0 {
		return
	}

	epochNumber := k.GetEpochNumber(ctx)
	defer k.SetEpochNumber(ctx, epochNumber+1)

	k.Logger(ctx).Info("restaking keeper process epochs",
		"blockHeight", blockHeight,
		"blocksPerEpoch", blocksPerEpoch,
		"currentEpochNumber", epochNumber,
	)

	chainsEpochOutboundState, err := k.GetAllOutboundStates(ctx)
	if err != nil {
		return
	}

	// sync shares change record
	for _, state := range chainsEpochOutboundState {
		k.Logger(ctx).Info("restaking keeper process epochs sync shares by epoch range",
			"stateEpochNumber", state.EpochNumber,
			"currentEpochNumber", epochNumber,
			"chainId", state.ChainId,
			"outboundState", state)

		var xmsgIndexes []string
		var err error

		chainParams, found := k.relayerKeeper.GetChainParamsByChainID(ctx, int64(state.ChainId))
		if !found {
			continue
		}

		switch state.OutboundStatus {
		case types.OutboundStatus_OUTBOUND_STATUS_INITIALIZING:
			if xmsgIndexes, err = k.syncAllShares(ctx, uint64(blockHeight), state, chainParams); err != nil {
				k.Logger(ctx).Error("sync all shares", "error", err)
				continue
			}
		case types.OutboundStatus_OUTBOUND_STATUS_NORMAL:
			if xmsgIndexes, err = k.syncSharesByEpochRange(ctx, uint64(blockHeight), chainParams, uint64(state.EpochNumber)+1, uint64(epochNumber), uint64(chainParams.ChainId)); err != nil {
				k.Logger(ctx).Error("sync shares by epoch range", "error", err)
				continue
			}
		default:
			continue
		}

		k.Logger(ctx).Info("sync shares by epoch range", "xmsgIndexes", xmsgIndexes)

		if len(xmsgIndexes) == 0 {
			state.EpochNumber = uint64(epochNumber)
			k.SetOutboundState(ctx, state)
			continue
		}

		// TODO: call hooks maybe
		k.SetEpochOperatorSharesSyncTxs(ctx, uint64(chainParams.ChainId), uint64(epochNumber), xmsgIndexes)
		state.OutboundStatus = types.OutboundStatus_OUTBOUND_STATUS_SYNCING
		k.SetOutboundState(ctx, state)
	}
}
