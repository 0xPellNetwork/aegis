package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

// TODO: remove this after the upgrade
func CmdDeleteBallot() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-ballot",
		Short: "Delete a ballot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgDeleteBallot{
				Signer:      clientCtx.GetFromAddress().String(),
				BallotIndex: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}
