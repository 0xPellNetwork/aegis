package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"
)

func CmdUpdateBlocksPerEpoch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-blocks-per-epoch [blocks-per-epoch]",
		Short: "update blocks per epoch",
		Long: `update blocks per epoch.
		
		Parameters:
		  blocks-per-epoch: number of blocks per epoch (uint64)
		Example:
		  pellcored tx xmsg update-blocks-per-epoch 100 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			blocksPerEpoch, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid blocks per epoch: %w", err)
			}

			msg := &restakingtypes.MsgUpdateBlocksPerEpoch{
				Signer:         clientCtx.GetFromAddress().String(),
				BlocksPerEpoch: blocksPerEpoch,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
