package keeper

import (
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethermint "github.com/evmos/ethermint/types"
	evmmodule "github.com/evmos/ethermint/x/evm"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/ethermint/x/evm/vm/geth"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritymodule "github.com/0xPellNetwork/aegis/x/authority"
	authoritykeeper "github.com/0xPellNetwork/aegis/x/authority/keeper"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	emissionsmodule "github.com/0xPellNetwork/aegis/x/emissions"
	emissionskeeper "github.com/0xPellNetwork/aegis/x/emissions/keeper"
	emissionstypes "github.com/0xPellNetwork/aegis/x/emissions/types"
	lightclientkeeper "github.com/0xPellNetwork/aegis/x/lightclient/keeper"
	pevmmodule "github.com/0xPellNetwork/aegis/x/pevm"
	pevmkeeper "github.com/0xPellNetwork/aegis/x/pevm/keeper"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayermodule "github.com/0xPellNetwork/aegis/x/relayer"
	relayerkeeper "github.com/0xPellNetwork/aegis/x/relayer/keeper"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgmodule "github.com/0xPellNetwork/aegis/x/xmsg"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// NewContext creates a new sdk.Context for testing purposes with initialized header
func NewContext(stateStore store.CommitMultiStore) sdk.Context {
	header := tmproto.Header{
		Height:  1,
		ChainID: "ignite_186-1",
		Time:    time.Now().UTC(),
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}
	ctx := sdk.NewContext(stateStore, header, false, log.NewNopLogger())
	ctx = ctx.WithHeaderHash(tmhash.Sum([]byte("header")))
	return ctx
}

// SDKKeepers is a struct containing regular SDK module keepers for test purposes
type SDKKeepers struct {
	ParamsKeeper    paramskeeper.Keeper
	AuthKeeper      authkeeper.AccountKeeper
	BankKeeper      bankkeeper.Keeper
	StakingKeeper   stakingkeeper.Keeper
	SlashingKeeper  slashingkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper
	EvmKeeper       *evmkeeper.Keeper
}

// PellKeepers is a struct containing Pell module keepers for test purposes
type PellKeepers struct {
	AuthorityKeeper   *authoritykeeper.Keeper
	EmissionsKeeper   *emissionskeeper.Keeper
	ObserverKeeper    *relayerkeeper.Keeper
	LightclientKeeper *lightclientkeeper.Keeper
	XmsgKeeper        *xmsgkeeper.Keeper
	PevmKeeper        *pevmkeeper.Keeper
}

var moduleAccountPerms = map[string][]string{
	authtypes.FeeCollectorName:                      nil,
	distrtypes.ModuleName:                           nil,
	stakingtypes.BondedPoolName:                     {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:                  {authtypes.Burner, authtypes.Staking},
	evmtypes.ModuleName:                             {authtypes.Minter, authtypes.Burner},
	xmsgtypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
	pevmtypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
	emissionstypes.ModuleName:                       {authtypes.Minter},
	emissionstypes.UndistributedObserverRewardsPool: nil,
	emissionstypes.UndistributedTssRewardsPool:      nil,
}

// ModuleAccountAddrs returns all the app's module account addresses.
func ModuleAccountAddrs(maccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// ParamsKeeper instantiates a param keeper for testing purposes
func ParamsKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
) paramskeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(paramstypes.StoreKey)
	tkeys := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)

	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ss.MountStoreWithDB(tkeys, storetypes.StoreTypeTransient, db)

	return paramskeeper.NewKeeper(cdc, pevmtypes.Amino, storeKey, tkeys)
}

// AccountKeeper instantiates an account keeper for testing purposes
func AccountKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	paramKeeper paramskeeper.Keeper,
) authkeeper.AccountKeeper {
	storeKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	return authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		ethermint.ProtoAccount,
		moduleAccountPerms,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

// BankKeeper instantiates a bank keeper for testing purposes
func BankKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	paramKeeper paramskeeper.Keeper,
	authKeeper authkeeper.AccountKeeper,
) bankkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(banktypes.StoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	blockedAddrs := make(map[string]bool)

	return bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authKeeper,
		blockedAddrs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)
}

// StakingKeeper instantiates a staking keeper for testing purposes
func StakingKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	paramKeeper paramskeeper.Keeper,
) *stakingkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	return stakingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
}

// SlashingKeeper instantiates a slashing keeper for testing purposes
func SlashingKeeper(
	cdc codec.Codec,
	amino *codec.LegacyAmino,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	stakingKeeper stakingkeeper.Keeper,
	paramKeeper paramskeeper.Keeper,
) slashingkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	return slashingkeeper.NewKeeper(
		cdc,
		amino,
		runtime.NewKVStoreService(storeKey),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

// DistributionKeeper instantiates a distribution keeper for testing purposes
func DistributionKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
	paramKeeper paramskeeper.Keeper,
) distrkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(distrtypes.StoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	return distrkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

// ProtocolVersionSetter mock
type ProtocolVersionSetter struct{}

func (vs ProtocolVersionSetter) SetProtocolVersion(uint64) {}

// UpgradeKeeper instantiates an upgrade keeper for testing purposes
func UpgradeKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
) *upgradekeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(upgradetypes.StoreKey)
	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	skipUpgradeHeights := make(map[int64]bool)
	vs := ProtocolVersionSetter{}

	return upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(storeKey),
		cdc,
		"",
		vs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

