package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

func CmdQueryEpochOperatorShareSnapshot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-operator-share-snapshot [epoch-number]",
		Short: "Query epoch operator share snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			epochNumber, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch number: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetEpochChangedOperatorSharesSnapshot(cmd.Context(), &types.QueryEpochChangedOperatorSharesSnapshotRequest{EpochNumber: epochNumber})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
