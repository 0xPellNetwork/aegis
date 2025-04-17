package keeper

import (
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type RegistryRouterEventSubscriber struct {
	Keeper
}

func NewRegistryRouterEventSubscriber(keeper Keeper) *RegistryRouterEventSubscriber {
	return &RegistryRouterEventSubscriber{Keeper: keeper}
}

// ProcessLogs processes the logs
func (h RegistryRouterEventSubscriber) ProcessLogs(ctx sdk.Context, _ uint64, toAddress ethcommon.Address, log *ethtypes.Log, txOrigin string) (*sdkmath.Int, error) {
	// logs from registry router
	filter, err := registryrouter.NewRegistryRouterFilterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	// log function parameters
	ctx.Logger().Info("Processing logs", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)

	// log filter
	if event, err := filter.ParseSyncGroup(*log); err == nil {
		ctx.Logger().Info("Processing SyncGroup event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processSyncGroupEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseAddedSupportedChain(*log); err == nil {
		ctx.Logger().Info("Processing AddedSupportedChain event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processAddSupportChainEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncCreateGroup(*log); err == nil {
		ctx.Logger().Info("Processing SyncCreateGroup event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processCreateGroupEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncRegisterOperator(*log); err == nil {
		ctx.Logger().Info("Processing SyncRegisterOperator event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processRegisterOperatorEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncRegisterOperatorWithChurn(*log); err == nil {
		ctx.Logger().Info("Processing SyncRegisterOperatorWithChurn event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processRegisterOperatorWithChurnEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncDeregisterOperator(*log); err == nil {
		ctx.Logger().Info("Processing SyncDeregisterOperator event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processDeregisterOperatorEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncUpdateOperators(*log); err == nil {
		ctx.Logger().Info("Processing SyncUpdateOperators event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processUpdateOperatorsEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncUpdateOperatorsForGroup(*log); err == nil {
		ctx.Logger().Info("Processing SyncUpdateOperatorsForGroup event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processUpdateOperatorsForGroupEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseOperatorSocketUpdate(*log); err == nil {
		ctx.Logger().Info("Processing OperatorSocketUpdate event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processOperatorSocketUpdateEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncEjectOperators(*log); err == nil {
		ctx.Logger().Info("Processing SyncEjectOperators event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processEjectOperatorsEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseChurnApproverUpdated(*log); err == nil {
		ctx.Logger().Info("Processing ChurnApproverUpdated event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processChurnApproverUpdatedEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseEjectorUpdated(*log); err == nil {
		ctx.Logger().Info("Processing EjectorUpdated event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processEjectorUpdatedEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncSetOperatorSetParams(*log); err == nil {
		ctx.Logger().Info("Processing SyncSetOperatorSetParams event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processSetOperatorSetParamsEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncSetGroupEjectionParams(*log); err == nil {
		ctx.Logger().Info("Processing SyncSetGroupEjectionParams event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processSetGroupEjectionParamsEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseSyncEjectionCooldown(*log); err == nil {
		ctx.Logger().Info("Processing SyncEjectionCooldown event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processEjectionCooldownEvent(ctx, txOrigin, event)
	}

	if event, err := filter.ParseMinimumStakeForGroupUpdated(*log); err == nil {
		ctx.Logger().Info("Processing MinimumStakeForGroupUpdated event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processMinimumStakeForGroupUpdatedEvent(ctx, txOrigin, event)
	}

	stakeFilter, err := stakeregistryrouter.NewStakeRegistryRouterFilterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	if event, err := stakeFilter.ParseSyncAddPools(*log); err == nil {
		ctx.Logger().Info("Processing SyncAddPools event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processAddPoolsEvent(ctx, txOrigin, event)
	}

	if event, err := stakeFilter.ParseSyncRemovePools(*log); err == nil {
		ctx.Logger().Info("Processing SyncRemovePools event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processRemovePoolsEvent(ctx, txOrigin, event)
	}

	if event, err := stakeFilter.ParseSyncModifyPoolParams(*log); err == nil {
		ctx.Logger().Info("Processing SyncModifyPoolParams event", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}
		return h.processModifyPoolParamsEvent(ctx, txOrigin, event)
	}

	ctx.Logger().Info("Processing logs finished", "toAddress", toAddress.Hex(), "txOrigin", txOrigin)
	return nil, nil
}
