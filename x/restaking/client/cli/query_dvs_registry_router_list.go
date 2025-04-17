package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

func CmdDVSRegistryRouterList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dvs-registry-router-list",
		Short: "Query registry router list",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.QueryDVSRegistryRouterList(cmd.Context(),
				&types.QueryDVSRegistryRouterListRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
