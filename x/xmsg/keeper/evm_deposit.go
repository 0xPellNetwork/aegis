package keeper

import (
	"fmt"
	"math/big"
	"reflect"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/pkg/errors"

	"github.com/pell-chain/pellcore/pkg/chains"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// HandleEVMEvents handles all events from an inbound tx
// returns (isContractReverted, err)
// (true, non-nil) means CallEVM() reverted
func (k Keeper) HandleEVMEvents(ctx sdk.Context, xmsg *types.Xmsg) (bool, error) {
	inboundSender := xmsg.GetInboundTxParams().Sender
	inboundSenderChainID := xmsg.GetInboundTxParams().SenderChainId

	if len(ctx.TxBytes()) > 0 {
		// add event for tendermint transaction hash format
		hash := tmbytes.HexBytes(tmtypes.Tx(ctx.TxBytes()).Hash())
		ethTxHash := ethcommon.BytesToHash(hash)
		xmsg.GetCurrentOutTxParam().OutboundTxHash = ethTxHash.String()
		// #nosec G701 always positive
		xmsg.GetCurrentOutTxParam().OutboundTxExternalHeight = uint64(ctx.BlockHeight())
	}

	from, err := chains.DecodeAddressFromChainID(inboundSenderChainID, inboundSender)
	if err != nil {
		return false, fmt.Errorf("HandleEVMEvents: unable to decode address: %s", err.Error())
	}

	var to ethcommon.Address
	var evmTxResponse *evmtypes.MsgEthereumTxResponse
	var action string
	contractCall := false

	k.Logger(ctx).Info("HandleEVMEvents: inbound events",
		"type", reflect.TypeOf(xmsg.InboundTxParams.InboundPellTx.GetPellData()),
		"xmsgIndex", xmsg.Index,
		"txHash", xmsg.InboundTxParams.InboundTxHash,
	)

	switch xmsg.InboundTxParams.InboundPellTx.GetPellData().(type) {
	case *types.InboundPellEvent_StakerDeposited:
		to, err = k.pevmKeeper.GetPellStrategyManagerProxyContractAddress(ctx)
		if err != nil {
			return false, errors.Wrap(types.ErrReceiverIsEmpty, err.Error())
		}
		stakerDeposited := xmsg.InboundTxParams.InboundPellTx.GetStakerDeposited()
		action = stakerDeposited.String()
		evmTxResponse, contractCall, err = k.pevmKeeper.CallSyncDepositStateOnPellStrategyManager(
			ctx,
			from,
			inboundSenderChainID,
			ethcommon.HexToAddress(stakerDeposited.Staker),
			ethcommon.HexToAddress(stakerDeposited.Strategy),
			stakerDeposited.Shares.BigInt(),
		)
	case *types.InboundPellEvent_StakerDelegated:
		to, err = k.pevmKeeper.GetPellDelegationManagerProxyContractAddress(ctx)
		if err != nil {
			return false, errors.Wrap(types.ErrReceiverIsEmpty, err.Error())
		}
		stakerDelegated := xmsg.InboundTxParams.InboundPellTx.GetStakerDelegated()
		action = stakerDelegated.String()
		evmTxResponse, contractCall, err = k.pevmKeeper.CallSyncDelegatedStateOnPellDelegationManager(
			ctx,
			from,
			inboundSenderChainID,
			ethcommon.HexToAddress(stakerDelegated.Staker),
			ethcommon.HexToAddress(stakerDelegated.Operator),
		)
	case *types.InboundPellEvent_WithdrawalQueued:
		to, err = k.pevmKeeper.GetPellDelegationManagerProxyContractAddress(ctx)
		if err != nil {
			return false, errors.Wrap(types.ErrReceiverIsEmpty, err.Error())
		}
		withdrawalQueued := xmsg.InboundTxParams.InboundPellTx.GetWithdrawalQueued()
		action = withdrawalQueued.String()
		if err != nil {
			return false, errors.Wrap(types.ErrCannotProcessWithdrawal, err.Error())
		}
		evmTxResponse, contractCall, err = k.pevmKeeper.CallSyncWithdrawalStateOnPellDelegationManager(
			ctx,
			inboundSenderChainID,
			ethcommon.HexToAddress(withdrawalQueued.Withdrawal.Staker),
			withdrawalQueued,
		)
	case *types.InboundPellEvent_StakerUndelegated:
		to, err = k.pevmKeeper.GetPellDelegationManagerProxyContractAddress(ctx)
		if err != nil {
			return false, errors.Wrap(types.ErrReceiverIsEmpty, err.Error())
		}
		stakerUndelegated := xmsg.InboundTxParams.InboundPellTx.GetStakerUndelegated()
		action = stakerUndelegated.String()
		if err != nil {
			return false, errors.Wrap(types.ErrCannotProcessWithdrawal, err.Error())
		}
		evmTxResponse, contractCall, err = k.pevmKeeper.CallSyncUndelegateStateOnPellDelegationManager(
			ctx,
			inboundSenderChainID,
			ethcommon.HexToAddress(stakerUndelegated.Staker),
		)
	case *types.InboundPellEvent_RegisterChainDvsToPell:
		ctx.Logger().Debug("HandleEVMEvents: RegisterChainDvsToPell")
		action := xmsg.InboundTxParams.InboundPellTx.GetRegisterChainDvsToPell()
		evmTxResponse, contractCall, err = k.pevmKeeper.CallAddSupportedChainOnRegistryRouter(ctx, action)
	case *types.InboundPellEvent_PellSent:
		ctx.Logger().Debug("HandleEVMEvents: PellSent")
		action := xmsg.InboundTxParams.InboundPellTx.GetPellSent()
		evmTxResponse, contractCall, err = k.pevmKeeper.CallProcessPellSent(ctx, action, xmsg.Index)
	default:
		return false, fmt.Errorf("unknown pell event in xmsg[%s]", xmsg.Index)
	}

	if err != nil {
		k.Logger(ctx).Error("HandleEVMEvents: failed to call contract", "error", err)
		return pevmtypes.IsContractReverted(evmTxResponse, err) || errShouldRevertXmsg(err), err
	}

	k.Logger(ctx).Info("HandleEVMEvents: contract call success", "action", action, "response", evmTxResponse.String())

	// non-empty msg.Message means this is a contract call; therefore the logs should be processed.
	if !evmTxResponse.Failed() && contractCall {
		logs := evmtypes.LogsToEthereum(evmTxResponse.Logs)
		if len(logs) > 0 {
			ctx = ctx.WithValue("inXmsgIndices", xmsg.Index)
			txOrigin := xmsg.InboundTxParams.TxOrigin
			if txOrigin == "" {
				txOrigin = inboundSender
			}

			if err := k.processInternalEventLogs(ctx, logs, txOrigin, to); err != nil {
				return false, err
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(sdk.EventTypeMessage,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute("action", action),
					sdk.NewAttribute("contract", to.String()),
					sdk.NewAttribute("xmsgIndex", xmsg.Index),
				),
			)
		}
	}
	return false, nil
}

// processInternalEventLogs processes the internal event logs
func (k Keeper) processInternalEventLogs(ctx sdk.Context, logs []*ethtypes.Log, txOrigin string, to ethcommon.Address) error {
	// TODO: refactor this to reuse evm hooks structure
	internalMsg := buildEVMMsgByInternalLogs(txOrigin, to)
	if err := k.Hooks().PostTxProcessing(ctx, internalMsg, &ethtypes.Receipt{Logs: logs}); err != nil {
		return err
	}

	// internal event hooks for other modules
	for _, hook := range k.internalEventHooks {
		if err := hook.HandleEventLogs(ctx, to, logs, txOrigin); err != nil {
			return err
		}
	}

	return nil
}

// buildEVMMsgByInternalLogs constructs an internal EVM message based on system logs to reuse EVM hooks structure.
// It creates a simulated transaction message with minimal required parameters.
// Parameters:
//   - txOrigin: the original transaction sender address
//   - to: the destination contract address
func buildEVMMsgByInternalLogs(txOrigin string, to ethcommon.Address) coretypes.Message {
	return coretypes.NewMessage(common.HexToAddress(txOrigin), &to, 0, nil, 0, nil, big.NewInt(0), big.NewInt(0), []byte{}, ethtypes.AccessList{}, false)
}

// errShouldRevertXmsg returns true if the xmsg should revert from the error of the deposit
// we revert the xmsg if a non-contract is tried to be called, if the liquidity cap is reached, or if the zrc20 is paused
func errShouldRevertXmsg(err error) bool {
	return errors.Is(err, pevmtypes.ErrCallNonContract)
}
