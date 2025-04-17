package cli

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func CmdBlockProof() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-proof [chain-id] [height]",
		Short: "list block proof with chainid height",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			chainId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			height, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			params := &types.QueryBlockProofRequest{
				ChainId: chainId,
				Height:  uint64(height),
			}

			res, err := queryClient.BlockProof(
				context.Background(), params,
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
