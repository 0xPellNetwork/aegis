package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// restrictedAddressBook is a map of restricted addresses
var restrictedAddressBook = map[string]bool{}

const (
	filename string = "pellclient_config.json"
	folder   string = "config"

	DefaultMaxMsgLen uint8 = 13
)

// Save saves PellClient config
func Save(config *Config, path string) error {
	folderPath := filepath.Join(path, folder)
	err := os.MkdirAll(folderPath, 0750)
	if err != nil {
		return err
	}
	file := filepath.Join(path, folder, filename)
	file = filepath.Clean(file)

	jsonFile, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(file, jsonFile, 0600)
	if err != nil {
		return err
	}
	return nil
}

// Load loads PellClient config from a filepath
func Load(path string) (Config, error) {
	// retrieve file
	file := filepath.Join(path, folder, filename)
	file, err := filepath.Abs(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to get absolute path: %w", err)
	}
	file = filepath.Clean(file)

	// read config
	cfg := NewConfig()
	input, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file %s: %w", file, err)
	}
	err = json.Unmarshal(input, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// read keyring backend and use test by default
	if cfg.KeyringBackend == KeyringBackendUndefined {
		cfg.KeyringBackend = KeyringBackendTest
	}

	if cfg.KeyringBackend != KeyringBackendFile && cfg.KeyringBackend != KeyringBackendTest {
		return Config{}, fmt.Errorf("invalid keyring backend %s", cfg.KeyringBackend)
	}

	if cfg.PellTxMsgLength == 0 {
		cfg.PellTxMsgLength = DefaultMaxMsgLen
	}

	// fields sanitization
	cfg.TssPath = GetPath(cfg.TssPath)
	cfg.PreParamsPath = GetPath(cfg.PreParamsPath)
	cfg.PellCoreHome = path

	// load compliance config
	LoadComplianceConfig(cfg)

	return cfg, nil
}

func LoadComplianceConfig(cfg Config) {
	restrictedAddressBook = cfg.GetRestrictedAddressBook()
}

func GetPath(inputPath string) string {
	path := strings.Split(inputPath, "/")
	if len(path) > 0 {
		if path[0] == "~" {
			home, err := os.UserHomeDir()
			if err != nil {
				return ""
			}
			path[0] = home
			return filepath.Join(path...)
		}
	}
	return inputPath
}

// ContainRestrictedAddress returns true if any one of the addresses is restricted
// Note: the addrs can contains both ETH and BTC addresses
func ContainRestrictedAddress(addrs ...string) bool {
	for _, addr := range addrs {
		if addr != "" && restrictedAddressBook[strings.ToLower(addr)] {
			return true
		}
	}
	return false
}
