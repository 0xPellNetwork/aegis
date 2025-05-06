package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// CmdGetCrosschainFeeParam returns the crosschain fee param for a given chain id
func CmdGetCrosschainFeeParam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-crosschain-fee-param [chain-id]",
		Short: "Query GetCrosschainFeeParam",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqChainID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryCrosschainFeeParamByChainIdRequest{
				ChainId: reqChainID,
			}
			res, err := queryClient.CrosschainFeeParamByChainId(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdGetCrosschainFeeParams returns all crosschain fee params
func CmdListCrosschainFeeParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-crosschain-fee-params",
		Short: "Query ListCrosschainFeeParams",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryCrosschainFeeParamsRequest{}
			res, err := queryClient.CrosschainFeeParams(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
