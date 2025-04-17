package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

func CmdCreateRegistryRouter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-registry-router [chain-approver] [churn-approver] [ejector] [pauser] [unpauser] [initial-paused-status]",
		Short: "Broadcast message CreateRegistryRouter",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			initialPausedStatus, err := strconv.ParseInt(args[5], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRegistryRouter(
				clientCtx.GetFromAddress().String(),
				args[0], // chainApprover
				args[1], // churnApprover
				args[2], // ejector
				args[3], // pauser
				args[4], // unpauser
				initialPausedStatus,
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
