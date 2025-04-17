// staker_delegated.go
package handler

import (
	"context"
	"sort"

	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/registryinteractor.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/onrik/ethrpc"
	"github.com/rs/zerolog"

	"github.com/0xPellNetwork/aegis/relayer/chains/evm"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/compliance"
	"github.com/0xPellNetwork/aegis/relayer/config"
	"github.com/0xPellNetwork/aegis/relayer/pellcore"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ interfaces.ChainEventHandler = &RegisterChainDVSToPellHandler{}

type RegisterChainDVSToPellHandler struct {
	EvmJSONRPC       interfaces.EVMJSONRPCClient
	ContractAddr     common.Address
	Contract         *registryinteractor.RegistryInteractor
	ChainId          int64
	CoreChainId      int64
	SignerAddress    string
	InBoundLogger    zerolog.Logger
	ComplianceLogger zerolog.Logger
}

// Add this struct to hold all three event types
type RegisterToPellEvents struct {
	CentralSchedulerEvent *registryinteractor.RegistryInteractorRegisterCentralSchedulerToPell
	StakeManagerEvent     *registryinteractor.RegistryInteractorRegisterStakeManagerToPell
	EjectionManagerEvent  *registryinteractor.RegistryInteractorRegisterEjectionManagerToPell
}

// TxEvents is used to collect and sort the completed events by block number.
type TxEvents struct {
	BlockNumber uint64
	TxHash      string
	Events      *RegisterToPellEvents
}

// HandleBlocks processes the blocks from startBlock to toBlock and returns the last scanned block
// need three types of events in one tx to build the message
// 1. CentralSchedulerEvent
// 2. StakeManagerEvent
// 3. EjectionManagerEvent
func (h *RegisterChainDVSToPellHandler) HandleBlocks(startBlock, toBlock uint64, eventStore *map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) (uint64, error) {
	lastScannedBlock := startBlock
	// Update the map to use the new struct
	eventMap := make(map[string]*RegisterToPellEvents)

	// Filter and process all three types of events
	if err := h.filterCentralSchedulerEvent(startBlock, toBlock, &lastScannedBlock, eventMap); err != nil {
		return lastScannedBlock - 1, err
	}

	if err := h.filterStakeManagerEvent(startBlock, toBlock, &lastScannedBlock, eventMap); err != nil {
		return lastScannedBlock - 1, err
	}

	if err := h.filterEjectionManagerEvent(startBlock, toBlock, &lastScannedBlock, eventMap); err != nil {
		return lastScannedBlock - 1, err
	}

	// Collect only those txHashes that have all three events, then sort them by block number.
	var txEvents []TxEvents
	for txHash, events := range eventMap {
		if events.CentralSchedulerEvent != nil &&
			events.StakeManagerEvent != nil &&
			events.EjectionManagerEvent != nil {

			// All these events belong to the same transaction, so they share the same BlockNumber
			blockNumber := events.CentralSchedulerEvent.Raw.BlockNumber

			txEvents = append(txEvents, TxEvents{
				TxHash:      txHash,
				Events:      events,
				BlockNumber: blockNumber,
			})
		}
	}

	// Sort by block number in ascending order.
	sort.Slice(txEvents, func(i, j int) bool {
		return txEvents[i].BlockNumber < txEvents[j].BlockNumber
	})

	// Process the collected events in sorted order.
	for _, txEvent := range txEvents {
		tx, err := h.EvmJSONRPC.EthGetTransactionByHash(txEvent.TxHash)
		if err != nil {
			return lastScannedBlock - 1, err
		}

		if err = evm.ValidateEvmTransaction(tx); err != nil {
			return lastScannedBlock - 1, err
		}

		blockNumber := txEvent.Events.CentralSchedulerEvent.Raw.BlockNumber
		sender := common.HexToAddress(tx.From)

		if msg := h.BuildInboundVoteMsg(txEvent.Events, sender); msg != nil {
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

func (h *RegisterChainDVSToPellHandler) filterCentralSchedulerEvent(startBlock, toBlock uint64, lastScannedBlock *uint64, eventMap map[string]*RegisterToPellEvents) error {
	iter, err := h.Contract.FilterRegisterCentralSchedulerToPell(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})

	if err != nil {
		h.InBoundLogger.Error().Err(err).
			Uint64("start_block", startBlock).
			Uint64("to_block", toBlock).
			Str("contract", h.ContractAddr.Hex()).
			Msg("failed to filter CentralSchedulerToPell event")
		return err
	}
	defer iter.Close()

	for iter.Next() {
		event := iter.Event
		*lastScannedBlock = event.Raw.BlockNumber

		if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, event.Raw.TxHash.String(), evm.TopicsCentralSchedulerToPell); err != nil {
			h.InBoundLogger.Error().Err(err).
				Str("tx_hash", event.Raw.TxHash.String()).
				Msg("failed to validate CentralSchedulerToPell event")
			continue
		}

		txHash := event.Raw.TxHash.Hex()
		h.InBoundLogger.Info().
			Str("tx_hash", txHash).
			Msg("CentralSchedulerToPell type event detected")

		if _, exists := eventMap[txHash]; !exists {
			eventMap[txHash] = &RegisterToPellEvents{}
		}
		eventMap[txHash].CentralSchedulerEvent = event
	}

	return nil
}

func (h *RegisterChainDVSToPellHandler) filterStakeManagerEvent(startBlock, toBlock uint64, lastScannedBlock *uint64, eventMap map[string]*RegisterToPellEvents) error {
	iter, err := h.Contract.FilterRegisterStakeManagerToPell(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})

	if err != nil {
		h.InBoundLogger.Error().Err(err).
			Uint64("start_block", startBlock).
			Uint64("to_block", toBlock).
			Msg("failed to filter RegisterStakeManagerToPell event")
		return err
	}
	defer iter.Close()

	for iter.Next() {
		event := iter.Event
		*lastScannedBlock = event.Raw.BlockNumber

		if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, event.Raw.TxHash.String(), evm.TopicsRegisterStakeManagerToPell); err != nil {
			h.InBoundLogger.Error().Err(err).
				Str("tx_hash", event.Raw.TxHash.String()).
				Msg("failed to validate RegisterStakeManagerToPell event")
			continue
		}

		txHash := event.Raw.TxHash.Hex()
		h.InBoundLogger.Info().
			Str("tx_hash", txHash).
			Msg("RegisterStakeManagerToPell type event detected")

		if _, exists := eventMap[txHash]; !exists {
			eventMap[txHash] = &RegisterToPellEvents{}
		}
		eventMap[txHash].StakeManagerEvent = event
	}

	return nil
}

