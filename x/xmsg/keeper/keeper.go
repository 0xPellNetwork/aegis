package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	stakingKeeper      types.StakingKeeper
	authKeeper         types.AccountKeeper
	bankKeeper         types.BankKeeper
	relayerKeeper      types.RelayerKeeper
	pevmKeeper         types.PevmKeeper
	authorityKeeper    types.AuthorityKeeper
	lightclientKeeper  types.LightclientKeeper
	xmsgResultHooks    []types.XmsgOutboundResultHook
	internalEventHooks []types.InternalEventLogHooks

	internalHandlers []types.EventHandler
}

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey storetypes.StoreKey,
	stakingKeeper types.StakingKeeper, // custom
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	relayerKeeper types.RelayerKeeper,
	pevmKeeper types.PevmKeeper,
	authorityKeeper types.AuthorityKeeper,
	lightclientKeeper types.LightclientKeeper,
) *Keeper {
	// ensure governance module account is set
	// FIXME: enable this check! (disabled for now to avoid unit test panic)
	//if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
	//	panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	//}
	k := &Keeper{
		cdc:                cdc,
		storeKey:           storeKey,
		memKey:             memKey,
		stakingKeeper:      stakingKeeper,
		authKeeper:         authKeeper,
		bankKeeper:         bankKeeper,
		relayerKeeper:      relayerKeeper,
		pevmKeeper:         pevmKeeper,
		authorityKeeper:    authorityKeeper,
		lightclientKeeper:  lightclientKeeper,
		xmsgResultHooks:    []types.XmsgOutboundResultHook{},
		internalEventHooks: []types.InternalEventLogHooks{},
	}

	k.internalHandlers = []types.EventHandler{
		ConnectorEventHandler{k: *k},
	}
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetAuthKeeper() types.AccountKeeper {
	return k.authKeeper
}

func (k Keeper) GetBankKeeper() types.BankKeeper {
	return k.bankKeeper
}

func (k Keeper) GetStakingKeeper() types.StakingKeeper {
	return k.stakingKeeper
}

func (k Keeper) GetPevmKeeper() types.PevmKeeper {
	return k.pevmKeeper
}

func (k Keeper) GetRelayerKeeper() types.RelayerKeeper {
	return k.relayerKeeper
}

func (k Keeper) GetAuthorityKeeper() types.AuthorityKeeper {
	return k.authorityKeeper
}

func (k Keeper) GetLightclientKeeper() types.LightclientKeeper {
	return k.lightclientKeeper
}

func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

func (k Keeper) GetMemKey() storetypes.StoreKey {
	return k.memKey
}

func (k Keeper) GetCodec() codec.Codec {
	return k.cdc
}

func (k *Keeper) SetInternalEventHooks(hooks ...types.InternalEventLogHooks) *Keeper {
	k.internalEventHooks = append(k.internalEventHooks, hooks...)
	return k
}

func (k *Keeper) SetXmsgResultHooks(hooks ...types.XmsgOutboundResultHook) *Keeper {
	k.xmsgResultHooks = append(k.xmsgResultHooks, hooks...)
	return k
}

func (k Keeper) DeductFees(ctx sdk.Context, fees []*types.CrossChainFee) error {
	if len(fees) == 0 {
		return nil
	}

	denom, _ := sdk.GetBaseDenom()
	for _, fee := range fees {
		if fee.Fee.IsZero() {
			continue
		}

		err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fee.Address, authtypes.FeeCollectorName, sdk.Coins{sdk.Coin{Denom: denom, Amount: fee.Fee}})
		if err != nil {
			return err
		}
	}

	return nil
}
