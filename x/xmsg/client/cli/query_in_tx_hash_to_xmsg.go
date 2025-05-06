package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func CmdListInTxHashToXmsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-in-tx-hash-to-xmsg",
		Short: "list all inTxHashToXmsg",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllInTxHashToXmsgRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.InTxHashToXmsgAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowInTxHashToXmsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-in-tx-hash-to-xmsg [in-tx-hash]",
		Short: "shows a inTxHashToXmsg",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argInTxHash := args[0]

			params := &types.QueryGetInTxHashToXmsgRequest{
				InTxHash: argInTxHash,
			}

			res, err := queryClient.InTxHashToXmsg(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdInTxHashToXmsgData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "in-tx-hash-to-xmsg-data [in-tx-hash]",
		Short: "query a xmsg data from a in tx hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argInTxHash := args[0]

			params := &types.QueryInTxHashToXmsgDataRequest{
				InTxHash: argInTxHash,
			}

			res, err := queryClient.InTxHashToXmsgData(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