func (h *RegisterChainDVSToPellHandler) filterEjectionManagerEvent(startBlock, toBlock uint64, lastScannedBlock *uint64, eventMap map[string]*RegisterToPellEvents) error {
	iter, err := h.Contract.FilterRegisterEjectionManagerToPell(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})

	if err != nil {
		h.InBoundLogger.Error().Err(err).
			Uint64("start_block", startBlock).
			Uint64("to_block", toBlock).
			Msg("failed to filter RegisterEjectionManagerToPell event")
		return err
	}
	defer iter.Close()

	for iter.Next() {
		event := iter.Event
		*lastScannedBlock = event.Raw.BlockNumber

		if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, event.Raw.TxHash.String(), evm.TopicsRegisterEjectionManagerToPell); err != nil {
			h.InBoundLogger.Error().Err(err).
				Str("tx_hash", event.Raw.TxHash.String()).
				Msg("failed to validate RegisterEjectionManagerToPell event")
			continue
		}

		txHash := event.Raw.TxHash.Hex()
		h.InBoundLogger.Info().
			Str("tx_hash", txHash).
			Msg("RegisterEjectionManagerToPell type event detected")

		if _, exists := eventMap[txHash]; !exists {
			eventMap[txHash] = &RegisterToPellEvents{}
		}
		eventMap[txHash].EjectionManagerEvent = event
	}

	return nil
}

