package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// CmdTxAddAllowedXmsgSenders creates a CLI command to add allowed xmsg senders.
// This command adds a list of allowed xmsg senders to the xmsg module.
func CmdAddAllowedXmsgSender() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-allowed-xmsg-sender [builder]...",
		Short: "add xmsg builders",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddAllowedXmsgSender(clientCtx.GetFromAddress().String(), args)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdRemoveAllowedXmsgSender creates a CLI command to remove allowed xmsg sender.
func CmdRemoveAllowedXmsgSender() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-allowed-xmsg-sender [builder]...",
		Short: "remove xmsg builders",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRemoveAllowedXmsgSender(clientCtx.GetFromAddress().String(), args)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
