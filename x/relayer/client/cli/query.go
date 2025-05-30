package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
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
		CmdQueryParams(),
		CmdBallotByIdentifier(),
		CmdObserverSet(),
		CmdGetSupportedChains(),
		CmdGetChainParamsForChain(),
		CmdGetChainParams(),
		CmdListNodeAccount(),
		CmdShowNodeAccount(),
		CmdShowCrosschainFlags(),
		CmdShowKeygen(),
		CmdShowObserverCount(),
		CmdBlameByIdentifier(),
		CmdGetAllBlameRecords(),
		CmdGetBlameByChainAndNonce(),
		CmdGetTssAddress(),
		CmdListTssHistory(),
		CmdShowTSS(),
		CmdGetTssAddressByFinalizedPellHeight(),
		CmdListChainNonces(),
		CmdShowChainNonces(),
		CmdListPendingNonces(),
	)

	return cmd
}
