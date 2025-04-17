package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

func CmdDVSGroupQueryStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dvs-group-sync-status [tx-hash]",
		Short: "Query DVS group sync status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txHash := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.QueryDVSGroupSyncStatus(cmd.Context(), &types.QueryDVSGroupSyncStatusRequest{
				TxHash: txHash,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
