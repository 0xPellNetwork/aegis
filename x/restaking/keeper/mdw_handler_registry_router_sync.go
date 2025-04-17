package keeper

import (
	"errors"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

type MiddlewareSyncHandler struct {
	Keeper
}

func NewMiddlewareHistoryEventHandler(keeper Keeper) *MiddlewareSyncHandler {
	return &MiddlewareSyncHandler{
		Keeper: keeper,
	}
}

// SyncHistoryEvent syncs the history event for a given registry router address and chain id
func (h *MiddlewareSyncHandler) SyncHistoryEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncGroup) (*sdkmath.Int, error) {
	registryRouterAddr := event.Raw.Address
	txHash := event.Raw.TxHash.String()
	chainID := event.DestinationChainId.Uint64()
	groupNumbers := event.GroupNumbers
	ctx.Logger().Info("Syncing history event", "registryRouterAddr", registryRouterAddr, "txHash", txHash, "chainID", chainID, "groupNumbers", groupNumbers)
	res := sdkmath.ZeroInt()

	dvs, exist := h.GetDVSSupportedChain(ctx, registryRouterAddr, chainID)
	if !exist {
		ctx.Logger().Error("Failed to get supported DVS", "registryRouterAddr", registryRouterAddr)
		return nil, errors.New("supported DVS not found")
	}

	groupSyncFee, err := h.handleGroupSync(ctx, dvs, registryRouterAddr, txHash, groupNumbers)
	if err != nil {
		ctx.Logger().Error("Failed to handle group sync", "error", err)
		return nil, err
	}

	res = res.Add(*groupSyncFee)

	operatorRegistrationFee, err := h.handleOperatorRegistrationSync(ctx, dvs, registryRouterAddr, txHash, groupNumbers)
	if err != nil {
		ctx.Logger().Error("Failed to handle operator registration sync", "error", err)
		return nil, err
	}

	res = res.Add(*operatorRegistrationFee)

	return &res, nil
}

// handleGroupSync handles the group data sync
func (h *MiddlewareSyncHandler) handleGroupSync(ctx sdk.Context, dvs *types.DVSInfo, registryRouterAddr ethcommon.Address, txHash string, groupNumbers []byte) (*sdkmath.Int, error) {
	fee := sdkmath.ZeroInt()
	// Process each group in the list
	for _, groupNumber := range groupNumbers {
		group, err := h.Keeper.GetGroupByGroupNumber(ctx, registryRouterAddr, uint64(groupNumber))
		if err != nil {
			ctx.Logger().Error("Failed to get group by group number in handle group sync", "error", err)
			return nil, err
		}

		ctx.Logger().Info("Processing group", "registryRouterAddr", registryRouterAddr.Hex(), "groupNumber", group.GroupNumber, "poolParams", group.PoolParams)

		if err := h.processGroup(ctx, group, registryRouterAddr, txHash, dvs); err != nil {
			ctx.Logger().Error("Failed to process group", "error", err)
			return nil, err
		}

		feeParam, exist := h.xmsgKeeper.GetCrosschainEventFee(ctx, int64(dvs.ChainId))
		if exist {
			fee = fee.Add(feeParam.RegistryRouterSyncGroupEventFee)
		}

		if len(group.PoolParams) <= types.MaxPoolParamsPerMsg {
			ctx.Logger().Info("No pools to process", "registryRouterAddr", registryRouterAddr.Hex(), "groupNumber", group.GroupNumber, "poolParams", group.PoolParams)
			continue
		}

		processPoolsFee, err := h.processPools(ctx, group, registryRouterAddr, txHash, dvs)
		if err != nil {
			ctx.Logger().Error("Failed to process pools", "error", err)
			return nil, err
		}

		fee = fee.Add(*processPoolsFee)
	}

	return &fee, nil
}

// processGroup processes each group and generates events
func (h *MiddlewareSyncHandler) processGroup(ctx sdk.Context, group *types.Group, registryRouterAddr ethcommon.Address, txHash string, dvs *types.DVSInfo) error {
	var poolParamsStore []registryrouter.IStakeRegistryRouterPoolParams

	// get first 2 pool params and build the message
	endIndex := len(group.PoolParams)
	if endIndex > types.MaxPoolParamsPerMsg {
		endIndex = types.MaxPoolParamsPerMsg
	}

	poolParamsStore, err := ConvertPoolParamsFromStoreToEvent(group.PoolParams[:endIndex])
	if err != nil {
		return err
	}

	message, err := encodeSyncCreateGroup(
		uint8(group.GroupNumber),
		ConvertOperatorSetParamFromStoreToEvent(group.OperatorSetParam),
		new(big.Int).SetUint64(group.MinimumStake),
		poolParamsStore,
	)
	if err != nil {
		return err
	}

	receiverChainID := new(big.Int).SetUint64(dvs.ChainId)
	receiverAddr := ethcommon.HexToAddress(dvs.CentralScheduler)

	xmsg, err := h.processInboundEvent(ctx, types.EmptyTxOrigin, receiverChainID, registryRouterAddr, receiverAddr, message, &ethtypes.Log{TxHash: ethcommon.HexToHash(txHash)})
	if err != nil {
		return err
	}

	ctx.Logger().Info("Group sync progress added in process group",
		"receiver", receiverAddr.Hex(),
		"poolParams", poolParamsStore,
		"txHash", txHash,
		"xmsgHash", xmsg.Index,
		"groupNumber", group.GroupNumber,
		"operatorSetParam", group.OperatorSetParam,
		"minStake", group.MinimumStake,
	)
	return h.AddGroupSync(ctx, txHash, xmsg.Index)
}

