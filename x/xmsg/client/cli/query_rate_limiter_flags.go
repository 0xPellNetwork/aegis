package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func CmdShowUpdateRateLimiterFlags() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update_rate_limit_flags",
		Short: "shows the rate limiter flags",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RateLimiterFlags(context.Background(), &types.QueryRateLimiterFlagsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
