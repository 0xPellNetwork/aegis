package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	memKey          storetypes.StoreKey
	stakingKeeper   types.StakingKeeper
	slashingKeeper  types.SlashingKeeper
	pevmKeeper      types.PevmKeeper
	relayerKeeper   types.RelayerKeeper
	restakingKeeper types.RestakingKeeper
	authorityKeeper types.AuthorityKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey, memKey storetypes.StoreKey,
	stakingKeeper types.StakingKeeper,
	slashingKeeper types.SlashingKeeper,
	pevmKeeper types.PevmKeeper,
	relayerKeeper types.RelayerKeeper,
	restakingKeeper types.RestakingKeeper,
	authorityKeeper types.AuthorityKeeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		stakingKeeper:   stakingKeeper,
		slashingKeeper:  slashingKeeper,
		pevmKeeper:      pevmKeeper,
		relayerKeeper:   relayerKeeper,
		restakingKeeper: restakingKeeper,
		authorityKeeper: authorityKeeper,
	}
}

func (k Keeper) GetRelayerKeeper() types.RelayerKeeper {
	return k.relayerKeeper
}

func (k Keeper) GetPevmKeeper() types.PevmKeeper {
	return k.pevmKeeper
}

func (k Keeper) GetStakingKeeper() types.StakingKeeper {
	return k.stakingKeeper
}

func (k Keeper) GetRestakingKeeper() types.RestakingKeeper {
	return k.restakingKeeper
}

func (k Keeper) GetSlashingKeeper() types.SlashingKeeper {
	return k.slashingKeeper
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

func (k Keeper) GetAuthorityKeeper() types.AuthorityKeeper {
	return k.authorityKeeper
}
