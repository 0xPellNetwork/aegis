package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

type Keeper struct {
	cdc               codec.BinaryCodec
	storeKey          storetypes.StoreKey
	memKey            storetypes.StoreKey
	paramstore        paramtypes.Subspace
	stakingKeeper     types.StakingKeeper
	slashingKeeper    types.SlashingKeeper
	authorityKeeper   types.AuthorityKeeper
	lightclientKeeper types.LightclientKeeper
	pevmKeeper        types.PevmKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	stakingKeeper types.StakingKeeper,
	slashinKeeper types.SlashingKeeper,
	authorityKeeper types.AuthorityKeeper,
	lightclientKeeper types.LightclientKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		memKey:            memKey,
		paramstore:        ps,
		stakingKeeper:     stakingKeeper,
		slashingKeeper:    slashinKeeper,
		authorityKeeper:   authorityKeeper,
		lightclientKeeper: lightclientKeeper,
	}
}

func (k Keeper) GetSlashingKeeper() types.SlashingKeeper {
	return k.slashingKeeper
}

func (k Keeper) GetStakingKeeper() types.StakingKeeper {
	return k.stakingKeeper
}

func (k Keeper) GetAuthorityKeeper() types.AuthorityKeeper {
	return k.authorityKeeper
}

func (k Keeper) GetLightclientKeeper() types.LightclientKeeper {
	return k.lightclientKeeper
}

func (k *Keeper) SetPevmKeeper(pevmKeeper types.PevmKeeper) {
	k.pevmKeeper = pevmKeeper
}

func (k Keeper) GetPevmKeeper() types.PevmKeeper {
	return k.pevmKeeper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}
