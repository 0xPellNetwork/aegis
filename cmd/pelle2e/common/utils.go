package common

import (
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/app"
	"github.com/0xPellNetwork/aegis/e2e/config"
)

const (
	FlagConfigFile = "config"
)

// GetConfig returns config from file from the command line flag
func GetConfig(cmd *cobra.Command) (*config.PellConfig, error) {
	configFile, err := cmd.Flags().GetString(FlagConfigFile)
	if err != nil {
		return nil, err
	}

	// use default config if no config file is specified
	if configFile == "" {
		return config.Default(), nil
	}

	configFile, err = filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	return config.LoadConfig(configFile)
}

// setCosmosConfig set account prefix to pell
func SetCosmosConfig() {
	cosmosConf := sdk.GetConfig()
	cosmosConf.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cosmosConf.Seal()
}
