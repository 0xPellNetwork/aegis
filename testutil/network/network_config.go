package network

import (
	"fmt"
	"time"

	pruningtypes "cosmossdk.io/store/pruning/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/0xPellNetwork/aegis/app"
	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
)

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() Config {
	encoding := app.MakeEncodingConfig()

	appConstructor := func(val ValidatorI) servertypes.Application {
		return app.NewPellApp(
			val.GetCtx().Logger,
			dbm.NewMemDB(),
			nil,
			true,
			map[int64]bool{},
			val.GetCtx().Config.RootDir,
			0,
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			[]wasmkeeper.Option{wasmkeeper.WithWasmEngine(&wasmtesting.MockWasmEngine{})},
			baseapp.SetChainID(""),
		)
	}

	return Config{
		Codec:             encoding.Codec,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor:    appConstructor,

		GenesisState:  app.ModuleBasics.DefaultGenesis(encoding.Codec),
		TimeoutCommit: 2 * time.Second,
		ChainID:       "ignite_8888-2",
		NumValidators: 10,
		Mnemonics: []string{
			"race draft rival universe maid cheese steel logic crowd fork comic easy truth drift tomorrow eye buddy head time cash swing swift midnight borrow",
			"hand inmate canvas head lunar naive increase recycle dog ecology inhale december wide bubble hockey dice worth gravity ketchup feed balance parent secret orchard",
			"cool little feel apple shoulder member menu owner sure update combine execute copper candy orient record pioneer wet vapor junior quiz choice topic logic",
			"result guess around primary tissue tiger witness tired canyon clog gift field merry tribe honey popular bring cargo cricket crew hand arrow quantum broom",
			"canyon impact autumn parrot sister roof father wing valve result matrix subject step similar actor effort lake comic patch moral lobster charge veteran barely",
			"pulp false tongue shield brave broom hurdle attract laugh taxi victory budget fox spirit abstract inside avoid win more cigar perfect opera attract clump",
			"idea oxygen faculty harsh citizen section group carbon waste symbol village inspire slim acquire grab donate champion diary north come kitchen emotion dance melody",
			"tortoise wife false victory define seek frequent nasty answer wire erosion thumb scrub seek cluster state analyst addict antique panic century image radar agree",
			"bacon weird jazz control lumber pottery install parrot paper range license flip gadget cargo armor they pioneer media ordinary agent adjust primary doll access",
			"muffin market delay mutual abandon swamp order orbit rose easy sunny retire autumn weekend involve pelican elbow gesture current chicken stock theme antique fringe",
		},
		BondDenom:       config.BaseDenom,
		MinGasPrices:    fmt.Sprintf("0.000006%s", config.BaseDenom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}
}
