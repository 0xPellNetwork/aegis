package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	xsecuritytypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

func CmdAddpools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-pools [ChainId] [Pool] [Multiplier]",
		Short: "Broadcast message AddPools",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chainID, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			multiplier, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddPools(
				clientCtx.GetFromAddress().String(),
				[]*xsecuritytypes.PoolParams{
					{
						ChainId:    uint64(chainID),
						Pool:       args[1],
						Multiplier: uint64(multiplier),
					},
				},
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
