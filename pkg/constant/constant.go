package constant

import (
	"time"
)

const (
	// PellBlockTime is the block time of the PellChain network
	// It's a rough estimate that can be used in non-critical path to estimate the time of a block
	PellBlockTime = 6000 * time.Millisecond

	// DonationMessage is the message for donation transactions
	// Transaction sent to the TSS or ERC20 Custody address containing this message are considered as a donation
	DonationMessage = "I am rich!"

	// CmdWhitelistERC20 is used for Xmsg of type cmd to give the instruction to the TSS to whitelist an ERC20 on an exeternal chain
	CmdWhitelistERC20 = "cmd_whitelist_erc20"

	// CmdMigrateTssFunds is used for Xmsg of type cmd to give the instruction to the TSS to transfer its funds on a new address
	CmdMigrateTssFunds = "cmd_migrate_tss_funds"

	// BTCWithdrawalDustAmount is the minimum satoshis that can be withdrawn from pEVM to avoid outbound dust output
	BTCWithdrawalDustAmount = 1000
)
