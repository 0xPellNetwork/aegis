package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

func CmdDVSSupportedChainList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dvs-supported-chain-list [registry-router-address]",
		Short: "Query DVS supported chain list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			registryRouterAddr := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.QueryDVSSupportedChainList(cmd.Context(),
				&types.QueryDVSSupportedChainListRequest{
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
