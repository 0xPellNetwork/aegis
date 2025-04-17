package cli

import (
	"strconv"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/x/relayer/types"
)

func CmdUpdateObserver() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-observer [old-observer-address] [new-observer-address] [update-reason]",
		Short: "Broadcast message add-observer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			updateReasonInt, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return err
			}
			// #nosec G701 parsed in range
			updateReason, err := parseUpdateReason(int32(updateReasonInt))
			if err != nil {
				return err
			}
			msg := types.NewMsgUpdateObserver(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				updateReason,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parseUpdateReason(i int32) (types.RelayerUpdateReason, error) {
	if _, ok := types.RelayerUpdateReason_name[i]; ok {
		switch i {
		case 1:
			return types.RelayerUpdateReason_TOMBSTONED, nil
		case 2:
			return types.RelayerUpdateReason_ADMIN_UPDATE, nil
		}
	}
	return types.RelayerUpdateReason_TOMBSTONED, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid update reason")
}
