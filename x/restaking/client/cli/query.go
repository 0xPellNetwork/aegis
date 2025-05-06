package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group pevm queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryEpochInfo())
	cmd.AddCommand(CmdQueryEpochOperatorShareSnapshot())
	cmd.AddCommand(CmdQueryOutboundStateByChainID())
	cmd.AddCommand(CmdDVSGroupQueryStatus())
	cmd.AddCommand(CmdDVSSupportedChainList())
	cmd.AddCommand(CmdDVSSupportedChainStatus())
	cmd.AddCommand(CmdDVSRegistryRouterList())
	cmd.AddCommand(CmdDVSGroupDataList())
	cmd.AddCommand(CmdDVSGroupOperatorRegistrationList())

	return cmd
}
