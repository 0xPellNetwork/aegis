package keeper

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// processStoreAddSupportChainEvent processes the add supported chain event
func (h *RegistryRouterEventSubscriber) processStoreAddSupportChainEvent(ctx sdk.Context, event *registryrouter.RegistryRouterAddedSupportedChain) error {
	// Create DVSInfo from the event data
	dvsInfo := &types.DVSInfo{
		ChainId:          event.DvsInfo.ChainId.Uint64(),
		CentralScheduler: event.DvsInfo.CentralScheduler.Hex(),
		StakeManager:     event.DvsInfo.StakeManager.Hex(),
		EjectionManager:  event.DvsInfo.EjectionManager.Hex(),
		OutboundState:    types.OutboundStatus_OUTBOUND_STATUS_NORMAL,
	}
	h.Logger(ctx).Info("processStoreAddSupportChainEvent", "event", event, "dvsInfo", dvsInfo)
	return h.AddDVSSupportedChain(ctx, event.Raw.Address, dvsInfo)
}

// processStoreCreateGroupEvent processes the create quorum event
func (h *RegistryRouterEventSubscriber) processStoreCreateGroupEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncCreateGroup) error {
	poolParams := ConvertPoolParamsFromEventToStore(event.PoolParams)
	groupData := &types.Group{
		GroupNumber: uint64(event.GroupNumber),
		OperatorSetParam: &types.OperatorSetParam{
			MaxOperatorCount:        event.OperatorSetParams.MaxOperatorCount,
			KickBipsOfOperatorStake: uint32(event.OperatorSetParams.KickBIPsOfOperatorStake),
			KickBipsOfTotalStake:    uint32(event.OperatorSetParams.KickBIPsOfTotalStake),
		},
		MinimumStake: event.MinimumStake.Uint64(),
		PoolParams:   poolParams,
	}
	h.Logger(ctx).Info("processStoreCreateGroupEvent", "event", event, "groupData", groupData, "registryRouter", event.Raw.Address)
	return h.AddGroupData(ctx, event.Raw.Address, groupData)

}

// processStoreSetOperatorSetParamsEvent processes the set operator set params event
func (h *RegistryRouterEventSubscriber) processStoreSetOperatorSetParamsEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncSetOperatorSetParams) error {
	operatorSetParams := types.OperatorSetParam{
		MaxOperatorCount:        event.OperatorSetParams.MaxOperatorCount,
		KickBipsOfOperatorStake: uint32(event.OperatorSetParams.KickBIPsOfOperatorStake),
		KickBipsOfTotalStake:    uint32(event.OperatorSetParams.KickBIPsOfTotalStake),
	}

	return h.SetOperatorSetParams(ctx, event.Raw.Address, event.GroupNumber, operatorSetParams)
}

// processStoreSetGroupEjectionParamsEvent processes the set quorum ejection params event
func (h *RegistryRouterEventSubscriber) processStoreSetGroupEjectionParamsEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncSetGroupEjectionParams) error {
	quorumEjectionParams := types.GroupEjectionParam{
		RateLimitWindow:       event.GroupEjectionParams.RateLimitWindow,
		EjectableStakePercent: uint32(event.GroupEjectionParams.EjectableStakePercent),
	}

	return h.SetGroupEjectionParams(ctx, event.Raw.Address, event.GroupNumber, quorumEjectionParams)
}

// processStoreRegisterOperatorEvent processes the register operator event
func (h *RegistryRouterEventSubscriber) processStoreRegisterOperatorEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncRegisterOperator) error {
	registration := &types.GroupOperatorRegistrationV2{
		Operator:     event.Operator.Hex(),
		OperatorId:   event.OperatorId[:],
		GroupNumbers: event.GroupNumbers,
		Socket:       event.Socket,
		PubkeyParams: ConvertPubkeyRegistrationParamsFromEventToStore(event.Params),
	}

	h.Logger(ctx).Info("processStoreRegisterOperatorEvent", "event", event, "registryRouter", event.Raw.Address)
	return h.AddGroupOperatorRegistration(ctx, event.Raw.Address, registration)
}

