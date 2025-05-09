package testutils

import ethcommon "github.com/ethereum/go-ethereum/common"

const (
	// tss addresses
	TSSAddressEVMMainnet = "0x70e967acFcC17c3941E87562161406d41676FD83"
	TSSAddressBTCMainnet = "bc1qm24wp577nk8aacckv8np465z3dvmu7ry45el6y"
	TssPubkeyEVMMainnet  = "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc"

	TSSAddressEVMIgnite3 = "0x8531a5aB847ff5B22D855633C25ED1DA3255247e"
	TSSAddressBTCIgnite3 = "tb1qy9pqmk2pd9sv63g27jt8r657wy0d9ueeh0nqur"
	TssPubkeyEVMIgnite3  = "pellpub1addwnpepq28c57cvcs0a2htsem5zxr6qnlvq9mzhmm76z3jncsnzz32rclangr2g35p"

	// some other addresses
	OtherAddress1 = "0x21248Decd0B7EcB0F30186297766b8AB6496265b"
	OtherAddress2 = "0x33A351C90aF486AebC35042Bb0544123cAed26AB"
	OtherAddress3 = "0x86B77E4fBd07CFdCc486cAe4F2787fB5C5a62cd3"

	// evm event names for test data naming
	EventPellSent = "PellSent"
)

// ConnectorManagerAddresses contains constants ERC20 connector addresses for testing
var ConnectorAddresses = map[int64]ethcommon.Address{
	// mainnet
	1:  ethcommon.HexToAddress("0x000007Cf399229b2f5A4D043F20E90C9C98B7C6a"),
	56: ethcommon.HexToAddress("0x000063A6e758D9e2f438d430108377564cf4077D"),

	// testnet
	5:        ethcommon.HexToAddress("0x00005E3125aBA53C5652f9F0CE1a4Cf91D8B15eA"),
	97:       ethcommon.HexToAddress("0x0000ecb8cdd25a18F12DAA23f6422e07fBf8B9E1"),
	11155111: ethcommon.HexToAddress("0x3963341dad121c9CD33046089395D66eBF20Fb03"),

	// localnet
	1337: ethcommon.HexToAddress("0xD28D6A0b8189305551a0A8bd247a6ECa9CE781Ca"),
}

// StrategyManagerAddresses contains constants ERC20 connector addresses for testing
var StrategyManagerAddresses = map[int64]ethcommon.Address{
	// mainnet
	1:  ethcommon.HexToAddress("0x000007Cf399229b2f5A4D043F20E90C9C98B7C6a"),
	56: ethcommon.HexToAddress("0x000063A6e758D9e2f438d430108377564cf4077D"),

	// testnet
	5:        ethcommon.HexToAddress("0x00005E3125aBA53C5652f9F0CE1a4Cf91D8B15eA"),
	97:       ethcommon.HexToAddress("0x0000ecb8cdd25a18F12DAA23f6422e07fBf8B9E1"),
	11155111: ethcommon.HexToAddress("0x3963341dad121c9CD33046089395D66eBF20Fb03"),

	// localnet
	1337: ethcommon.HexToAddress("0xD28D6A0b8189305551a0A8bd247a6ECa9CE781Ca"),
}

// DelegationManagerAddresses contains constants ERC20 custody addresses for testing
var DelegationManagerAddresses = map[int64]ethcommon.Address{
	// mainnet
	1:  ethcommon.HexToAddress("0x0000030Ec64DF25301d8414eE5a29588C4B0dE10"),
	56: ethcommon.HexToAddress("0x00000fF8fA992424957F97688015814e707A0115"),

	// testnet
	5:        ethcommon.HexToAddress("0x000047f11C6E42293F433C82473532E869Ce4Ec5"),
	97:       ethcommon.HexToAddress("0x0000a7Db254145767262C6A81a7eE1650684258e"),
	11155111: ethcommon.HexToAddress("0x84725b70a239d3Faa7C6EF0C6C8E8b6c8e28338b"),
}
