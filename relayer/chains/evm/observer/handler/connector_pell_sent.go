// staker_delegated.go
package handler

import (
	"context"
	"encoding/base64"

	"cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/connector/pellconnector.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/onrik/ethrpc"
	"github.com/rs/zerolog"

	"github.com/pell-chain/pellcore/relayer/chains/evm"
	"github.com/pell-chain/pellcore/relayer/chains/interfaces"
	"github.com/pell-chain/pellcore/relayer/compliance"
	"github.com/pell-chain/pellcore/relayer/config"
	"github.com/pell-chain/pellcore/relayer/pellcore"
	evmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var _ interfaces.ChainEventHandler = &PellSentHandler{}

type PellSentHandler struct {
	EvmJSONRPC       interfaces.EVMJSONRPCClient
	ContractAddr     common.Address
	Contract         *pellconnector.PellConnector
	ChainId          int64
	CoreChainId      int64
	SignerAddress    string
	InBoundLogger    zerolog.Logger
	ComplianceLogger zerolog.Logger
}

func (h *PellSentHandler) HandleBlocks(startBlock, toBlock uint64, eventStore *map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) (uint64, error) {
	lastScannedBlock := startBlock
	// filter events
	iter, err := h.Contract.FilterPellSent(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	}, nil, nil)

	if err != nil {
		return lastScannedBlock - 1, err
	}

	guard := make(map[string]bool)
	for iter.Next() {
		event := iter.Event
		if err := evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, "", evm.TopicsPellSent); err != nil {
			h.InBoundLogger.Error().Err(err).Msg("failed to validate PellSent event")
			continue
		}

		tx, err := h.EvmJSONRPC.EthGetTransactionByHash(event.Raw.TxHash.Hex())
		if err != nil {
			return lastScannedBlock - 1, err
		}

		if err := evm.ValidateEvmTransaction(tx); err != nil {
			return lastScannedBlock - 1, err
		}

		// guard against multiple events in the same tx
		if guard[event.Raw.TxHash.Hex()] {
			h.InBoundLogger.Error().Msgf("multiple PellSent remote call events detected in tx %s", event.Raw.TxHash)
			continue
		}
		guard[event.Raw.TxHash.Hex()] = true

		h.InBoundLogger.Info().Msgf("PellSent event detected in tx %s", event.Raw.TxHash)

		// build inbound vote message
		blockNumber := event.Raw.BlockNumber
		sender := common.HexToAddress(tx.From)
		if msg := h.BuildInboundVoteMsg(event, sender); msg != nil {
			if msgs, ok := (*eventStore)[blockNumber]; !ok {
				(*eventStore)[blockNumber] = []*xmsgtypes.MsgVoteOnObservedInboundTx{msg}
			} else {
				msgs = append(msgs, msg)
				(*eventStore)[blockNumber] = msgs
			}
		}
	}

	return toBlock, nil
}

func (h *PellSentHandler) CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error) {
	msgs := make([]*xmsgtypes.MsgVoteOnObservedInboundTx, 0)
	for _, log := range receipt.Logs {
		event, err := h.Contract.ParsePellSent(*log)
		if err == nil && event != nil {
			// sanity check tx event
			if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, tx.Hash, evm.TopicsPellSent); err != nil {
				h.InBoundLogger.Error().Err(err).Msgf("checkEvmTxLog error on intx %s chain %d", tx.Hash, h.ChainId)
				return msgs, err
			}

			if msg := h.BuildInboundVoteMsg(event, common.HexToAddress(tx.From)); msg != nil {
				msgs = append(msgs, msg)
				return msgs, nil
			}

			break // only one event is allowed per tx
		}
	}
	return msgs, nil
}

func (h *PellSentHandler) BuildInboundVoteMsg(
	event *pellconnector.PellConnectorPellSent,
	sender common.Address,
) *xmsgtypes.MsgVoteOnObservedInboundTx {
	// compliance check
	if config.ContainRestrictedAddress(sender.Hex()) {
		compliance.PrintComplianceLog(h.InBoundLogger, h.ComplianceLogger,
			false, h.ChainId, event.Raw.TxHash.Hex(), sender.Hex(), "", "ERC20")
		return nil
	}

	paramType, err := evmtypes.ToPellSentParamType(event.PellParams)
	if err != nil {
		h.InBoundLogger.Error().Err(err).Msg("failed to convert PellSentParamType")
		return nil
	}

	inboundPellTx := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_PellSent{
			PellSent: &xmsgtypes.PellSent{
				TxOrigin:            event.Raw.TxHash.Hex(),
				Sender:              event.Raw.Address.Hex(),
				ReceiverChainId:     event.DestinationChainId.Int64(),
				Receiver:            common.BytesToAddress(event.DestinationAddress).String(),
				Message:             base64.StdEncoding.EncodeToString(event.Message),
				PellParams:          paramType.String(),
				PellValue:           math.NewUintFromBigInt(event.PellValueAndGas),
				DestinationGasLimit: math.NewUintFromBigInt(event.DestinationGasLimit),
			},
		},
	}
	return pellcore.GetInBoundVoteMessage(
		sender.Hex(),
		h.ChainId,
		event.PellTxSenderAddress.Hex(),
		event.Raw.Address.Hex(),
		h.CoreChainId,
		event.Raw.TxHash.Hex(),
		event.Raw.BlockNumber,
		event.DestinationGasLimit.Uint64(),
		h.SignerAddress,
		event.Raw.Index,
		inboundPellTx,
	)
}