// processStoreRegisterOperatorWithChurnEvent processes the register operator with churn event
func (h *RegistryRouterEventSubscriber) processStoreRegisterOperatorWithChurnEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncRegisterOperatorWithChurn) error {
	registration := &types.GroupOperatorRegistrationV2{
		Operator:     event.Operator.Hex(),
		OperatorId:   event.OperatorId[:],
		GroupNumbers: event.GroupNumbers,
		Socket:       event.Socket,
		PubkeyParams: ConvertPubkeyRegistrationParamsFromEventToStore(event.Params),
	}

	h.Logger(ctx).Info("processStoreRegisterOperatorWithChurnEvent", "event", event, "registryRouter", event.Raw.Address)
	return h.AddGroupOperatorRegistration(ctx, event.Raw.Address, registration)
}

// processStoreAddPoolsEvent processes the add strategies event
func (h *RegistryRouterEventSubscriber) processStoreAddPoolsEvent(ctx sdk.Context, event *stakeregistryrouter.StakeRegistryRouterSyncAddPools) error {
	params := ConvertPoolParamsFromStakeEventToStore(event.PoolParams)
	registryRouterAddr, err := h.GetStakeRegistryRouterAddress(ctx, event.Raw.Address)
	if err != nil {
		return err
	}

	h.Logger(ctx).Info("processStoreAddPoolsEvent", "registryRouterAddr", registryRouterAddr, "groupNumber", event.GroupNumber, "params", params)
	return h.AddPools(ctx, registryRouterAddr, event.GroupNumber, params)
}

// processStoreRemovePoolsEvent processes the remove strategies event
func (h *RegistryRouterEventSubscriber) processStoreRemovePoolsEvent(ctx sdk.Context, event *stakeregistryrouter.StakeRegistryRouterSyncRemovePools) error {
	registryRouterAddr, err := h.GetStakeRegistryRouterAddress(ctx, event.Raw.Address)
	if err != nil {
		return err
	}

	h.Logger(ctx).Info("processStoreRemovePoolsEvent", "registryRouterAddr", registryRouterAddr, "groupNumber", event.GroupNumber, "indicesToRemove", event.IndicesToRemove)
	return h.RemovePools(ctx, registryRouterAddr, event.GroupNumber, event.IndicesToRemove)
}

// processStoreModifyPoolParamsEvent processes the modify strategy params event
func (h *RegistryRouterEventSubscriber) processStoreModifyPoolParamsEvent(ctx sdk.Context, event *stakeregistryrouter.StakeRegistryRouterSyncModifyPoolParams) error {
	registryRouterAddr, err := h.GetStakeRegistryRouterAddress(ctx, event.Raw.Address)
	if err != nil {
		return err
	}

	h.Logger(ctx).Info("processStoreModifyPoolParamsEvent", "registryRouterAddr", registryRouterAddr, "groupNumber", event.GroupNumber, "poolIndices", event.PoolIndices, "newMultipliers", event.NewMultipliers)
	return h.ModifyPoolParams(ctx, registryRouterAddr, event.GroupNumber, event.PoolIndices, event.NewMultipliers)
}

// processStoreDeRegisterOperatorEvent processes the deregister operator event
func (h *RegistryRouterEventSubscriber) processStoreDeRegisterOperatorEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncDeregisterOperator) error {
	h.Logger(ctx).Info("processStoreDeRegisterOperatorEvent", "registryRouterAddr", event.Raw.Address, "operator", event.Operator)
	return h.RemoveGroupOperatorRegistration(ctx, event.Raw.Address, event.Operator)
}

// processStoreEjectOperatorsEvent processes the eject operators event
func (h *RegistryRouterEventSubscriber) processStoreEjectOperatorsEvent(ctx sdk.Context, event *registryrouter.RegistryRouterSyncEjectOperators) error {
	h.Logger(ctx).Info("processStoreEjectOperatorsEvent", "registryRouterAddr", event.Raw.Address, "operatorIds", event.OperatorIds)
	return h.RemoveGroupOperatorRegistrationByIds(ctx, event.Raw.Address, event.OperatorIds)
}
