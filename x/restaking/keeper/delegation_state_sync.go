package keeper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/utils"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const BATCH_SYNC_DELEGATION_SHARES_FN_NAME = "batchSyncDelegatedShares"

const BATCH_SIZE = 50

// syncSharesByEpochRange syncs operator share changes that occurred between the specified epoch range to the target chain.
// If any shares were modified during these epochs, they will be batched (in groups of BATCH_SIZE) and sent
// to the target chain via cross-chain messages(xmsg). This ensures the operator shares stay synchronized across chains.
func (k Keeper) syncSharesByEpochRange(ctx sdk.Context, blockNum uint64, chainParams *relayertypes.ChainParams, startEpoch, endEpoch uint64, chainID uint64) (xmsgIndexes []string, err error) {
	sharesChange := k.GetChangedOperatorSharesSnapshotByEpochRange(ctx, startEpoch, endEpoch)
	if sharesChange == nil || len(sharesChange.OperatorShares) == 0 {
		return nil, nil
	}

	k.Logger(ctx).Info("syncSharesByEpochRange", "startEpoch", startEpoch, "endEpoch", endEpoch, "chainID", chainID, "sharesChange", sharesChange)

	syncIndex := 0

	for i := 0; i < len(sharesChange.OperatorShares); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(sharesChange.OperatorShares) {
			end = len(sharesChange.OperatorShares)
		}

		batch := sharesChange.OperatorShares[i:end]

		event, err := buildOperatorSharesPellSent(chainParams, batch)
		if err != nil {
			k.Logger(ctx).Error("syncSharesByEpochRange: failed to build event", "error", err)
			return nil, err
		}

		xmsgIndex, err := k.processSyncEvent(ctx, chainID, endEpoch, blockNum, syncIndex, chainParams, event)
		if err != nil {
			k.Logger(ctx).Error("syncSharesByEpochRange: failed to process sync event", "error", err)
			return nil, err
		}

		k.Logger(ctx).Debug("syncSharesByEpochRange: processed sync event", "xmsgIndex", xmsgIndex)

		syncIndex++
		xmsgIndexes = append(xmsgIndexes, xmsgIndex)
	}

	return xmsgIndexes, nil
}

// Synchronize all operator shares information
func (k Keeper) syncAllShares(ctx sdk.Context, blockNum uint64, outboundState *types.EpochOutboundState, chainParams *relayertypes.ChainParams) (xmsgIndexes []string, err error) {
	operatorShares := k.GetAllShares(ctx)
	if len(operatorShares) == 0 {
		return nil, nil
	}

	endEpoch := k.GetEpochNumber(ctx)
	syncIndex := 0

	for i := 0; i < len(operatorShares); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(operatorShares) {
			end = len(operatorShares)
		}

		batch := operatorShares[i:end]

		event, err := buildOperatorSharesPellSent(chainParams, batch)
		if err != nil {
			return nil, err
		}

		xmsgIndex, err := k.processSyncEvent(ctx, endEpoch-1, blockNum, outboundState.ChainId, syncIndex, chainParams, event)
		if err != nil {
			return nil, err
		}

		syncIndex++
		xmsgIndexes = append(xmsgIndexes, xmsgIndex)
	}

	return xmsgIndexes, nil
}

// build operator shares pell sent
func buildOperatorSharesPellSent(chainParams *relayertypes.ChainParams, operatorShares []*types.OperatorShares) (*xmsgtypes.InboundPellEvent, error) {
	chainIds := make([]*big.Int, len(operatorShares))
	operators := make([]ethcommon.Address, len(operatorShares))
	strategies := make([]ethcommon.Address, len(operatorShares))
	shares := make([]*big.Int, len(operatorShares))

	for i, share := range operatorShares {
		chainIds[i] = new(big.Int).SetUint64(share.ChainId)
		operators[i] = ethcommon.HexToAddress(share.Operator)
		strategies[i] = ethcommon.HexToAddress(share.Strategy)
		shares[i] = share.Shares.BigInt()
	}

	// pack data
	data, err := omniOperatorSharesManagerMetaDataABI.Pack(BATCH_SYNC_DELEGATION_SHARES_FN_NAME,
		chainIds, operators, strategies, shares)
	if err != nil {
		return nil, fmt.Errorf("failed to pack data: %w", err)
	}

	return &xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_PellSent{
			PellSent: &xmsgtypes.PellSent{
				TxOrigin:        types.ModuleAddressEVM.Hex(),
				Sender:          types.ModuleAddressEVM.Hex(),
				ReceiverChainId: chainParams.ChainId,
				Receiver:        chainParams.OmniOperatorSharesManagerContractAddress,
				Message:         base64.StdEncoding.EncodeToString(data),
				PellParams:      pevmtypes.ReceiveCall.String(),
			},
		},
	}, nil
}

// process sync event. build xmsg and process
func (k Keeper) processSyncEvent(ctx sdk.Context, chainID, epochNumber, blockNum uint64, batchIndex int, chainParams *relayertypes.ChainParams, event *xmsgtypes.InboundPellEvent) (xmsgIndex string, err error) {
	systemTxId := utils.GenerateSystemTxId(pevmtypes.SystemTxTypeSyncDelegationShares, epochNumber, uint8(batchIndex))

	receiverChain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, int64(chainID))
	if receiverChain == nil {
		return "", errorsmod.Wrapf(relayertypes.ErrSupportedChains, "chain with chainID %d not supported", chainID)
	}

	xmsg, err := k.buildXmsg(ctx, systemTxId, blockNum, batchIndex, chainParams, event)
	if err != nil {
		return "", err
	}

	xmsg.SetPendingOutbound("PellConnector pell-send event setting to pending outbound directly")

	// Get gas price and amount
	gasprice, found := k.xmsgKeeper.GetGasPrice(ctx, int64(chainID))
	if !found {
		return "", fmt.Errorf("gasprice not found for %d", chainID)
	}

	xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = fmt.Sprint(gasprice.Prices[gasprice.MedianIndex])

	return xmsg.Index, k.xmsgKeeper.ProcessXmsg(ctx, *xmsg, receiverChain)
}

func (k Keeper) buildXmsg(ctx sdk.Context, systemTxId string, blockNum uint64, eventIndex int, chainParams *relayertypes.ChainParams, event *xmsgtypes.InboundPellEvent) (*xmsgtypes.Xmsg, error) {
	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, errors.New("cannot process logs without TSS keys")
	}

	senderChain, err := chains.PellChainFromChainID(ctx.ChainID())
	if err != nil {
		return nil, fmt.Errorf("syncSharesByType: failed to convert chainID: %s", err.Error())
	}

	voteMsg := xmsgtypes.NewMsgVoteOnObservedInboundTx(
		"",
		types.ModuleAddressEVM.Hex(),
		senderChain.Id,
		types.ModuleAddressEVM.Hex(),
		chainParams.OmniOperatorSharesManagerContractAddress,
		chainParams.ChainId,
		systemTxId, // system tx hash. nil
		blockNum,
		chainParams.GasLimit, // TODO: gas limit maybe need to be changed.
		uint(eventIndex),
		*event,
	)

	xmsg, err := xmsgtypes.NewXmsg(ctx, *voteMsg, tss.TssPubkey)
	if err != nil {
		return nil, fmt.Errorf("failed to create xmsg: %w", err)
	}

	return &xmsg, nil
}
