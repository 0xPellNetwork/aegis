package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ types.EventHandler = &ConnectorEventHandler{}

type ConnectorEventHandler struct {
	k Keeper
}

// HandleEvent handles the event from the log. It returns the crosschain fees for the event
func (h ConnectorEventHandler) HandleEvent(ctx sdk.Context, _ uint64, toAddress ethcommon.Address, logs []*ethtypes.Log, txOrigin string) ([]*types.CrossChainFee, error) {
	ctx.Logger().Info("ConnectorEventHandler: HandleEvent started", "toAddress", toAddress.Hex(), "logs", logs, "txOrigin", txOrigin)

	addr, err := h.GetContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	fees := make([]*types.CrossChainFee, 0)

	for _, log := range logs {
		event, err := h.ParseEvent(addr, log)
		if err != nil {
			continue
		}

		switch pellSentEvent := event.(type) {
		case *pellconnector.PellConnectorPellSent:
			h.k.Logger(ctx).Info("ConnectorEventHandler: pellSentEvent detected", "event", pellSentEvent)
			// These cannot be processed without TSS keys, return an error if TSS is not found
			tss, found := h.k.relayerKeeper.GetTSS(ctx)
			if !found {
				return nil, errorsmod.Wrap(types.ErrCannotFindTSSKeys, "Cannot process logs without TSS keys")
			}

			// Do not process withdrawal events if inbound is disabled
			if !h.k.relayerKeeper.IsInboundEnabled(ctx) {
				return nil, observertypes.ErrInboundDisabled
			}

			// check if the sender is a xmsg builder
			if !h.k.IsAllowedXmsgSender(ctx, pellSentEvent.PellTxSenderAddress.Hex()) {
				ctx.Logger().Warn(
					"ConnectorEventHandler: pellSentEvent.PellTxSenderAddress: is not a allowed xmsg sender",
					"sender", pellSentEvent.PellTxSenderAddress.Hex(),
				)
				continue
			}
			// If the event is valid, we will process it and create a new Xmsg
			// If the process fails, we will return an error and roll back the transaction
			if err := h.k.ProcessPellSentEvent(ctx, pellSentEvent, toAddress, txOrigin, tss); err != nil {
				return nil, err
			}

			if feeParam, exist := h.k.GetCrosschainEventFee(ctx, pellSentEvent.DestinationChainId.Int64()); exist && feeParam.IsSupported {
				fees = append(fees, &types.CrossChainFee{
					Address: sdk.AccAddress(pellSentEvent.SourceTxOriginAddress.Bytes()),
					Fee:     feeParam.PellSentEventFee,
				})
			}
		default:
			continue
		}
	}

	return fees, nil
}

// get connector contract address
func (h ConnectorEventHandler) GetContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	return h.k.pevmKeeper.GetPellConnectorContractAddress(ctx)
}

// parse event from log
func (h ConnectorEventHandler) ParseEvent(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error) {
	if len(log.Topics) == 0 {
		return nil, fmt.Errorf("ParsePellSentEvent: invalid log - no topics")
	}

	pellConnectorPEVM, err := pellconnector.NewPellConnectorFilterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	event, err := pellConnectorPEVM.ParsePellSent(*log)
	if err != nil {
		return nil, err
	}

	if event.Raw.Address != contractAddr {
		return nil, fmt.Errorf("ParsePellSentEvent: event address %s does not match connector %s",
			event.Raw.Address.Hex(), contractAddr.Hex())
	}
	return event, nil
}
