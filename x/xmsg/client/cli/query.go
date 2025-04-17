package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group xmsg queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdListOutTxTracker(),
		CmdShowOutTxTracker(),
		CmdListGasPrice(),
		CmdShowGasPrice(),

		CmdListSend(),
		CmdShowSend(),
		CmdLastPellHeight(),
		CmdInTxHashToXmsgData(),
		CmdListInTxHashToXmsg(),
		CmdShowInTxHashToXmsg(),

		CmdPendingXmsg(),
		CmdListInTxTrackerByChain(),
		CmdListInTxTrackers(),
		CmdListPendingXmsgWithinRateLimit(),

		CmdShowUpdateRateLimiterFlags(),

		CmdListAllowedXmsgSenders(),
		CmdBlockProof(),

		CmdGetCrosschainFeeParam(),
		CmdListCrosschainFeeParams(),
	)

	return cmd
}
