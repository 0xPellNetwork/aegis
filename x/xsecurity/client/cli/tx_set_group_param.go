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

func CmdSetGroupParam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-group-param [MaxOperatorCount] [KickBipsOfOperatorStake] [KickBipsOfTotalStake]",
		Short: "Broadcast message SetGroupParam",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			maxOperatorCount, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			kickBipsOfOperatorStake, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return err
			}

			kickBipsOfTotalStake, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgSetGroupParam(
				clientCtx.GetFromAddress().String(),
				&xsecuritytypes.OperatorSetParam{
					MaxOperatorCount:        uint32(maxOperatorCount),
					KickBipsOfOperatorStake: uint32(kickBipsOfOperatorStake),
					KickBipsOfTotalStake:    uint32(kickBipsOfTotalStake),
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
