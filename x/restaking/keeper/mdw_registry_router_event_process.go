package keeper

import (
	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (h *RegistryRouterEventSubscriber) processSyncGroupEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncGroup) (*sdkmath.Int, error) {
	return h.Keeper.middlewareSyncHandler.SyncHistoryEvent(ctx, event)
}

// processAddSupportChainEvent processes the add support chain event
func (h *RegistryRouterEventSubscriber) processAddSupportChainEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterAddedSupportedChain) (*sdkmath.Int, error) {
	if err := h.processStoreAddSupportChainEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process add support chain event by pevm", "error", err)
		return nil, err
	}

	// don't need to process sync
	return nil, nil
}

// processCreateGroupEvent processes the create group event
func (h *RegistryRouterEventSubscriber) processCreateGroupEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncCreateGroup) (*sdkmath.Int, error) {
	if err := h.processStoreCreateGroupEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process create group event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncCreateGroupEvent(ctx, txOrigin, event)
}

// processRegisterOperatorEvent processes the register operator event
func (h *RegistryRouterEventSubscriber) processRegisterOperatorEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncRegisterOperator) (*sdkmath.Int, error) {
	if err := h.processStoreRegisterOperatorEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process register operator event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncRegisterOperatorEvent(ctx, txOrigin, event)
}

// processRegisterOperatorWithChurnEvent processes the register operator with churn event
func (h *RegistryRouterEventSubscriber) processRegisterOperatorWithChurnEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncRegisterOperatorWithChurn) (*sdkmath.Int, error) {
	if err := h.processStoreRegisterOperatorWithChurnEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process register operator with churn event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncRegisterOperatorWithChurnEvent(ctx, txOrigin, event)
}

// processDeregisterOperatorEvent processes the deregister operator event
func (h *RegistryRouterEventSubscriber) processDeregisterOperatorEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncDeregisterOperator) (*sdkmath.Int, error) {
	if err := h.processStoreDeRegisterOperatorEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process deregister operator event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncDeRegisterOperatorEvent(ctx, txOrigin, event)
}

// processUpdateOperatorsEvent processes the update operators event
func (h *RegistryRouterEventSubscriber) processUpdateOperatorsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncUpdateOperators) (*sdkmath.Int, error) {
	// don't need to process store
	return h.processSyncUpdateOperatorsEvent(ctx, txOrigin, event)
}

// processUpdateOperatorsForGroupEvent processes the update operators for group event
func (h *RegistryRouterEventSubscriber) processUpdateOperatorsForGroupEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncUpdateOperatorsForGroup) (*sdkmath.Int, error) {
	// don't need to process store
	return h.processSyncUpdateOperatorsForGroupEvent(ctx, txOrigin, event)
}

// processOperatorSocketUpdateEvent processes the operator socket update event
// TODO: don't need it now
func (h *RegistryRouterEventSubscriber) processOperatorSocketUpdateEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterOperatorSocketUpdate) (*sdkmath.Int, error) {
	return nil, nil
}

// processEjectOperatorsEvent processes the eject operators event
func (h *RegistryRouterEventSubscriber) processEjectOperatorsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncEjectOperators) (*sdkmath.Int, error) {
	if err := h.processStoreEjectOperatorsEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process eject operators event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncEjectOperatorsEvent(ctx, txOrigin, event)
}

// processChurnApproverUpdatedEvent processes the churn approver updated event
func (h *RegistryRouterEventSubscriber) processChurnApproverUpdatedEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterChurnApproverUpdated) (*sdkmath.Int, error) {
	return nil, nil
}

// processEjectorUpdatedEvent processes the ejector updated event
// TODO: don't need it now
func (h *RegistryRouterEventSubscriber) processEjectorUpdatedEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterEjectorUpdated) (*sdkmath.Int, error) {
	return nil, nil
}

// processSetOperatorSetParamsEvent processes the set operator set params event
func (h *RegistryRouterEventSubscriber) processSetOperatorSetParamsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncSetOperatorSetParams) (*sdkmath.Int, error) {
	if err := h.processStoreSetOperatorSetParamsEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process set operator set params event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncSetOperatorSetParamsEvent(ctx, txOrigin, event)
}

// processSetGroupEjectionParamsEvent processes the set group ejection params event
func (h *RegistryRouterEventSubscriber) processSetGroupEjectionParamsEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncSetGroupEjectionParams) (*sdkmath.Int, error) {
	if err := h.processStoreSetGroupEjectionParamsEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process set group ejection params event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncSetGroupEjectionParamsEvent(ctx, txOrigin, event)
}

// processEjectionCooldownEvent processes the ejection cooldown event
func (h *RegistryRouterEventSubscriber) processEjectionCooldownEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterSyncEjectionCooldown) (*sdkmath.Int, error) {
	// don't need to process store

	return h.processSyncEjectionCooldownEvent(ctx, txOrigin, event)
}

// processMinimumStakeForGroupUpdatedEvent processes the minimum stake for group updated event
// TODO: don't need it now
func (h *RegistryRouterEventSubscriber) processMinimumStakeForGroupUpdatedEvent(ctx sdk.Context, txOrigin string, event *registryrouter.RegistryRouterMinimumStakeForGroupUpdated) (*sdkmath.Int, error) {
	return nil, nil
}

// processAddPoolsEvent processes the add pools event
func (h *RegistryRouterEventSubscriber) processAddPoolsEvent(ctx sdk.Context, txOrigin string, event *stakeregistryrouter.StakeRegistryRouterSyncAddPools) (*sdkmath.Int, error) {
	if err := h.processStoreAddPoolsEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process add pools event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncAddPoolsEvent(ctx, txOrigin, event)
}

// processRemovePoolsEvent processes the remove pools event
func (h *RegistryRouterEventSubscriber) processRemovePoolsEvent(ctx sdk.Context, txOrigin string, event *stakeregistryrouter.StakeRegistryRouterSyncRemovePools) (*sdkmath.Int, error) {
	if err := h.processStoreRemovePoolsEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process add pools event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncRemovePoolsEvent(ctx, txOrigin, event)
}

// processModifyPoolParamsEvent processes the modify pool params event
func (h *RegistryRouterEventSubscriber) processModifyPoolParamsEvent(ctx sdk.Context, txOrigin string, event *stakeregistryrouter.StakeRegistryRouterSyncModifyPoolParams) (*sdkmath.Int, error) {
	if err := h.processStoreModifyPoolParamsEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process modify pool params event by pevm", "error", err)
		return nil, err
	}

	return h.processSyncModifyPoolParamsEvent(ctx, txOrigin, event)
}
