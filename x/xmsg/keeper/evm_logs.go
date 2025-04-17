package keeper

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pell-chain/pellcore/pkg/chains"
	evmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func (k Keeper) ProcessXmsg(ctx sdk.Context, xmsg types.Xmsg, receiverChain *chains.Chain) error {
	inXmsgIndices, ok := ctx.Value("inXmsgIndices").(string)
	if ok {
		xmsg.InboundTxParams.InboundTxHash = inXmsgIndices
	}

	if err := k.UpdateNonce(ctx, receiverChain.Id, &xmsg); err != nil {
		k.Logger(ctx).Error("ProcessXmsg: update nonce failed", "error", err)
		return fmt.Errorf("ProcessWithdrawalEvent: update nonce failed: %s", err.Error())
	}

	k.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, xmsg)
	ctx.Logger().Info("ProcessXmsg successfully processed", "xmsgIndex", xmsg.Index, "inXmsgIndices", inXmsgIndices)
	return nil
}

// ProcessPellSentEvent creates a new Xmsg to process the pellsent event
// error indicates system error and non-recoverable; should abort
//
// event PellSent(
//
//	  address sourceTxOriginAddress,
//	  address indexed pellTxSenderAddress,
//	  uint256 indexed destinationChainId,
//	  bytes destinationAddress,
//	  bytes message,
//	  bytes pellParams
//	);
func (k Keeper) ProcessPellSentEvent(
	ctx sdk.Context,
	event *pellconnector.PellConnectorPellSent,
	emittingContract ethcommon.Address,
	txOrigin string,
	tss observertypes.TSS,
) error {
	ctx.Logger().Info(fmt.Sprintf("ProcessPellSentEvent sent to %s to chain with chainId %d",
		hex.EncodeToString(event.DestinationAddress), event.DestinationChainId))

	receiverChainID := event.DestinationChainId
	receiverChain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, receiverChainID.Int64())
	if receiverChain == nil {
		return errorsmod.Wrapf(observertypes.ErrSupportedChains, "chain with chainID %d not supported", receiverChainID)
	}
	chainParams, found := k.relayerKeeper.GetChainParamsByChainID(ctx, receiverChain.Id)
	if !found {
		return observertypes.ErrChainParamsNotFound
	}
	if receiverChain.IsExternalChain() &&
		(chainParams.ConnectorContractAddress == "") {
		return types.ErrUnableToSendCoinType
	}
	toAddr := "0x" + hex.EncodeToString(event.DestinationAddress)

	senderChain, err := chains.PellChainFromChainID(ctx.ChainID())
	if err != nil {
		return fmt.Errorf("ProcessPellSentEvent: failed to convert chainID: %s", err.Error())
	}

	paramType, err := evmtypes.ToPellSentParamType(event.PellParams)
	if err != nil {
		return fmt.Errorf("ProcessPellSentEvent: failed to convert PellSentParamType: %s", err.Error())
	}

	pellSent := types.InboundPellEvent{
		PellData: &types.InboundPellEvent_PellSent{
			PellSent: &types.PellSent{
				TxOrigin:            event.SourceTxOriginAddress.Hex(),
				Sender:              event.PellTxSenderAddress.Hex(),
				ReceiverChainId:     receiverChain.Id,
				Receiver:            hex.EncodeToString(event.DestinationAddress),
				Message:             base64.StdEncoding.EncodeToString(event.Message),
				PellParams:          paramType.String(),
				PellValue:           math.NewUintFromBigInt(event.PellValueAndGas),
				DestinationGasLimit: math.NewUintFromBigInt(event.DestinationGasLimit),
			},
		},
	}

	msg := types.NewMsgVoteOnObservedInboundTx(
		"",
		emittingContract.Hex(),
		senderChain.Id,
		txOrigin,
		toAddr,
		receiverChain.Id,
		event.Raw.TxHash.String(),
		event.Raw.BlockNumber,
		chainParams.GasLimit,
		event.Raw.Index,
		pellSent,
	)
	ctx.Logger().Debug("ProcessPellSentEvent: created MsgVoteOnObservedInboundTx", "msg gas limit", chainParams.GasLimit)

	// Create a new xmsg with status as pending Inbound, this is created directly from the event without waiting for any observer votes
	xmsg, err := types.NewXmsg(ctx, *msg, tss.TssPubkey)
	if err != nil {
		return fmt.Errorf("ProcessPellSentEvent: failed to initialize xmsg: %s", err.Error())
	}
	xmsg.SetPendingOutbound("PellConnector pell-send event setting to pending outbound directly")
	// Get gas price and amount
	gasprice, found := k.GetGasPrice(ctx, receiverChain.Id)
	if !found {
		return fmt.Errorf("gasprice not found for %s", receiverChain)
	}
	xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = fmt.Sprint(gasprice.Prices[gasprice.MedianIndex])

	EmitEventPellSent(ctx, xmsg)
	return k.ProcessXmsg(ctx, xmsg, receiverChain)
}

// ParsePellSentEvent tries extracting PellSent event from the input logs using the pell send contract;
// It only returns a not-nil event if the event has been correctly validated as a valid PellSent event
func ParsePellSentEvent(log ethtypes.Log, connectorPEVM ethcommon.Address) (*pellconnector.PellConnectorPellSent, error) {
	if len(log.Topics) == 0 {
		return nil, fmt.Errorf("ParsePellSentEvent: invalid log - no topics")
	}
	pellConnectorPEVM, err := pellconnector.NewPellConnectorFilterer(log.Address, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	event, err := pellConnectorPEVM.ParsePellSent(log)
	if err != nil {
		return nil, err
	}

	if event.Raw.Address != connectorPEVM {
		return nil, fmt.Errorf("ParsePellSentEvent: event address %s does not match connector %s",
			event.Raw.Address.Hex(), connectorPEVM.Hex())
	}
	return event, nil
}
