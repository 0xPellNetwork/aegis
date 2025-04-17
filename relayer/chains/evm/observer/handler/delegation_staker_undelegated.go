// staker_delegated.go
package handler

import (
	"context"

	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
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
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var _ interfaces.ChainEventHandler = &StakerUndelegatedHandler{}

type StakerUndelegatedHandler struct {
	EvmJSONRPC       interfaces.EVMJSONRPCClient
	ContractAddr     common.Address
	Contract         *delegationmanager.DelegationManager
	ChainId          int64
	CoreChainId      int64
	SignerAddress    string
	InBoundLogger    zerolog.Logger
	ComplianceLogger zerolog.Logger
}

func (h *StakerUndelegatedHandler) HandleBlocks(startBlock, toBlock uint64, eventStore *map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) (uint64, error) {
	lastScannedBlock := startBlock
	// filter events
	iter, err := h.Contract.FilterStakerUndelegated(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	}, []common.Address{}, []common.Address{})
	if err != nil {
		return lastScannedBlock - 1, err
	}

	guard := make(map[string]bool)
	for iter.Next() {
		event := iter.Event
		if err := evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, "", evm.TopicsPellStakerUndelegated); err != nil { // [signature, stakerAddress, operatorAddress]
			h.InBoundLogger.Error().Msgf("failed to validate StakerUndelegated event")
			continue
		}

		tx, err := h.EvmJSONRPC.EthGetTransactionByHash(event.Raw.TxHash.Hex())
		if err != nil {
			return lastScannedBlock - 1, err
		}

		if err = evm.ValidateEvmTransaction(tx); err != nil {
			return lastScannedBlock - 1, err
		}

		// guard against multiple events in the same tx
		if guard[event.Raw.TxHash.Hex()] {
			h.InBoundLogger.Error().
				Str("tx_hash", event.Raw.TxHash.Hex()).
				Msg("multiple StakerUndelegated remote call events detected")
			continue
		}
		guard[event.Raw.TxHash.Hex()] = true

		h.InBoundLogger.Info().
			Str("tx_hash", event.Raw.TxHash.Hex()).
			Msg("StakerUndelegated event detected")

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

func (h *StakerUndelegatedHandler) CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error) {
	msgs := make([]*xmsgtypes.MsgVoteOnObservedInboundTx, 0)
	for _, log := range receipt.Logs {
		event, err := h.Contract.ParseStakerUndelegated(*log)
		if err == nil && event != nil {
			// sanity check tx event
			if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, tx.Hash, evm.TopicsPellStakerUndelegated); err != nil {
				h.InBoundLogger.Error().Err(err).
					Str("tx_hash", tx.Hash).
					Msg("checkEvmTxLog error")
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

func (h *StakerUndelegatedHandler) BuildInboundVoteMsg(
	event *delegationmanager.DelegationManagerStakerUndelegated,
	sender common.Address,
) *xmsgtypes.MsgVoteOnObservedInboundTx {
	// compliance check
	if config.ContainRestrictedAddress(sender.Hex()) {
		compliance.PrintComplianceLog(h.InBoundLogger, h.ComplianceLogger,
			false, h.ChainId, event.Raw.TxHash.Hex(), sender.Hex(), "", "ERC20")
		return nil
	}

	inboundPellTx := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_StakerUndelegated{
			StakerUndelegated: &xmsgtypes.StakerUndelegated{
				Staker:   event.Staker.Hex(),
				Operator: event.Operator.Hex(),
			},
		},
	}
	return pellcore.GetInBoundVoteMessage(
		sender.Hex(),
		h.ChainId,
		event.Staker.Hex(),
		event.Staker.Hex(),
		h.CoreChainId,
		event.Raw.TxHash.Hex(),
		event.Raw.BlockNumber,
		0,
		h.SignerAddress,
		event.Raw.Index,
		inboundPellTx,
	)
}
