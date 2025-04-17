package keeper

import (
	"encoding/base64"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pell-chain/pellcore/pkg/utils"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	SyncOperatorRegisteredEvent      = "syncRegisterAsOperator"
	SyncOperatorDetailsModifiedEvent = "syncModifyOperatorDetails"
)

type eventBuilder func(chainParams *relayertypes.ChainParams, operator types.Operator) (*xmsgtypes.InboundPellEvent, error)

type OperatorHandler struct {
	Keeper
}

var _ xmsgtypes.EventHandler = OperatorHandler{}

func (h OperatorHandler) HandleEvent(ctx sdk.Context, epochNum uint64, _ ethcommon.Address, logs []*ethtypes.Log, txOrigin string) ([]*xmsgtypes.CrossChainFee, error) {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.Keeper.Logger(ctx).Info("restaking OperatorEventHandler", "txOrigin", txOrigin, "xmsgIndex", inXmsgIndices, "logs", logs)

	addr, err := h.GetContractAddress(ctx)
	if err != nil || addr == (ethcommon.Address{}) {
		return nil, err
	}

	crossChainFees := make([]*xmsgtypes.CrossChainFee, 0)

	for _, log := range logs {
		event, err := h.ParseEvent(addr, log)
		if err != nil {
			h.Logger(ctx).Error("operator handler failed to parse event",
				"error", err,
				"topics", log.Topics,
				"data", log.Data,
			)
			continue
		}
		h.Logger(ctx).Info("operator handler parsed event",
			"event", event,
			"xmsgIndex", inXmsgIndices,
		)

		var fee *sdkmath.Int
		var feeAddr ethcommon.Address
		switch evt := event.(type) {
		case *pelldelegationmanager.PellDelegationManagerOperatorRegistered:
			fee, err = h.handleRegisterAsOperatorEvent(ctx, epochNum, evt)
			if err != nil {
				h.Logger(ctx).Error("operator handler failed to handle register as operator event", "error", err)
			}
			feeAddr = evt.Operator
		case *pelldelegationmanager.PellDelegationManagerOperatorDetailsModified:
			fee, err = h.handleModifyOperatorDetailsEvent(ctx, epochNum, evt)
			if err != nil {
				h.Logger(ctx).Error("operator handler failed to handle modify operator details event", "error", err)
			}
			feeAddr = evt.Operator
		}

		if fee != nil && !fee.IsZero() {
			crossChainFees = append(crossChainFees, &xmsgtypes.CrossChainFee{
				Address: sdk.AccAddress(feeAddr.Bytes()),
				Fee:     *fee,
			})
		}
	}

	return crossChainFees, nil
}

// GetContractAddress returns the contract address
func (h OperatorHandler) GetContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	return h.pevmKeeper.GetPellDelegationManagerProxyContractAddress(ctx)
}

// EventParser defines the event type and its corresponding parsing function
type EventParser struct {
	EventType interface{}
	Parser    func(log ethtypes.Log) (interface{}, error)
}

