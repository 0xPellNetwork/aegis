package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// CmdListAllowedXmsgSenders creates a CLI command to list all xmsg builders.
// This command queries the xmsg module for all registered xmsg builders
// and returns the result as a formatted proto message.
func CmdListAllowedXmsgSenders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-allowed-xmsg-senders",
		Short: "list all xmsg builders",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ListAllowedXmsgSenders(cmd.Context(), &types.QueryListAllowedXmsgSendersRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
