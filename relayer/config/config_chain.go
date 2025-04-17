package config

import (
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
)

const (
	BtcConfirmationCount    = 1
	DevEthConfirmationCount = 2

	// TssTestPrivkey is the private key of the TSS address
	// #nosec G101 - used for testing only
	TssTestPrivkey = "2082bc9775d6ee5a05ef221a9d1c00b3cc3ecb274a4317acc0a182bc1e05d1bb"
	TssTestAddress = "0xE80B6467863EbF8865092544f441da8fD3cF6074"

	// Number of blocks to scan per period under normal conditions
	DefaultBlocksPerPeriod = 100

	// MaxLatestIndexedBlockGap defines the maximum allowed gap between the current block
	// and the latest indexed block height on chain. The scanning process will stop when
	// this threshold is exceeded to prevent invalid votes.
	DefaultMaxLatestIndexedBlockGap = uint64(2000)
)

func GetPellConnectorABI() string {
	return pellconnector.PellConnectorABI
}

// New constructs Config optionally with default values.
func New() Config {
	return Config{
		EVMChainConfigs: evmChainsConfigs,
		BitcoinConfig:   bitcoinConfigRegnet,
	}
}

var bitcoinConfigRegnet = BTCConfig{
	RPCUsername: "smoketest", // smoketest is the previous name for E2E test, we keep this name for compatibility between client versions in upgrade test
	RPCPassword: "123",
	RPCHost:     "bitcoin:18443",
	RPCParams:   "regtest",
}

var evmChainsConfigs = map[int64]EVMConfig{}
