package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// TODO: remove it after next upgrade
// CmdInboundTxAdminMaintaince is a command to maintain the inbound tx
func CmdInboundTxAdminMaintaince() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inbound_admin_maintaince",
		Short: "inbound admin maintaince",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Verify that we have a valid from address
			fromAddress := clientCtx.GetFromAddress()
			if fromAddress.Empty() {
				return fmt.Errorf("from address cannot be empty")
			}

			chainId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			fromBlockHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			toBlockHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgInboundTxMaintenance{
				Signer:          fromAddress.String(),
				ChainId:         chainId,
				FromBlockHeight: fromBlockHeight,
				ToBlockHeight:   toBlockHeight,
			}

			fmt.Println(*msg)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
