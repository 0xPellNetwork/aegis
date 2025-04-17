package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

func CmdUpdateLSTStakingEnabled() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-lst-staking-enabled [enabled]",
		Short: "Broadcast message UpdataLSTStakingEnabled",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			parseBool, err := strconv.ParseBool(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateLSTStakingEnabled(
				clientCtx.GetFromAddress().String(),
				parseBool,
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
