package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

type Keeper struct {
	cdc                   codec.BinaryCodec
	storeKey              storetypes.StoreKey
	memKey                storetypes.StoreKey
	authKeeper            types.AccountKeeper
	evmKeeper             types.EVMKeeper
	bankKeeper            types.BankKeeper
	relayerKeeper         types.RelayerKeeper
	authorityKeeper       types.AuthorityKeeper
	pevmKeeper            types.PevmKeeper
	eventHandler          []xmsgtypes.EventHandler
	xmsgKeeper            types.XmsgKeeper
	middlewareSyncHandler *MiddlewareSyncHandler
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	evmKeeper types.EVMKeeper,
	bankKeeper types.BankKeeper,
	relayerKeeper types.RelayerKeeper,
	authorityKeeper types.AuthorityKeeper,
	pevmKeeper types.PevmKeeper,
	xmsgKeeper types.XmsgKeeper,
) *Keeper {
	keeper := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		authKeeper:      authKeeper,
		evmKeeper:       evmKeeper,
		bankKeeper:      bankKeeper,
		relayerKeeper:   relayerKeeper,
		authorityKeeper: authorityKeeper,
		pevmKeeper:      pevmKeeper,
		xmsgKeeper:      xmsgKeeper,
	}

	keeper.relayerKeeper.SetPevmKeeper(pevmKeeper)

	keeper.middlewareSyncHandler = NewMiddlewareHistoryEventHandler(*keeper)
	middlewareHandler := NewMiddlewareEventHandler(*keeper)
	middlewareHandler.RegisterAllEventSubscriber()

	keeper.eventHandler = []xmsgtypes.EventHandler{
		DelegationHandler{Keeper: *keeper},
		OperatorHandler{Keeper: *keeper},
		middlewareHandler,
	}

	return keeper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetAuthKeeper() types.AccountKeeper {
	return k.authKeeper
}

func (k Keeper) GetEVMKeeper() types.EVMKeeper {
	return k.evmKeeper
}

func (k Keeper) GetBankKeeper() types.BankKeeper {
	return k.bankKeeper
}

func (k Keeper) GetRelayerKeeper() types.RelayerKeeper {
	return k.relayerKeeper
}

func (k Keeper) GetAuthorityKeeper() types.AuthorityKeeper {
	return k.authorityKeeper
}

func (k Keeper) Cdc() codec.BinaryCodec {
	return k.cdc
}
