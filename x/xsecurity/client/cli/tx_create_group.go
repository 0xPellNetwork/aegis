package cli

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	xsecuritytypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

func CmdCreateGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-group [MaxOperatorCount] [KickBipsOfOperatorStake] [KickBipsOfTotalStake] [ChainId] [Pool] [Multiplier] [RateLimitWindow] [EjectableStakePercent] [MinStake]",
		Short: "Broadcast message CreateGroup",
		Args:  cobra.ExactArgs(9),
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

			chainID, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return err
			}

			multiplier, err := strconv.ParseInt(args[5], 10, 32)
			if err != nil {
				return err
			}

			rateLimitWindow, err := strconv.ParseInt(args[6], 10, 32)
			if err != nil {
				return err
			}

			ejectableStakePercent, err := strconv.ParseInt(args[7], 10, 32)
			if err != nil {
				return err
			}

			minStake, err := strconv.ParseInt(args[8], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateGroup(
				clientCtx.GetFromAddress().String(),
				&xsecuritytypes.OperatorSetParam{
					MaxOperatorCount:        uint32(maxOperatorCount),
					KickBipsOfOperatorStake: uint32(kickBipsOfOperatorStake),
					KickBipsOfTotalStake:    uint32(kickBipsOfTotalStake),
				},
				[]*xsecuritytypes.PoolParams{
					{
						ChainId:    uint64(chainID),
						Pool:       args[4],
						Multiplier: uint64(multiplier),
					},
				},
				&xsecuritytypes.GroupEjectionParam{
					RateLimitWindow:       uint32(rateLimitWindow),
					EjectableStakePercent: uint32(ejectableStakePercent),
				},
				sdkmath.NewInt(minStake),
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
