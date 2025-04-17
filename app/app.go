package app

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	simappparams "cosmossdk.io/simapp/params"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	nftmodule "cosmossdk.io/x/nft/module"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	baseante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	cparams "github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	ibccallbacks "github.com/cosmos/ibc-go/modules/apps/callbacks"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee"
	ibcfeekeeper "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/keeper"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	evmante "github.com/evmos/ethermint/app/ante"
	enccodec "github.com/evmos/ethermint/encoding/codec"
	"github.com/evmos/ethermint/ethereum/eip712"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/ethermint/x/evm"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/ethermint/x/evm/vm/geth"
	"github.com/evmos/ethermint/x/feemarket"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	"github.com/0xPellNetwork/aegis/app/ante"
	"github.com/0xPellNetwork/aegis/docs/openapi"
	srvflags "github.com/0xPellNetwork/aegis/server/flags"
	authoritymodule "github.com/0xPellNetwork/aegis/x/authority"
	authoritykeeper "github.com/0xPellNetwork/aegis/x/authority/keeper"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	emissionsmodule "github.com/0xPellNetwork/aegis/x/emissions"
	emissionskeeper "github.com/0xPellNetwork/aegis/x/emissions/keeper"
	emissionstypes "github.com/0xPellNetwork/aegis/x/emissions/types"
	lightclientmodule "github.com/0xPellNetwork/aegis/x/lightclient"
	lightclientkeeper "github.com/0xPellNetwork/aegis/x/lightclient/keeper"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	pevmmodule "github.com/0xPellNetwork/aegis/x/pevm"
	pevmkeeper "github.com/0xPellNetwork/aegis/x/pevm/keeper"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayermodule "github.com/0xPellNetwork/aegis/x/relayer"
	relayerkeeper "github.com/0xPellNetwork/aegis/x/relayer/keeper"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	restakingmodule "github.com/0xPellNetwork/aegis/x/restaking"
	restakingkeeper "github.com/0xPellNetwork/aegis/x/restaking/keeper"
	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgmodule "github.com/0xPellNetwork/aegis/x/xmsg"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
	xsecuritymodule "github.com/0xPellNetwork/aegis/x/xsecurity"
	xsecuritykeeper "github.com/0xPellNetwork/aegis/x/xsecurity/keeper"
	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

const Name = "pellcore"

func init() {
	// manually update the power reduction by replacing micro (u) -> atto (a) evmos
	sdk.DefaultPowerReduction = ethermint.PowerReduction
	// modify fee market parameter defaults through global
	//feemarkettypes.DefaultMinGasPrice = v5.MainnetMinGasPrices
	//feemarkettypes.DefaultMinGasMultiplier = v5.MainnetMinGasMultiplier
}

var (
	AccountAddressPrefix = "pell"
	NodeDir              = ".pellcored"

	// AddrLen is the allowed length (in bytes) for an address.
	//
	// NOTE: In the SDK, the default value is 255.
	AddrLen = 20
)

var (
	// DefaultNodeHome default home directories for wasmd
	DefaultNodeHome = os.ExpandEnv("$HOME/") + NodeDir

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = AccountAddressPrefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = AccountAddressPrefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = AccountAddressPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
)

var (
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic([]govclient.ProposalHandler{
			paramsclient.ProposalHandler,
		}),
		cparams.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		vesting.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		nftmodule.AppModuleBasic{},
		consensus.AppModuleBasic{},

		evm.AppModuleBasic{},
		feemarket.AppModuleBasic{},
		authoritymodule.AppModuleBasic{},
		lightclientmodule.AppModuleBasic{},
		xmsgmodule.AppModuleBasic{},
		relayermodule.AppModuleBasic{},
		pevmmodule.AppModuleBasic{},
		emissionsmodule.AppModuleBasic{},
		restakingmodule.AppModuleBasic{},
		xsecuritymodule.AppModuleBasic{},

		// non sdk modules
		wasm.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		transfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		ibcfee.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		nft.ModuleName:                 nil,
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		ibcfeetypes.ModuleName:         nil,
		icatypes.ModuleName:            nil,
		//	wasmtypes.ModuleName:           {authtypes.Burner},

		xmsgtypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
		evmtypes.ModuleName:                             {authtypes.Minter, authtypes.Burner},
		pevmtypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
		emissionstypes.ModuleName:                       nil,
		emissionstypes.UndistributedObserverRewardsPool: nil,
		emissionstypes.UndistributedTssRewardsPool:      nil,
		wasmtypes.ModuleName:                            {authtypes.Burner},
		restakingtypes.ModuleName:                       nil,
		xsecuritytypes.ModuleName:                       nil,
	}

	// module accounts that are NOT allowed to receive tokens
	blockedReceivingModAcc = map[string]bool{
		distrtypes.ModuleName:          true,
		authtypes.FeeCollectorName:     true,
		stakingtypes.BondedPoolName:    true,
		stakingtypes.NotBondedPoolName: true,
		govtypes.ModuleName:            true,
	}
)