func (h *RegisterChainDVSToPellHandler) CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error) {
	msgs := make([]*xmsgtypes.MsgVoteOnObservedInboundTx, 0)
	eventMap := make(map[string]*RegisterToPellEvents)

	// Process all logs looking for our three event types
	for _, log := range receipt.Logs {
		// Try to parse as CentralSchedulerEvent
		if event, err := h.Contract.ParseRegisterCentralSchedulerToPell(*log); err == nil && event != nil {
			if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, tx.Hash, evm.TopicsCentralSchedulerToPell); err != nil {
				h.InBoundLogger.Error().Err(err).
					Str("tx_hash", tx.Hash).
					Msg("failed to validate CentralSchedulerToPell event")
				continue
			}
			txHash := event.Raw.TxHash.Hex()
			if _, exists := eventMap[txHash]; !exists {
				eventMap[txHash] = &RegisterToPellEvents{}
			}
			eventMap[txHash].CentralSchedulerEvent = event
		}

		// Try to parse as StakeManagerEvent
		if stakeEvent, err := h.Contract.ParseRegisterStakeManagerToPell(*log); err == nil && stakeEvent != nil {
			if err = evm.ValidateEvmTxLog(&stakeEvent.Raw, h.ContractAddr, tx.Hash, evm.TopicsRegisterStakeManagerToPell); err != nil {
				h.InBoundLogger.Error().Err(err).
					Str("tx_hash", tx.Hash).
					Msg("failed to validate RegisterStakeManagerToPell event")
				continue
			}
			txHash := stakeEvent.Raw.TxHash.Hex()
			if _, exists := eventMap[txHash]; !exists {
				eventMap[txHash] = &RegisterToPellEvents{}
			}
			eventMap[txHash].StakeManagerEvent = stakeEvent
		}

		// Try to parse as EjectionManagerEvent
		if ejectionEvent, err := h.Contract.ParseRegisterEjectionManagerToPell(*log); err == nil && ejectionEvent != nil {
			if err = evm.ValidateEvmTxLog(&ejectionEvent.Raw, h.ContractAddr, tx.Hash, evm.TopicsRegisterEjectionManagerToPell); err != nil {
				h.InBoundLogger.Error().Err(err).
					Str("tx_hash", tx.Hash).
					Msg("failed to validate RegisterEjectionManagerToPell event")
				continue
			}
			txHash := ejectionEvent.Raw.TxHash.Hex()
			if _, exists := eventMap[txHash]; !exists {
				eventMap[txHash] = &RegisterToPellEvents{}
			}
			eventMap[txHash].EjectionManagerEvent = ejectionEvent
		}
	}

	// Check for complete event sets and build messages
	for txHash, events := range eventMap {
		// Only process if we have all three events
		if events.CentralSchedulerEvent != nil &&
			events.StakeManagerEvent != nil &&
			events.EjectionManagerEvent != nil {

			h.InBoundLogger.Info().
				Str("tx_hash", txHash).
				Msg("Found complete event set")

			if msg := h.BuildInboundVoteMsg(events, common.HexToAddress(tx.From)); msg != nil {
				msgs = append(msgs, msg)
			}
		}
	}

	return msgs, nil
}

func (h *RegisterChainDVSToPellHandler) BuildInboundVoteMsg(
	event *RegisterToPellEvents,
	sender common.Address,
) *xmsgtypes.MsgVoteOnObservedInboundTx {
	// compliance check
	if config.ContainRestrictedAddress(sender.Hex()) {
		compliance.PrintComplianceLog(h.InBoundLogger, h.ComplianceLogger,
			false, h.ChainId, event.CentralSchedulerEvent.Raw.TxHash.Hex(), sender.Hex(), "", "ERC20")
		return nil
	}
	inboundPellTx := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_RegisterChainDvsToPell{
			RegisterChainDvsToPell: &xmsgtypes.RegisterChainDVSToPell{
				ChainId:              uint64(h.ChainId),
				RegistryRouterOnPell: event.CentralSchedulerEvent.RegistryRouterOnPell.Hex(),
				CentralScheduler:     event.CentralSchedulerEvent.CentralScheduler.Hex(),
				DvsChainApproverSignature: &xmsgtypes.SignatureWithSaltAndExpiry{
					Signature: event.CentralSchedulerEvent.DvsChainApproverSignature.Signature,
					Salt:      event.CentralSchedulerEvent.DvsChainApproverSignature.Salt[:],
					Expiry:    event.CentralSchedulerEvent.DvsChainApproverSignature.Expiry.Uint64(),
				},
				EjectionManager: event.EjectionManagerEvent.EjectionManager.Hex(),
				StakeManager:    event.StakeManagerEvent.StakeManager.Hex(),
			},
		},
	}
	return pellcore.GetInBoundVoteMessage(
		sender.Hex(),
		h.ChainId,
		event.CentralSchedulerEvent.RegistryRouterOnPell.Hex(),
		event.CentralSchedulerEvent.RegistryRouterOnPell.Hex(),
		h.CoreChainId,
		event.CentralSchedulerEvent.Raw.TxHash.Hex(),
		event.CentralSchedulerEvent.Raw.BlockNumber,
		0,
		h.SignerAddress,
		event.CentralSchedulerEvent.Raw.Index,
		inboundPellTx,
	)
}
