package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func CmdListPendingXmsgWithinRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list_pending_xmsg_within_rate_limit",
		Short: "list all pending Xmsg within rate limit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ListPendingXmsgWithinRateLimit(
				context.Background(), &types.QueryListPendingXmsgWithinRateLimitRequest{},
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