// processQuorums processes each quorum and generates events
func (h *MiddlewareSyncHandler) processPools(ctx sdk.Context, group *types.Group, registryRouterAddr ethcommon.Address, txHash string, dvs *types.DVSInfo) (*sdkmath.Int, error) {
	if len(group.PoolParams) <= types.MaxPoolParamsPerMsg {
		h.Logger(ctx).Info("No pools to process", "registryRouterAddr", registryRouterAddr.Hex())
		return nil, nil
	}

	poolParamsStore := group.PoolParams[types.MaxPoolParamsPerMsg:]

	poolParams, err := ConvertPoolParamsFromStoreToStakeEvent(poolParamsStore)
	if err != nil {
		return nil, err
	}

	xmsgIndices := make([]string, 0)

	fee := sdkmath.ZeroInt()
	// Split poolParams into chunks
	for i := 0; i < len(poolParams); i += types.MaxPoolParamsPerMsg {
		end := i + types.MaxPoolParamsPerMsg
		if end > len(poolParams) {
			end = len(poolParams)
		}

		chunkParams := poolParams[i:end]
		callMsg, err := encodeSyncAddPools(byte(group.GroupNumber), chunkParams)
		if err != nil {
			return nil, err
		}

		receiverChainID := new(big.Int).SetUint64(dvs.ChainId)
		receiverAddr := ethcommon.HexToAddress(dvs.StakeManager)
		xmsg, err := h.processInboundEvent(ctx, types.EmptyTxOrigin, receiverChainID, registryRouterAddr, receiverAddr, callMsg, &ethtypes.Log{TxHash: ethcommon.HexToHash(txHash)})
		if err != nil {
			return nil, err
		}

		if feeParam, exist := h.xmsgKeeper.GetCrosschainEventFee(ctx, int64(dvs.ChainId)); exist {
			fee = fee.Add(feeParam.RegistryRouterSyncGroupEventFee)
		}

		h.Logger(ctx).Info("Pools sync progress added", "receiver", receiverAddr.Hex(), "strategyParams", chunkParams, "txHash", txHash, "xmsgHash", xmsg.Index)
		xmsgIndices = append(xmsgIndices, xmsg.Index)
	}

	return &fee, h.AddGroupSyncs(ctx, txHash, xmsgIndices)
}

// handleOperatorRegistrationSync handles the operator registration sync
func (h *MiddlewareSyncHandler) handleOperatorRegistrationSync(ctx sdk.Context, dvs *types.DVSInfo, registryRouterAddr ethcommon.Address, txHash string, groupNumbers []byte) (*sdkmath.Int, error) {
	fee := sdkmath.ZeroInt()

	list, found := h.Keeper.GetGroupOperatorRegistrationList(ctx, registryRouterAddr)
	if !found {
		ctx.Logger().Info("Failed to get registration operator registration list", "registryRouterAddr", registryRouterAddr.Hex())
		return nil, nil
	}

	// Process each registration in the list
	for _, registration := range list.OperatorRegisteredInfos {

		// Find the intersection of registration.GroupNumbers and groupNumbers, i.e., elements that exist in both slices
		needSyncGroupNumbers := IntersectBytes(registration.GroupNumbers, groupNumbers)

		message, err := encodeSyncRegisterOperator(
			ethcommon.HexToAddress(registration.Operator),
			needSyncGroupNumbers,
			ConvertPubkeyRegistrationParamsFromStoreToEvent(registration.PubkeyParams),
		)
		if err != nil {
			continue
		}

		receiverChainID := big.NewInt(0).SetUint64(dvs.ChainId)
		receiverAddr := ethcommon.HexToAddress(dvs.CentralScheduler)

		xmsg, err := h.processInboundEvent(ctx, types.EmptyTxOrigin, receiverChainID, registryRouterAddr, receiverAddr, message, &ethtypes.Log{TxHash: ethcommon.HexToHash(txHash)})
		if err != nil {
			ctx.Logger().Error("Failed to build xmsg from event", "error", err)
			continue
		}

		if feeParam, exist := h.xmsgKeeper.GetCrosschainEventFee(ctx, int64(dvs.ChainId)); exist {
			fee = fee.Add(feeParam.DelegationOperatorSyncFee)
		}

		if err := h.AddGroupSync(ctx, txHash, xmsg.Index); err != nil {
			ctx.Logger().Error("Failed to add registration sync progress", "error", err)
			return nil, err
		}

		ctx.Logger().Info("Group sync progress added in registration sync", "sender", registryRouterAddr.Hex(), "txHash", txHash, "xmsgHash", xmsg.Index)
	}

	return &fee, nil
}
