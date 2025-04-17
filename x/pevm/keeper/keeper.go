package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	memKey          storetypes.StoreKey
	authKeeper      types.AccountKeeper
	evmKeeper       types.EVMKeeper
	bankKeeper      types.BankKeeper
	relayerKeeper   types.RelayerKeeper
	authorityKeeper types.AuthorityKeeper
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
	}

	keeper.relayerKeeper.SetPevmKeeper(keeper)
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
