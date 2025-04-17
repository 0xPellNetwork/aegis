package keeper

import (
	"reflect"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type RegistryRouterFactoryEventSubscriber struct {
	Keeper
}

func NewRegistryRouterFactoryEventSubscriber(keeper Keeper) *RegistryRouterFactoryEventSubscriber {
	return &RegistryRouterFactoryEventSubscriber{Keeper: keeper}
}

// ProcessLogs processes the logs
func (h RegistryRouterFactoryEventSubscriber) ProcessLogs(ctx sdk.Context, _ uint64, toAddress ethcommon.Address, log *ethtypes.Log, txOrigin string) (*sdkmath.Int, error) {
	// logs from registry router factory
	factoryFilterer, err := registryrouterfactory.NewRegistryRouterFactoryFilterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, nil
	}

	if event, err := factoryFilterer.ParseRegistryRouterCreated(*log); err == nil {
		if !strings.EqualFold(event.Raw.Address.Hex(), log.Address.Hex()) {
			return nil, nil
		}

		return nil, h.processRegistryRouterCreatedEvent(ctx, txOrigin, event)
	}

	return nil, nil
}

// processRegistryRouterCreatedEvent processes the registry router created event
func (h *RegistryRouterFactoryEventSubscriber) processRegistryRouterCreatedEvent(ctx sdk.Context, txOrigin string, event *registryrouterfactory.RegistryRouterFactoryRegistryRouterCreated) error {
	if err := h.processStoreRegistryRouterCreatedEvent(ctx, event); err != nil {
		ctx.Logger().Error("Failed to process event by pevm", "event", reflect.TypeOf(event), "error", err)
		return nil
	}

	return nil
}

// processStoreRegistryRouterCreatedEvent processes the registry router created event
func (h *RegistryRouterFactoryEventSubscriber) processStoreRegistryRouterCreatedEvent(ctx sdk.Context, event *registryrouterfactory.RegistryRouterFactoryRegistryRouterCreated) error {
	if err := h.AddRegistryRouterAddress(ctx, []ethcommon.Address{
		event.RegistryRouter,
		event.StakeRegistryRouter,
	}); err != nil {
		ctx.Logger().Error("Failed to add registry router address", "error", err)
		return err
	}

	return nil
}
