package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

func CmdDVSGroupDataList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dvs-group-data-list [registry-router-address]",
		Short: "Query DVS group data list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			registryRouterAddr := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.QueryDVSGroupDataList(cmd.Context(),
				&types.QueryDVSGroupDataListRequest{
					RegistryRouterAddress: registryRouterAddr,
				},
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
