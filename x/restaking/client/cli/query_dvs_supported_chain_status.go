package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

func CmdDVSSupportedChainStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dvs-supported-chain-status [registry-router-address] [chain-id]",
		Short: "Query DVS supported chain status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			registryRouterAddr := args[0]
			chainID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.QueryDVSSupportedChainStatus(cmd.Context(),
				&types.QueryDVSSupportedChainStatusRequest{
					RegistryRouterAddress: registryRouterAddr,
					ChainId:               chainID,
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
