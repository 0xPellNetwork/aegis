package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// TODO: remove it after next upgrade
// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdAddToWatchList(),
		CmdVoteGasPrice(),
		CmdXmsgOutboundVoter(),
		CmdXmsgInboundVoter(),
		CmdRemoveFromWatchList(),
		CmdUpdateTss(),
		CmdMigrateTssFunds(),
		CmdAddToInTxTracker(),
		CmdAbortStuckXmsg(),
		CmdAddAllowedXmsgSender(),
		CmdRemoveAllowedXmsgSender(),
		CmdInboundTxAdminMaintaince(),
		CmdUpsertCrosschainFeeParams(),
	)

	return cmd
}
