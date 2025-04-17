package cli

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func CmdUpsertCrosschainFeeParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upsert-crosschain-fee-params [crosschain-fee-params.json]",
		Short: "Broadcast message upsertCrosschainFeeParams",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var crosschainFeeParam types.CrosschainFeeParam
			file, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			file = filepath.Clean(file)
			input, err := os.ReadFile(file) // #nosec G304
			if err != nil {
				return err
			}

			if err = json.Unmarshal(input, &crosschainFeeParam); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpsertCrosschainFeeParams(
				clientCtx.GetFromAddress().String(),
				[]*types.CrosschainFeeParam{&crosschainFeeParam},
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
