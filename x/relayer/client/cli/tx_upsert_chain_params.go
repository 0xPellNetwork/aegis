package cli

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

func CmdUpsertChainParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upsert-chain-params [client-params.json]",
		Short: "Broadcast message upsertChainParams",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var clientParams types.ChainParams
			file, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			file = filepath.Clean(file)
			input, err := os.ReadFile(file) // #nosec G304
			if err != nil {
				return err
			}

			if err = json.Unmarshal(input, &clientParams); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpsertChainParams(
				clientCtx.GetFromAddress().String(),
				&clientParams,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
