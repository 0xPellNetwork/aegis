package config

import (
	"encoding/json"

	"github.com/spf13/afero"
)

var AppFs = afero.NewOsFs()

const (
	FlagConfig               = "config"
	defaultCfgFileName       = "pelltool_config.json"
	PellURL                  = "127.0.0.1:1317"
	BtcExplorerURL           = "https://blockstream.info/api/"
	EthRPCURL                = "https://ethereum-rpc.publicnode.com"
	StrategyManagerAddress   = "0x000007Cf399229b2f5A4D043F20E90C9C98B7C6a"
	DelegationManagerAddress = "0x0000030Ec64DF25301d8414eE5a29588C4B0dE10"
)

// Config is a struct the defines the configuration fields used by pelltool
type Config struct {
	PellURL                  string
	BtcExplorerURL           string
	EthRPCURL                string
	EtherscanAPIkey          string
	StrategyManagerAddress   string
	DelegationManagerAddress string
}

func DefaultConfig() *Config {
	return &Config{
		PellURL:                  PellURL,
		BtcExplorerURL:           BtcExplorerURL,
		EthRPCURL:                EthRPCURL,
		StrategyManagerAddress:   StrategyManagerAddress,
		DelegationManagerAddress: DelegationManagerAddress,
	}
}

func (c *Config) Save() error {
	file, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	err = afero.WriteFile(AppFs, defaultCfgFileName, file, 0600)
	return err
}

func (c *Config) Read(filename string) error {
	data, err := afero.ReadFile(AppFs, filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, c)
	return err
}

func GetConfig(filename string) (*Config, error) {
	//Check if cfgFile is empty, if so return default Config and save to file
	if filename == "" {
		cfg := DefaultConfig()
		err := cfg.Save()
		return cfg, err
	}

	//if file is specified, open file and return struct
	cfg := &Config{}
	err := cfg.Read(filename)
	return cfg, err
}
