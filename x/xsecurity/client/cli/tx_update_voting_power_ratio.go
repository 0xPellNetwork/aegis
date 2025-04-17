package cli

import (
	"strconv"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

func CmdUpdateVotingPowerRatio() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-voting-power-ratio [voting-power-ratio]",
		Short: "Broadcast message UpdateVotingPowerRatio",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			numerator, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			denominator, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateVotingPowerRatio(
				clientCtx.GetFromAddress().String(),
				math.NewInt(numerator),
				math.NewInt(denominator),
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
