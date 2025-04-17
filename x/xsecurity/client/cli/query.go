package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group observer queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryRegistryRouterAddress(),
		CmdQueryGroupInfo(),
		CmdOperatorRegistrationList(),
		CmdOperatorShares(),
		CmdQueryVotingPowerRatio(),
		CmdQueryLSTStakingEnabled(),
	)

	return cmd
}