var (
	_ runtime.AppI            = (*PellApp)(nil)
	_ servertypes.Application = (*PellApp)(nil)
)

// PellApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type PellApp struct {
	*baseapp.BaseApp

	encodingCfg *simappparams.EncodingConfig

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper

	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCFeeKeeper        ibcfeekeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper
	WasmKeeper          wasmkeeper.Keeper

	ScopedIBCKeeper           capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilitykeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper      capabilitykeeper.ScopedKeeper
	ScopedIBCFeeKeeper        capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper          capabilitykeeper.ScopedKeeper

	// evm keepers
	EvmKeeper       *evmkeeper.Keeper
	FeeMarketKeeper feemarketkeeper.Keeper

	// pellchain keepers
	AuthorityKeeper   authoritykeeper.Keeper
	LightclientKeeper lightclientkeeper.Keeper
	XmsgKeeper        xmsgkeeper.Keeper
	RelayerKeeper     *relayerkeeper.Keeper
	PevmKeeper        pevmkeeper.Keeper
	EmissionsKeeper   emissionskeeper.Keeper
	RestakingKeeper   restakingkeeper.Keeper
	XSecurityKeeper   xsecuritykeeper.Keeper

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager
	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

// NewPellApp returns a reference to an initialized PellApp.
func NewPellApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *PellApp {
	encodingConfig := MakeEncodingConfig()
	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	eip712.SetEncodingConfig(encodingConfig)

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, crisistypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, consensusparamtypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, circuittypes.StoreKey,
		authzkeeper.StoreKey, nftkeeper.StoreKey, group.StoreKey,
		// non sdk store keys
		capabilitytypes.StoreKey, ibcexported.StoreKey, ibctransfertypes.StoreKey, ibcfeetypes.StoreKey,
		icahosttypes.StoreKey,
		icacontrollertypes.StoreKey,
		wasmtypes.StoreKey,
		evmtypes.StoreKey,
		feemarkettypes.StoreKey,
		authoritytypes.StoreKey,
		lightclienttypes.StoreKey,
		xmsgtypes.StoreKey,
		relayertypes.StoreKey,
		pevmtypes.StoreKey,
		emissionstypes.StoreKey,
		restakingtypes.StoreKey,
		xsecuritytypes.StoreKey,
	)

	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, evmtypes.TransientKey, feemarkettypes.TransientKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	app := &PellApp{
		BaseApp:     bApp,
		encodingCfg: &encodingConfig,
		keys:        keys,
		tkeys:       tkeys,
		memKeys:     memKeys,
	}
	if homePath == "" {
		homePath = DefaultNodeHome
	}

	app.ParamsKeeper = initParamsKeeper(
		encodingConfig.Codec,
		encodingConfig.Amino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		encodingConfig.Codec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.CapabilityKeeper.Seal()

	// add keepers
	// use custom Ethermint account for contracts
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		ethermint.ProtoAccount,
		maccPerms,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		encodingConfig.Codec,
		encodingConfig.Amino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.CrisisKeeper = crisiskeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(encodingConfig.Codec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper)

	app.CircuitKeeper = circuitkeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[circuittypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.AccountKeeper.AddressCodec(),
	)
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		encodingConfig.Codec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		// runtime.NewKVStoreService(keys[group.StoreKey]),
		encodingConfig.Codec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	// set the governance module account as the authority for conducting upgrades
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		encodingConfig.Codec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	binaryCodec := codec.NewProtoCodec(encodingConfig.InterfaceRegistry)

	app.IBCKeeper = ibckeeper.NewKeeper(
		binaryCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		stakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		// This should be removed. It is still in place to avoid failures of modules that have not yet been upgraded.
		AddRoute(paramproposal.RouterKey, cparams.NewParamChangeProposalHandler(app.ParamsKeeper))

	govConfig := govtypes.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	govKeeper := govkeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		stakingKeeper,
		app.DistrKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	app.NFTKeeper = nftkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[nftkeeper.StoreKey]),
		encodingConfig.Codec,
		app.AccountKeeper,
		app.BankKeeper,
	)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		stakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)

	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	// IBC Fee Module keeper
	app.IBCFeeKeeper = ibcfeekeeper.NewKeeper(
		encodingConfig.Codec, keys[ibcfeetypes.StoreKey],
		app.IBCKeeper.ChannelKeeper, // may be replaced with IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper, app.AccountKeeper, app.BankKeeper,
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.AuthorityKeeper = authoritykeeper.NewKeeper(
		encodingConfig.Codec,
		keys[authoritytypes.StoreKey],
		keys[authoritytypes.MemStoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName),
	)

	app.LightclientKeeper = lightclientkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[lightclienttypes.StoreKey],
		keys[lightclienttypes.MemStoreKey],
		app.AuthorityKeeper,
	)

	app.RelayerKeeper = relayerkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[relayertypes.StoreKey],
		keys[relayertypes.MemStoreKey],
		app.GetSubspace(relayertypes.ModuleName),
		*stakingKeeper,
		app.SlashingKeeper,
		app.AuthorityKeeper,
		app.LightclientKeeper,
	)

	app.StakingKeeper = stakingKeeper
	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks(), app.RelayerKeeper.Hooks()),
	)

	app.EmissionsKeeper = *emissionskeeper.NewKeeper(
		encodingConfig.Codec,
		keys[emissionstypes.StoreKey],
		keys[emissionstypes.MemStoreKey],
		app.GetSubspace(emissionstypes.ModuleName),
		authtypes.FeeCollectorName,
		app.BankKeeper,
		app.StakingKeeper,
		app.RelayerKeeper,
		app.AccountKeeper,
	)

	// Create Ethermint keepers
	tracer := cast.ToString(appOpts.Get(srvflags.EVMTracer))
	feeSs := app.GetSubspace(feemarkettypes.ModuleName)
	app.FeeMarketKeeper = feemarketkeeper.NewKeeper(
		encodingConfig.Codec, authtypes.NewModuleAddress(govtypes.ModuleName),
		runtime.NewKVStoreService(keys[feemarkettypes.StoreKey]), tkeys[feemarkettypes.TransientKey],
		feeSs,
	)
	evmSs := app.GetSubspace(evmtypes.ModuleName)
	app.EvmKeeper = evmkeeper.NewKeeper(
		encodingConfig.Codec, runtime.NewKVStoreService(keys[evmtypes.StoreKey]), tkeys[evmtypes.TransientKey], authtypes.NewModuleAddress(govtypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, stakingKeeper,
		&app.FeeMarketKeeper, nil, geth.NewEVM,
		tracer, evmSs,
	)

	app.PevmKeeper = *pevmkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[pevmtypes.StoreKey],
		keys[pevmtypes.MemStoreKey],
		app.AccountKeeper,
		app.EvmKeeper,
		app.BankKeeper,
		app.RelayerKeeper,
		app.AuthorityKeeper,
	)

	app.XmsgKeeper = *xmsgkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[xmsgtypes.StoreKey],
		keys[xmsgtypes.MemStoreKey],
		*stakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.RelayerKeeper,
		&app.PevmKeeper,
		app.AuthorityKeeper,
		app.LightclientKeeper,
	)

	app.RestakingKeeper = *restakingkeeper.NewKeeper(
		encodingConfig.Codec,
		keys[restakingtypes.StoreKey],
		keys[restakingtypes.MemStoreKey],
		app.AccountKeeper,
		app.EvmKeeper,
		app.BankKeeper,
		app.RelayerKeeper,
		app.AuthorityKeeper,
		&app.PevmKeeper,
		&app.XmsgKeeper,
	)

	app.XSecurityKeeper = *xsecuritykeeper.NewKeeper(
		encodingConfig.Codec,
		keys[xsecuritytypes.StoreKey],
		keys[xsecuritytypes.MemStoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
		app.PevmKeeper,
		app.RelayerKeeper,
		app.RestakingKeeper,
		app.AuthorityKeeper,
	)

	app.XmsgKeeper.SetInternalEventHooks(app.RestakingKeeper.Hooks())
	app.XmsgKeeper.SetXmsgResultHooks(app.RestakingKeeper.Hooks())

	app.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], encodingConfig.Codec, app.MsgServiceRouter(), app.AccountKeeper, group.Config{
		MaxExecutionPeriod: 2 * time.Hour, // Two hours.
		MaxMetadataLen:     255,
	})

	app.EvmKeeper = app.EvmKeeper.SetHooks(evmkeeper.NewMultiEvmHooks(
		app.XmsgKeeper.Hooks(),
		app.RestakingKeeper.Hooks(),
	))

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	app.WasmKeeper = wasmkeeper.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCFeeKeeper, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmkeeper.BuiltInCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// Create fee enabled wasm ibc Stack
	var wasmStack porttypes.IBCModule
	wasmStackIBCHandler := wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCFeeKeeper)
	wasmStack = ibcfee.NewIBCMiddleware(wasmStackIBCHandler, app.IBCFeeKeeper)

	// Create Interchain Accounts Stack
	// SendPacket, since it is originating from the application to core IBC:
	// icaAuthModuleKeeper.SendTx -> icaController.SendPacket -> fee.SendPacket -> channel.SendPacket
	var icaControllerStack porttypes.IBCModule
	// integration point for custom authentication modules
	// see https://medium.com/the-interchain-foundation/ibc-go-v6-changes-to-interchain-accounts-and-how-it-impacts-your-chain-806c185300d7
	var noAuthzModule porttypes.IBCModule
	icaControllerStack = icacontroller.NewIBCMiddleware(noAuthzModule, app.ICAControllerKeeper)
	// app.ICAAuthModule = icaControllerStack.(ibcmock.IBCModule)
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, app.ICAControllerKeeper)
	icaControllerStack = ibccallbacks.NewIBCMiddleware(icaControllerStack, app.IBCFeeKeeper, wasmStackIBCHandler, wasm.DefaultMaxIBCCallbackGas)
	icaICS4Wrapper := icaControllerStack.(porttypes.ICS4Wrapper)
	icaControllerStack = ibcfee.NewIBCMiddleware(icaControllerStack, app.IBCFeeKeeper)
	// Since the callbacks middleware itself is an ics4wrapper, it needs to be passed to the ica controller keeper
	app.ICAControllerKeeper.WithICS4Wrapper(icaICS4Wrapper)

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack porttypes.IBCModule
	icaHostStack = icahost.NewIBCModule(app.ICAHostKeeper)
	icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, app.IBCFeeKeeper)

	// Create Transfer Stack
	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = ibccallbacks.NewIBCMiddleware(transferStack, app.IBCFeeKeeper, wasmStackIBCHandler, wasm.DefaultMaxIBCCallbackGas)
	transferICS4Wrapper := transferStack.(porttypes.ICS4Wrapper)
	transferStack = ibcfee.NewIBCMiddleware(transferStack, app.IBCFeeKeeper)
	// Since the callbacks middleware itself is an ics4wrapper, it needs to be passed to the ica controller keeper
	app.TransferKeeper.WithICS4Wrapper(transferICS4Wrapper)

	// Create static IBC router, add app routes, then set and seal it
	ibcRouter := porttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(wasmtypes.ModuleName, wasmStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(encodingConfig.Codec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(encodingConfig.Codec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		feegrantmodule.NewAppModule(encodingConfig.Codec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.encodingCfg.InterfaceRegistry),
		gov.NewAppModule(encodingConfig.Codec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		slashing.NewAppModule(encodingConfig.Codec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.encodingCfg.InterfaceRegistry),
		distr.NewAppModule(encodingConfig.Codec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(encodingConfig.Codec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		cparams.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(encodingConfig.Codec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.encodingCfg.InterfaceRegistry),
		groupmodule.NewAppModule(encodingConfig.Codec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.encodingCfg.InterfaceRegistry),
		nftmodule.NewAppModule(encodingConfig.Codec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.encodingCfg.InterfaceRegistry),
		consensus.NewAppModule(encodingConfig.Codec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(encodingConfig.Codec, app.CircuitKeeper),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)), // always be last to make sure that it checks for all invariants and not only part of them
		ibctm.AppModule{},

		// non sdk modules
		capability.NewAppModule(encodingConfig.Codec, *app.CapabilityKeeper, false),
		wasm.NewAppModule(encodingConfig.Codec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		ibcfee.NewAppModule(app.IBCFeeKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),

		evm.NewAppModule(app.EvmKeeper, app.AccountKeeper, evmSs),
		feemarket.NewAppModule(app.FeeMarketKeeper, feeSs),
		authoritymodule.NewAppModule(encodingConfig.Codec, app.AuthorityKeeper),
		lightclientmodule.NewAppModule(encodingConfig.Codec, app.LightclientKeeper),
		xmsgmodule.NewAppModule(encodingConfig.Codec, app.XmsgKeeper),
		relayermodule.NewAppModule(encodingConfig.Codec, *app.RelayerKeeper),
		pevmmodule.NewAppModule(encodingConfig.Codec, app.PevmKeeper),
		emissionsmodule.NewAppModule(encodingConfig.Codec, app.EmissionsKeeper),
		restakingmodule.NewAppModule(encodingConfig.Codec, app.RestakingKeeper),
		xsecuritymodule.NewAppModule(encodingConfig.Codec, app.XSecurityKeeper),
		authzmodule.NewAppModule(encodingConfig.Codec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.encodingCfg.InterfaceRegistry),
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
				},
			),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(encodingConfig.Amino)
	app.BasicModuleManager.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0

	app.ModuleManager.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		group.ModuleName,
		vestingtypes.ModuleName,

		capabilitytypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		wasmtypes.ModuleName,

		emissionstypes.ModuleName,
		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		xmsgtypes.ModuleName,
		relayertypes.ModuleName,
		pevmtypes.ModuleName,
		restakingtypes.ModuleName,
		xsecuritytypes.ModuleName,
		authz.ModuleName,
		authoritytypes.ModuleName,
		lightclienttypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		banktypes.ModuleName,
		authtypes.ModuleName,
		upgradetypes.ModuleName,
		distrtypes.ModuleName,
		emissionstypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		govtypes.ModuleName,
		paramstypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		crisistypes.ModuleName,
		evmtypes.ModuleName,

		capabilitytypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		wasmtypes.ModuleName,

		evmtypes.ModuleName,
		feemarkettypes.ModuleName,
		xmsgtypes.ModuleName,
		relayertypes.ModuleName,
		pevmtypes.ModuleName,
		authz.ModuleName,
		authoritytypes.ModuleName,
		lightclienttypes.ModuleName,
		restakingtypes.ModuleName,
		xsecuritytypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	// NOTE: Cross-chain module must be initialized after observer module, as pending nonces in crosschain needs the tss pubkey from observer module
	app.ModuleManager.SetOrderInitGenesis(InitGenesisModuleList()...)
	app.ModuleManager.SetOrderExportGenesis(InitGenesisModuleList()...)

	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(encodingConfig.Codec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.ModuleManager.RegisterServices(app.configurator)

	app.RegisterUpgradeHandlers()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(encodingConfig.TxConfig, wasmConfig, keys[wasmtypes.StoreKey])

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	// see cmd/wasmd/root.go: 206 - 214 approx
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedWasmKeeper = scopedWasmKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper

	app.setPostHandler()

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	// https://github.com/CosmWasm/wasmd/issues/1785  tmp fix
	protoFiles, err := proto.MergedRegistry()
	// if err != nil {
	// 	panic(err)
	// }

	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// // Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

// Name returns the name of the App
func (app *PellApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *PellApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *PellApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *PellApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *PellApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	return app.ModuleManager.InitGenesis(ctx, app.encodingCfg.Codec, genesisState)
}

// LoadHeight loads a particular height
func (app *PellApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *PellApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *PellApp) LegacyAmino() *codec.LegacyAmino {
	return app.encodingCfg.Amino
}

// AppCodec returns Pell app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *PellApp) AppCodec() codec.Codec {
	return app.encodingCfg.Codec
}

// InterfaceRegistry returns Gaia's InterfaceRegistry
func (app *PellApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.encodingCfg.InterfaceRegistry
}

// TxConfig returns Gaia's InterfaceRegistry
func (app *PellApp) TxConfig() client.TxConfig {
	return app.encodingCfg.TxConfig
}

// Amino returns Gaia's InterfaceRegistry
func (app *PellApp) Amino() *codec.LegacyAmino {
	return app.encodingCfg.Amino
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *PellApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *PellApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *PellApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *PellApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *PellApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register legacy tx routes.
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	if apiConfig.Swagger {
		openapi.RegisterOpenAPIService(apiSvr.Router)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *PellApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.encodingCfg.InterfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *PellApp) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.encodingCfg.InterfaceRegistry,
		app.Query,
	)
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(upgradetypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(evidencetypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(genutiltypes.ModuleName)
	paramsKeeper.Subspace(feegrant.ModuleName)
	paramsKeeper.Subspace(group.ModuleName)
	paramsKeeper.Subspace(vestingtypes.ModuleName)

	paramsKeeper.Subspace(capabilitytypes.ModuleName)

	// register the IBC key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	paramsKeeper.Subspace(icatypes.ModuleName)

	paramsKeeper.Subspace(wasmtypes.ModuleName)
	//pkt := ibctransfertypes.ParamKeyTable().RegisterParamSet(&ibccoreclienttypes.Params{}).RegisterParamSet(&ibcconnectiontypes.Params{})
	// ethermint subspaces
	paramsKeeper.Subspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())
	paramsKeeper.Subspace(xmsgtypes.ModuleName)
	paramsKeeper.Subspace(relayertypes.ModuleName)
	paramsKeeper.Subspace(pevmtypes.ModuleName)
	paramsKeeper.Subspace(emissionstypes.ModuleName)
	paramsKeeper.Subspace(authoritytypes.ModuleName)
	paramsKeeper.Subspace(lightclienttypes.ModuleName)
	paramsKeeper.Subspace(restakingtypes.ModuleName)
	paramsKeeper.Subspace(xsecuritytypes.ModuleName)

	return paramsKeeper
}

// VerifyAddressFormat verifies the address is compatible with ethereum
func VerifyAddressFormat(bz []byte) error {
	if len(bz) == 0 {
		return cosmoserrors.Wrap(sdkerrors.ErrUnknownAddress, "invalid address; cannot be empty")
	}
	if len(bz) != AddrLen {
		return cosmoserrors.Wrapf(
			sdkerrors.ErrUnknownAddress,
			"invalid address length; got: %d, expect: %d", len(bz), AddrLen,
		)
	}

	return nil
}

// SimulationManager implements the SimulationApp interface
func (app *PellApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *PellApp) BlockedAddrs() map[string]bool {
	blockList := make(map[string]bool)
	for k, v := range blockedReceivingModAcc {
		addr := authtypes.NewModuleAddress(k)
		blockList[addr.String()] = v
	}
	return blockList
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range blockedReceivingModAcc {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}

func (app *PellApp) setAnteHandler(txConfig client.TxConfig, wasmConfig wasmtypes.WasmConfig, txCounterStoreKey *storetypes.KVStoreKey) {
	options := ante.HandlerOptions{
		HandlerOptions: baseante.HandlerOptions{
			AccountKeeper:          app.AccountKeeper,
			BankKeeper:             app.BankKeeper,
			SignModeHandler:        txConfig.SignModeHandler(),
			FeegrantKeeper:         app.FeeGrantKeeper,
			SigGasConsumer:         evmante.DefaultSigVerificationGasConsumer,
			TxFeeChecker:           evmante.NewDynamicFeeChecker(app.EvmKeeper),
			ExtensionOptionChecker: ethermint.HasDynamicFeeExtensionOption,
		},
		IBCKeeper:             app.IBCKeeper,
		TXCounterStoreService: runtime.NewKVStoreService(txCounterStoreKey),
		WasmKeeper:            &app.WasmKeeper,
		WasmConfig:            &wasmConfig,
		EvmKeeper:             app.EvmKeeper,
		FeeMarketKeeper:       app.FeeMarketKeeper,
		AccountKeeper:         app.AccountKeeper,
		BankKeeper:            app.BankKeeper,
		DisabledAuthzMsgs: []string{
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}), // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreatePermanentLockedAccount{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreatePeriodicVestingAccount{}),
		},
		MaxTxGasWanted: math.MaxUint64,
		CircuitKeeper:  &app.CircuitKeeper,
		RelayerKeeper:  app.RelayerKeeper,
	}

	anteHandler, err := ante.NewAnteHandler(options)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}

func (app *PellApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

func (app *PellApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// AutoCliOpts returns the autocli options for the app.
func (app *PellApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}
