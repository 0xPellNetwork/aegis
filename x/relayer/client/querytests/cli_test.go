package querytests

// import (
// 	"fmt"
// 	"testing"
// 	"time"

// 	"cosmossdk.io/simapp"
// 	pruningtypes "cosmossdk.io/store/pruning/types"
// 	"github.com/cosmos/cosmos-sdk/baseapp"
// 	"github.com/cosmos/cosmos-sdk/crypto/hd"
// 	"github.com/cosmos/cosmos-sdk/crypto/keyring"
// 	servertypes "github.com/cosmos/cosmos-sdk/server/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// 	"github.com/0xPellNetwork/aegis/app"
// 	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
// 	"github.com/0xPellNetwork/aegis/testutil/network"
// 	"github.com/stretchr/testify/suite"
// 	tmdb "github.com/tendermint/tm-db"
// )

// func TestCLIQuerySuite(t *testing.T) {
// 	cfg := CliTestConfig()
// 	suite.Run(t, NewCLITestSuite(cfg))
// }

// func CliTestConfig() network.Config {
// 	encoding := app.MakeEncodingConfig()
// 	return network.Config{
// 		Codec:             encoding.Codec,
// 		TxConfig:          encoding.TxConfig,
// 		LegacyAmino:       encoding.Amino,
// 		InterfaceRegistry: encoding.InterfaceRegistry,
// 		AccountRetriever:  authtypes.AccountRetriever{},
// 		AppConstructor: func(val network.Validator) servertypes.Application {
// 			return app.New(
// 				val.Ctx.Logger, dbm.NewMemDB(), nil, true, map[int64]bool{}, val.Ctx.Config.RootDir, 0,
// 				encoding,
// 				simapp.EmptyAppOptions{},
// 				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
// 				baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
// 			)
// 		},
// 		GenesisState:    app.ModuleBasics.DefaultGenesis(encoding.Codec),
// 		TimeoutCommit:   2 * time.Second,
// 		ChainID:         "athens_8888-2",
// 		NumOfValidators: 2,
// 		Mnemonics: []string{
// 			"race draft rival universe maid cheese steel logic crowd fork comic easy truth drift tomorrow eye buddy head time cash swing swift midnight borrow",
// 			"hand inmate canvas head lunar naive increase recycle dog ecology inhale december wide bubble hockey dice worth gravity ketchup feed balance parent secret orchard",
// 		},
// 		BondDenom:       config.BaseDenom,
// 		MinGasPrices:    fmt.Sprintf("0.000006%s", config.BaseDenom),
// 		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
// 		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
// 		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
// 		PruningStrategy: pruningtypes.PruningOptionNothing,
// 		CleanupDir:      true,
// 		SigningAlgo:     string(hd.Secp256k1Type),
// 		KeyringOptions:  []keyring.Option{},
// 	}
// }
