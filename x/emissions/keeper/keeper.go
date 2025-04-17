package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/emissions/types"
)

type (
	Keeper struct {
		cdc              codec.BinaryCodec
		storeKey         storetypes.StoreKey
		memKey           storetypes.StoreKey
		paramStore       types.ParamStore
		feeCollectorName string
		bankKeeper       types.BankKeeper
		stakingKeeper    types.StakingKeeper
		observerKeeper   types.RelayerKeeper
		authKeeper       types.AccountKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps types.ParamStore,
	feeCollectorName string,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	observerKeeper types.RelayerKeeper,
	authKeeper types.AccountKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		paramStore:       ps,
		feeCollectorName: feeCollectorName,
		bankKeeper:       bankKeeper,
		stakingKeeper:    stakingKeeper,
		observerKeeper:   observerKeeper,
		authKeeper:       authKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetFeeCollector() string {
	return k.feeCollectorName
}

func (k Keeper) GetBankKeeper() types.BankKeeper {
	return k.bankKeeper
}

func (k Keeper) GetStakingKeeper() types.StakingKeeper {
	return k.stakingKeeper
}

func (k Keeper) GetObserverKeeper() types.RelayerKeeper {
	return k.observerKeeper
}

func (k Keeper) GetAuthKeeper() types.AccountKeeper {
	return k.authKeeper
}

func (k Keeper) GetParamStore() types.ParamStore {
	return k.paramStore
}
