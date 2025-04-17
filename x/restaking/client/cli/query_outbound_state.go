package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

func CmdQueryOutboundStateByChainID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outbound-state [chain-id]",
		Short: "Query outbound state by chain id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chainID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetOutboundStateByChainID(cmd.Context(), &types.QueryOutboundStateByChainIDRequest{ChainId: chainID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
