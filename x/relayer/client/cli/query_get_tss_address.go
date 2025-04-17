package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

func CmdGetTssAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-tss-address [bitcoinChainId]]",
		Short: "Query current tss address",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			params := &types.QueryGetTssAddressRequest{}
			if len(args) == 1 {
				bitcoinChainID, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return err
				}
				params.BitcoinChainId = bitcoinChainID
			}

			res, err := queryClient.GetTssAddress(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetTssAddressByFinalizedPellHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-historical-tss-address [finalizedPellHeight] [bitcoinChainId]",
		Short: "Query tss address by finalized pell height (for historical tss addresses)",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			finalizedPellHeight, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			params := &types.QueryGetTssAddressByFinalizedHeightRequest{
				FinalizedPellHeight: finalizedPellHeight,
			}
			if len(args) == 2 {
				bitcoinChainID, err := strconv.ParseInt(args[1], 10, 64)
				if err != nil {
					return err
				}
				params.BitcoinChainId = bitcoinChainID
			}

			res, err := queryClient.GetTssAddressByFinalizedHeight(cmd.Context(), params)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
