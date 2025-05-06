package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// processDVSEvent processes the DVS event
func (h RegistryRouterEventSubscriber) processDVSEvent(
	ctx sdk.Context,
	txOrigin string,
	registryRouterAddr ethcommon.Address,
	buildCalldataFn func(dvs *types.DVSInfo) ([]byte, string, error),
	raw *evmtypes.Log,
) (*sdkmath.Int, error) {
	supportedChain, exist := h.GetDVSSupportedChainListByStatus(ctx, registryRouterAddr, types.OutboundStatus_OUTBOUND_STATUS_NORMAL)
	if !exist {
		ctx.Logger().Error("Failed to get supported DVS", "registryRouterAddr", registryRouterAddr.Hex())
		return nil, nil
	}

	fee := sdkmath.ZeroInt()

	// Process each supported DVS
	for _, dvs := range supportedChain {
		message, receiverAddr, err := buildCalldataFn(dvs)
		if err != nil {
			ctx.Logger().Warn("Failed to encode message", "error", err)
			continue
		}

		receiverChainID := big.NewInt(0).SetUint64(dvs.ChainId)
		xmsg, err := h.processInboundEvent(ctx, txOrigin, receiverChainID, registryRouterAddr, ethcommon.HexToAddress(receiverAddr), message, raw)
		if err != nil {
			ctx.Logger().Error("Failed to sent xmsg from event", "error", err)
			continue
		}

		if feeParam, exist := h.xmsgKeeper.GetCrosschainEventFee(ctx, int64(dvs.ChainId)); exist && feeParam.IsSupported {
			fee = fee.Add(feeParam.DvsDefaultFee)
		}

		ctx.Logger().Info("Sent xmsg from event", "xmsg", xmsg)
	}
	return &fee, nil
}

// processSyncCreateGroupEvent processes the create group event
func (h RegistryRouterEventSubscriber) processSyncCreateGroupEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncCreateGroup) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncCreateGroup(event.GroupNumber, event.OperatorSetParams, event.MinimumStake, event.PoolParams)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncCreateGroupEvent", "operatorSetParams", event.OperatorSetParams, "minimumStake", event.MinimumStake, "poolParams", event.PoolParams, "registryRouter", event.Raw.Address)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncSetOperatorSetParamsEvent processes the set operator set params event
func (h RegistryRouterEventSubscriber) processSyncSetOperatorSetParamsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncSetOperatorSetParams) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncSetOperatorSetParams(event.GroupNumber, event.OperatorSetParams)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncSetOperatorSetParamsEvent", "groupNumber", event.GroupNumber, "operatorSetParams", event.OperatorSetParams)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncSetGroupEjectionParamsEvent processes the set group ejection params event
func (h RegistryRouterEventSubscriber) processSyncSetGroupEjectionParamsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncSetGroupEjectionParams) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncSetGroupEjectionParams(event.GroupNumber, event.GroupEjectionParams)
		return message, dvs.EjectionManager, err
	}
	ctx.Logger().Info("processSyncSetGroupEjectionParamsEvent", "groupNumber", event.GroupNumber, "ejectionParams", event.GroupEjectionParams)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncEjectionCooldownEvent processes the ejection cooldown event
func (h RegistryRouterEventSubscriber) processSyncEjectionCooldownEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncEjectionCooldown) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncEjectionCooldown(event.EjectionCooldown)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncEjectionCooldownEvent", "cooldown", event.EjectionCooldown)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncRegisterOperatorEvent processes the register operator event
func (h RegistryRouterEventSubscriber) processSyncRegisterOperatorEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncRegisterOperator) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncRegisterOperator(event.Operator, event.GroupNumbers, event.Params)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncRegisterOperatorEvent", "operator", event.Operator.Hex(), "groupNumbers", event.GroupNumbers)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncRegisterOperatorWithChurnEvent processes the register operator with churn event
func (h RegistryRouterEventSubscriber) processSyncRegisterOperatorWithChurnEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncRegisterOperatorWithChurn) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncRegisterOperatorWithChurn(event.Operator, event.GroupNumbers, event.Params, event.OperatorKickParams)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncRegisterOperatorWithChurnEvent", "operator", event.Operator.Hex(), "groupNumbers", event.GroupNumbers)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncAddPoolsEvent processes the add strategies event
func (h RegistryRouterEventSubscriber) processSyncAddPoolsEvent(ctx sdk.Context, txOrigin string, event *stakeregistryrouter.StakeRegistryRouterSyncAddPools) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncAddPools(event.GroupNumber, event.PoolParams)
		return message, dvs.StakeManager, err
	}
	registryRouterAddr, err := h.GetStakeRegistryRouterAddress(ctx, event.Raw.Address)
	if err != nil {
		return nil, err
	}

	ctx.Logger().Info("processSyncAddPoolsEvent", "groupNumber", event.GroupNumber, "poolParams", event.PoolParams, "registryRouter", registryRouterAddr)
	return h.processDVSEvent(ctx, txOrigin, registryRouterAddr, encodeMessageFunc, &event.Raw)
}

