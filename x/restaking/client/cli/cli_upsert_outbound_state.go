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

// UpsertOutboundStateCmd returns a CLI command for upserting outbound state.
func CmdUpsertOutboundState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upsert-outbound-state [chain-id] [outbound-status] [epoch-number]",
		Short: "upsert outbound state",
		Long: `Upsert outbound state for a specific chain.
		
	Parameters:
	  chain-id: Chain identifier (uint64)
	  outbound-status: Status of outbound messages (0: Disabled, 1: Enabled)
	  epoch-number: Current epoch number (uint64)
	
	Example:
	  pellcored tx xmsg upsert-outbound-state 1 1 100 --from mykey`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse chain ID
			chainID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain ID: %w", err)
			}

			// Parse outbound status
			outboundStatus, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid outbound status: %w", err)
			}

			// Parse epoch number
			epochNumber, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid epoch number: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			state := &restakingtypes.EpochOutboundState{
				ChainId:        chainID,
				OutboundStatus: restakingtypes.OutboundStatus(outboundStatus),
				EpochNumber:    epochNumber,
			}

			msg := &restakingtypes.MsgUpsertOutboundState{
				Signer:        clientCtx.GetFromAddress().String(),
				OutboundState: state,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
