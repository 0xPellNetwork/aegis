package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	xsecuritytypes "github.com/pell-chain/pellcore/x/restaking/types"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

func CmdRemovepools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-pools [ChainId] [Pool] [Multiplier]",
		Short: "Broadcast message RemovePools",
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

			msg := types.NewMsgRemovePools(
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