// ParseEvent parses the event from the log
func (h OperatorHandler) ParseEvent(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error) {
	delegationManager, err := pelldelegationmanager.NewPellDelegationManagerFilterer(contractAddr, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	// operator registered event
	if operatorRegisteredEvent, err := delegationManager.ParseOperatorRegistered(*log); err == nil {
		if !strings.EqualFold(operatorRegisteredEvent.Raw.Address.Hex(), contractAddr.Hex()) {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				operatorRegisteredEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return operatorRegisteredEvent, nil
	}

	// operator details modified event
	if operatorDetailsModifiedEvent, err := delegationManager.ParseOperatorDetailsModified(*log); err == nil {
		if !strings.EqualFold(operatorDetailsModifiedEvent.Raw.Address.Hex(), contractAddr.Hex()) {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				operatorDetailsModifiedEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return operatorDetailsModifiedEvent, nil
	}

	return nil, nil
}

// handleRegisterAsOperatorEvent handles the operator registered event
func (h OperatorHandler) handleRegisterAsOperatorEvent(ctx sdk.Context, epochNum uint64, event *pelldelegationmanager.PellDelegationManagerOperatorRegistered) (*sdkmath.Int, error) {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.Keeper.Logger(ctx).Info("handleRegisterAsOperatorEvent", "event", *event, "epochNum", epochNum, "xmsgIndex", inXmsgIndices)

	operator := types.Operator{
		Operator:           event.Operator.Hex(),
		DelegationApprover: event.OperatorDetails.DelegationApprover.Hex(),
		StakerOptOutWindow: event.OperatorDetails.StakerOptOutWindow,
	}

	h.SetOperator(ctx, operator)

	return h.syncOperatorEvent(ctx, uint64(ctx.BlockHeight()), int(event.Raw.Index), operator, buildSyncRegisterAsOperatorEvent)
}

// handleModifyOperatorDetailsEvent handles the operator details modified event
func (h OperatorHandler) handleModifyOperatorDetailsEvent(ctx sdk.Context, epochNum uint64, event *pelldelegationmanager.PellDelegationManagerOperatorDetailsModified) (*sdkmath.Int, error) {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.Keeper.Logger(ctx).Info("handleModifyOperatorDetailsEvent", "event", *event, "epochNum", epochNum, "xmsgIndex", inXmsgIndices)

	operator, found := h.GetOperator(ctx, event.Operator.Hex())
	if !found {
		return nil, fmt.Errorf("handleModifyOperatorDetailsEvent: operator not found")
	}

	operator.DelegationApprover = event.NewOperatorDetails.DelegationApprover.Hex()
	operator.StakerOptOutWindow = event.NewOperatorDetails.StakerOptOutWindow

	h.SetOperator(ctx, operator)

	return h.syncOperatorEvent(ctx, uint64(ctx.BlockHeight()), int(event.Raw.Index), operator, buildSyncModifyOperatorDetails)
}

// buildSyncRegisterAsOperatorEvent builds the xmsg for the operator registered event
func buildSyncRegisterAsOperatorEvent(chainParams *relayertypes.ChainParams, operator types.Operator) (*xmsgtypes.InboundPellEvent, error) {
	// Create the AdapterOperatorDetails struct
	operatorDetails := struct {
		DeprecatedEarningsReceiver ethcommon.Address
		DelegationApprover         ethcommon.Address
		StakerOptOutWindow         uint32
	}{
		DeprecatedEarningsReceiver: ethcommon.Address{}, // zero address
		DelegationApprover:         ethcommon.HexToAddress(operator.DelegationApprover),
		StakerOptOutWindow:         operator.StakerOptOutWindow,
	}

	// Pack the parameters
	data, err := delegationManagerMetaDataABI.Pack(
		SyncOperatorRegisteredEvent,
		ethcommon.HexToAddress(operator.Operator),
		operatorDetails,
	)
	if err != nil {
		return nil, fmt.Errorf("buildSyncRegisterAsOperator: failed to pack parameters: %w", err)
	}

	return &xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_PellSent{
			PellSent: &xmsgtypes.PellSent{
				TxOrigin:        types.ModuleAddressEVM.Hex(),
				Sender:          types.ModuleAddressEVM.Hex(),
				ReceiverChainId: chainParams.ChainId,
				Receiver:        chainParams.DelegationManagerContractAddress,
				Message:         base64.StdEncoding.EncodeToString(data),
				PellParams:      pevmtypes.ReceiveCall.String(),
			},
		},
	}, nil
}

func (h OperatorHandler) syncOperatorEvent(ctx sdk.Context, blockNum uint64, eventIndex int, operator types.Operator, builder eventBuilder) (*sdkmath.Int, error) {
	chainsOutboundState, err := h.GetAllOutboundStates(ctx)
	if err != nil {
		h.Logger(ctx).Error("syncOperatorEvent: failed to get outbound states", "error", err)
		return nil, err
	}

	totalFee := sdkmath.NewInt(0)

	for i, state := range chainsOutboundState {
		chainParams, found := h.relayerKeeper.GetChainParamsByChainID(ctx, int64(state.ChainId))
		if !found {
			h.Logger(ctx).Error("syncOperatorEvent: chain params not found", "chainID", state.ChainId)
			continue
		}

		event, err := builder(chainParams, operator)
		if err != nil {
			h.Logger(ctx).Error("syncOperatorEvent: failed to build event", "error", err)
			continue
		}

		systemTxId := utils.GenerateSystemTxId(pevmtypes.SystemTxTypeSyncOperatorRegistered, blockNum, uint8(i))

		xmsg, err := h.buildXmsg(ctx, systemTxId, blockNum, eventIndex, chainParams, event)
		if err != nil {
			h.Logger(ctx).Error("syncOperatorEvent: failed to build xmsg", "error", err)
			continue
		}

		h.Logger(ctx).Info("syncOperatorEvent: xmsg built", "xmsg", xmsg.Index, "event", event, "chainID", state.ChainId, "systemTxId", systemTxId, "i", i)

		xmsg.SetPendingOutbound("PellConnector pell-send event setting to pending outbound directly")

		// Get gas price and amount
		gasprice, found := h.xmsgKeeper.GetGasPrice(ctx, int64(state.ChainId))
		if !found {
			h.Logger(ctx).Error("syncOperatorEvent: failed to get gas price", "error", err)
			continue
		}

		xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = fmt.Sprintf("%d", gasprice.Prices[gasprice.MedianIndex])

		receiverChain := h.relayerKeeper.GetSupportedChainFromChainID(ctx, int64(state.ChainId))
		if receiverChain == nil {
			h.Logger(ctx).Error("syncOperatorEvent: receiver chain not found", "chainID", state.ChainId)
			continue
		}

		if err = h.xmsgKeeper.ProcessXmsg(ctx, *xmsg, receiverChain); err != nil {
			h.Logger(ctx).Error("syncOperatorEvent: failed to process xmsg", "error", err)
			continue
		}

		crosschainFee, found := h.xmsgKeeper.GetCrosschainEventFee(ctx, int64(state.ChainId))
		h.Logger(ctx).Info("syncOperatorEvent: crosschain fee", "crosschainFee", crosschainFee, "chainID", state.ChainId)

		if found && crosschainFee.IsSupported {
			totalFee = totalFee.Add(crosschainFee.DelegationOperatorSyncFee)
		}
	}

	return &totalFee, nil
}

// buildSyncModifyOperatorDetails builds the xmsg for the operator details modified event
func buildSyncModifyOperatorDetails(chainParams *relayertypes.ChainParams, operator types.Operator) (*xmsgtypes.InboundPellEvent, error) {
	// Create the AdapterOperatorDetails struct
	operatorDetails := struct {
		DeprecatedEarningsReceiver ethcommon.Address
		DelegationApprover         ethcommon.Address
		StakerOptOutWindow         uint32
	}{
		DeprecatedEarningsReceiver: ethcommon.Address{}, // zero address
		DelegationApprover:         ethcommon.HexToAddress(operator.DelegationApprover),
		StakerOptOutWindow:         operator.StakerOptOutWindow,
	}

	// Pack the parameters
	data, err := delegationManagerMetaDataABI.Pack(
		SyncOperatorDetailsModifiedEvent,
		ethcommon.HexToAddress(operator.Operator),
		operatorDetails,
	)
	if err != nil {
		return nil, fmt.Errorf("buildSyncModifyOperatorDetails: failed to pack parameters: %w", err)
	}

	return &xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_PellSent{
			PellSent: &xmsgtypes.PellSent{
				TxOrigin:        types.ModuleAddressEVM.Hex(),
				Sender:          types.ModuleAddressEVM.Hex(),
				ReceiverChainId: chainParams.ChainId,
				Receiver:        chainParams.DelegationManagerContractAddress,
				Message:         base64.StdEncoding.EncodeToString(data),
				PellParams:      pevmtypes.ReceiveCall.String(),
			},
		},
	}, nil
}
