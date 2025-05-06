package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	FlagTxType = "tx-type"
)

const (
	TxStakerDeposited = "staker-deposited"
	TxStakerDelegated = "staker-delegated"
)

func CmdListSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-xmsg",
		Short: "list all Xmsg",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllXmsgRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.XmsgAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdPendingXmsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-pending-xmsg [chain-id] [limit]",
		Short: "shows pending Xmsg",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)
			chainID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			limit, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			params := &types.QueryListPendingXmsgRequest{
				ChainId: chainID,
				// #nosec G701 bit size verified
				Limit: uint32(limit),
			}

			res, err := queryClient.ListPendingXmsg(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-xmsg [index]",
		Short: "shows a Xmsg",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetXmsgRequest{
				Index: args[0],
			}

			res, err := queryClient.Xmsg(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Transaction CLI /////////////////////////
//pellcored tx pellcore xmsg-voter 0x96B05C238b99768F349135de0653b687f9c13fEE ETH 0x96B05C238b99768F349135de0653b687f9c13fEE ETH 1000000000000000000 0 message hash 100 --from=pell --keyring-backend=test --yes --chain-id=localnet_101-1

func CmdXmsgInboundVoter() *cobra.Command {
	cmd := &cobra.Command{
		Use: "inbound-voter [sender] [senderChainID] [txOrigin] [receiver] [receiverChainID] " +
			"[inTxHash] [inBlockHeight] [eventIndex] [pellData]",
		Short: "Broadcast message sendVoter",
		Long: `Broadcast message sendVoter. when tx-type is staker-deposited, 
			the pellData include "staker token strategy amount". when tx-type is 
			staker-delegated, the pellData include "staker operator"`,
		Args: cobra.RangeArgs(10, 12),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsSender := args[0]
			argsSenderChain, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			argsTxOrigin := args[2]
			argsReceiver := args[3]
			argsReceiverChain, err := strconv.ParseInt(args[4], 10, 64)
			if err != nil {
				return err
			}

			argsInTxHash := args[5]

			argsInBlockHeight, err := strconv.ParseUint(args[6], 10, 64)
			if err != nil {
				return err
			}

			// parse argsp[11] to uint type and not uint64
			argsEventIndex, err := strconv.ParseUint(args[7], 10, 32)
			if err != nil {
				return err
			}

			pellData, err := getPellData(cmd, args)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgVoteOnObservedInboundTx(
				clientCtx.GetFromAddress().String(),
				argsSender,
				argsSenderChain,
				argsTxOrigin,
				argsReceiver,
				argsReceiverChain,
				argsInTxHash,
				argsInBlockHeight,
				250_000,
				uint(argsEventIndex),
				pellData,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagTxType, "staker-deposited", fmt.Sprintf("[%s|%s] pell data in inbound tx.", TxStakerDeposited, TxStakerDelegated))
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdXmsgOutboundVoter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outbound-voter [sendHash] [outTxHash] [outBlockHeight] [outGasUsed] [outEffectiveGasPrice] [outEffectiveGasLimit] [Status] [chain] [outTXNonce] [failedReasonMsg]",
		Short: "Broadcast message receiveConfirmation",
		Args:  cobra.ExactArgs(10),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsSendHash := args[0]
			argsOutTxHash := args[1]

			argsOutBlockHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			argsOutGasUsed, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			argsOutEffectiveGasPrice, ok := math.NewIntFromString(args[4])
			if !ok {
				return errors.New("invalid effective gas price, enter 0 if unused")
			}

			argsOutEffectiveGasLimit, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return err
			}

			status, err := chains.ReceiveStatusFromString(args[6])
			if err != nil {
				return err
			}

			chain, err := strconv.ParseInt(args[7], 10, 64)
			if err != nil {
				return err
			}

			outTxNonce, err := strconv.ParseUint(args[8], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgVoteOnObservedOutboundTx(
				clientCtx.GetFromAddress().String(),
				argsSendHash,
				argsOutTxHash,
				argsOutBlockHeight,
				argsOutGasUsed,
				argsOutEffectiveGasPrice,
				argsOutEffectiveGasLimit,
				status,
				args[9],
				chain,
				outTxNonce,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdAbortStuckXmsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort-stuck-xmsg [index]",
		Short: "abort a stuck Xmsg",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			msg := &types.MsgAbortStuckXmsg{
				Signer:    clientCtx.GetFromAddress().String(),
				XmsgIndex: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getPellData(cmd *cobra.Command, args []string) (types.InboundPellEvent, error) {
	const argsBaseParamsLen = 8
	const argsStakerDepositedParamsLen = 4
	const argsStakerDelegatedParamsLen = 2

	txType, err := cmd.Flags().GetString(FlagTxType)
	if err != nil {
		return types.InboundPellEvent{}, err
	}

	var pellData types.InboundPellEvent
	switch txType {
	case TxStakerDeposited:
		if len(args) != (argsBaseParamsLen + argsStakerDepositedParamsLen) {
			return types.InboundPellEvent{}, fmt.Errorf("the number of %s parameters is incorrect", TxStakerDeposited)
		}

		pellData = types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDeposited{
				StakerDeposited: &types.StakerDeposited{
					Staker:   args[8],
					Token:    args[9],
					Strategy: args[10],
					Shares:   math.NewUintFromString(args[11]),
				},
			},
		}
	case TxStakerDelegated:
		if len(args) != (argsBaseParamsLen + argsStakerDelegatedParamsLen) {
			return types.InboundPellEvent{}, fmt.Errorf("the number of %s parameters is incorrect", TxStakerDelegated)
		}
		pellData = types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDelegated{
				StakerDelegated: &types.StakerDelegated{
					Staker:   args[8],
					Operator: args[9],
				},
			},
		}
	default:
		return types.InboundPellEvent{}, fmt.Errorf("don't known value of %s", FlagTxType)
	}

	return pellData, nil
}