// FeeMarketKeeper instantiates a feemarket keeper for testing purposes
func FeeMarketKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	paramKeeper paramskeeper.Keeper,
) feemarketkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(feemarkettypes.StoreKey)
	transientKey := storetypes.NewTransientStoreKey(feemarkettypes.TransientKey)

	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ss.MountStoreWithDB(transientKey, storetypes.StoreTypeTransient, db)

	return feemarketkeeper.NewKeeper(
		cdc,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		runtime.NewKVStoreService(storeKey),
		transientKey,
		paramKeeper.Subspace(feemarkettypes.ModuleName),
	)
}

// EVMKeeper instantiates an evm keeper for testing purposes
func EVMKeeper(
	cdc codec.Codec,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
	authKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	feemarketKeeper feemarketkeeper.Keeper,
	paramKeeper paramskeeper.Keeper,
) *evmkeeper.Keeper {
	storeKey := storetypes.NewKVStoreKey(evmtypes.StoreKey)
	transientKey := storetypes.NewTransientStoreKey(evmtypes.TransientKey)

	ss.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ss.MountStoreWithDB(transientKey, storetypes.StoreTypeTransient, db)

	k := evmkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		transientKey,
		authtypes.NewModuleAddress(govtypes.ModuleName),
		authKeeper,
		bankKeeper,
		stakingKeeper,
		feemarketKeeper,
		nil,
		geth.NewEVM,
		"",
		paramKeeper.Subspace(evmtypes.ModuleName),
	)

	return k
}

// NewSDKKeepers instantiates regular Cosmos SDK keeper such as staking with local storage for testing purposes
func NewSDKKeepers(
	cdc codec.Codec,
	amino *codec.LegacyAmino,
	db *dbm.MemDB,
	ss store.CommitMultiStore,
) SDKKeepers {
	paramsKeeper := ParamsKeeper(cdc, db, ss)
	authKeeper := AccountKeeper(cdc, db, ss, paramsKeeper)
	bankKeeper := BankKeeper(cdc, db, ss, paramsKeeper, authKeeper)
	stakingKeeper := StakingKeeper(cdc, db, ss, authKeeper, bankKeeper, paramsKeeper)
	feeMarketKeeper := FeeMarketKeeper(cdc, db, ss, paramsKeeper)
	evmKeeper := EVMKeeper(cdc, db, ss, authKeeper, bankKeeper, *stakingKeeper, feeMarketKeeper, paramsKeeper)
	slashingKeeper := SlashingKeeper(cdc, amino, db, ss, *stakingKeeper, paramsKeeper)
	return SDKKeepers{
		ParamsKeeper:    paramsKeeper,
		AuthKeeper:      authKeeper,
		BankKeeper:      bankKeeper,
		StakingKeeper:   *stakingKeeper,
		FeeMarketKeeper: feeMarketKeeper,
		EvmKeeper:       evmKeeper,
		SlashingKeeper:  slashingKeeper,
	}
}

// InitGenesis initializes the test modules genesis state
func (sdkk SDKKeepers) InitGenesis(ctx sdk.Context) {
	sdkk.AuthKeeper.InitGenesis(ctx, *authtypes.DefaultGenesisState())
	sdkk.BankKeeper.InitGenesis(ctx, banktypes.DefaultGenesisState())
	sdkk.StakingKeeper.InitGenesis(ctx, stakingtypes.DefaultGenesisState())
	evmGenesis := *evmtypes.DefaultGenesisState()
	evmGenesis.Params.EvmDenom = "apell"
	evmmodule.InitGenesis(ctx, sdkk.EvmKeeper, sdkk.AuthKeeper, evmGenesis)
}

// InitBlockProposer initialize the block proposer for test purposes with an associated validator
func (sdkk SDKKeepers) InitBlockProposer(t testing.TB, ctx sdk.Context) sdk.Context {
	// #nosec G404 test purpose - weak randomness is not an issue here
	r := rand.New(rand.NewSource(42))

	// Set validator in the store
	validator := sample.Validator(t, r)
	sdkk.StakingKeeper.SetValidator(ctx, validator)
	err := sdkk.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	require.NoError(t, err)

	// Validator is proposer
	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	return ctx.WithProposer(consAddr)
}

// InitGenesis initializes the test modules genesis state for defined Pell modules
func (zk PellKeepers) InitGenesis(ctx sdk.Context) {
	if zk.AuthorityKeeper != nil {
		authoritymodule.InitGenesis(ctx, *zk.AuthorityKeeper, *authoritytypes.DefaultGenesis())
	}
	if zk.EmissionsKeeper != nil {
		emissionsmodule.InitGenesis(ctx, *zk.EmissionsKeeper, *emissionstypes.DefaultGenesis())
	}
	if zk.ObserverKeeper != nil {
		relayermodule.InitGenesis(ctx, *zk.ObserverKeeper, *relayertypes.DefaultGenesis())
	}
	if zk.XmsgKeeper != nil {
		xmsgmodule.InitGenesis(ctx, *zk.XmsgKeeper, *xmsgtypes.DefaultGenesis())
	}
	if zk.PevmKeeper != nil {
		pevmmodule.InitGenesis(ctx, *zk.PevmKeeper, *pevmtypes.DefaultGenesis())
	}
}