// processSyncDeRegisterOperatorEvent processes the deregister operator event
func (h RegistryRouterEventSubscriber) processSyncRemovePoolsEvent(ctx sdk.Context, txOrigin string, event *stakeregistryrouter.StakeRegistryRouterSyncRemovePools) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncRemovePools(event.GroupNumber, event.IndicesToRemove)
		return message, dvs.StakeManager, err
	}
	registryRouterAddr, err := h.GetStakeRegistryRouterAddress(ctx, event.Raw.Address)
	if err != nil {
		return nil, err
	}

	ctx.Logger().Info("processSyncRemovePoolsEvent", "groupNumber", event.GroupNumber, "strategy indices", event.IndicesToRemove, "registryRouter", registryRouterAddr)
	return h.processDVSEvent(ctx, txOrigin, registryRouterAddr, encodeMessageFunc, &event.Raw)
}

// processSyncModifyPoolParamsEvent processes the modify strategy params event
func (h RegistryRouterEventSubscriber) processSyncModifyPoolParamsEvent(ctx sdk.Context, txOrigin string, event *stakeregistryrouter.StakeRegistryRouterSyncModifyPoolParams) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncModifyPoolParams(event.GroupNumber, event.PoolIndices, event.NewMultipliers)
		return message, dvs.StakeManager, err
	}
	registryRouterAddr, err := h.GetStakeRegistryRouterAddress(ctx, event.Raw.Address)
	if err != nil {
		return nil, err
	}

	ctx.Logger().Info("processSyncModifyPoolParamsEvent", "groupNumber", event.GroupNumber, "poolIndices", event.PoolIndices, "newMultipliers", event.NewMultipliers, "registryRouter", registryRouterAddr)
	return h.processDVSEvent(ctx, txOrigin, registryRouterAddr, encodeMessageFunc, &event.Raw)
}

// processSyncDeRegisterOperatorEvent processes the deregister operator event
func (h RegistryRouterEventSubscriber) processSyncDeRegisterOperatorEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncDeregisterOperator) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncDeregisterOperator(event.Operator, event.GroupNumbers)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncDeRegisterOperatorEvent", "operator", event.Operator.Hex(), "groupNumbers", event.GroupNumbers)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncUpdateOperatorsEvent processes the update operators event
func (h RegistryRouterEventSubscriber) processSyncUpdateOperatorsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncUpdateOperators) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncUpdateOperators(event.Operators)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncUpdateOperatorsEvent", "operators", event.Operators)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncUpdateOperatorsForGroupEvent processes the update operators for group event
func (h RegistryRouterEventSubscriber) processSyncUpdateOperatorsForGroupEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncUpdateOperatorsForGroup) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncUpdateOperatorsForGroup(event.OperatorsPerGroup, event.GroupNumbers)
		return message, dvs.CentralScheduler, err
	}
	ctx.Logger().Info("processSyncUpdateOperatorsForGroupEvent", "groupNumber", event.GroupNumbers, "operators", event.OperatorsPerGroup)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}

// processSyncEjectOperatorsEvent processes the eject operators event
func (h RegistryRouterEventSubscriber) processSyncEjectOperatorsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncEjectOperators) (*sdkmath.Int, error) {
	encodeMessageFunc := func(dvs *types.DVSInfo) ([]byte, string, error) {
		message, err := encodeSyncEjectOperators(event.OperatorIds)
		return message, dvs.EjectionManager, err
	}
	ctx.Logger().Info("processSyncEjectOperatorsEvent", "operators", event.OperatorIds)
	return h.processDVSEvent(ctx, txOrigin, event.Raw.Address, encodeMessageFunc, &event.Raw)
}
