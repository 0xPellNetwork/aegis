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

	"github.com/0xPellNetwork/aegis/relayer/chains/evm"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/compliance"
	"github.com/0xPellNetwork/aegis/relayer/config"
	"github.com/0xPellNetwork/aegis/relayer/pellcore"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ interfaces.ChainEventHandler = &WithdrawalQueuedHandler{}

type WithdrawalQueuedHandler struct {
	EvmJSONRPC       interfaces.EVMJSONRPCClient
	ContractAddr     common.Address
	Contract         *delegationmanager.DelegationManager
	ChainId          int64
	CoreChainId      int64
	SignerAddress    string
	InBoundLogger    zerolog.Logger
	ComplianceLogger zerolog.Logger
}

func (h *WithdrawalQueuedHandler) HandleBlocks(startBlock, toBlock uint64, eventStore *map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) (uint64, error) {
	lastScannedBlock := startBlock
	// filter events
	iter, err := h.Contract.FilterWithdrawalQueued(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})
	if err != nil {
		return lastScannedBlock - 1, err
	}

	for iter.Next() {
		event := iter.Event
		if err := evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, "", evm.TopicsPellWithdrawalQueued); err != nil { // [signature, stakerAddress, operatorAddress]
			h.InBoundLogger.Error().Msgf("failed to validate WithdrawalQueued event")
			continue
		}

		tx, err := h.EvmJSONRPC.EthGetTransactionByHash(event.Raw.TxHash.Hex())
		if err != nil {
			return lastScannedBlock - 1, err
		}

		if err := evm.ValidateEvmTransaction(tx); err != nil {
			return lastScannedBlock - 1, err
		}

		h.InBoundLogger.Info().
			Str("tx_hash", event.Raw.TxHash.Hex()).
			Msg("WithdrawalQueued event detected")

		// don't use guard against multiple events in the same tx
		// as it is possible to have multiple events in withdrawal queued
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

func (h *WithdrawalQueuedHandler) CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error) {
	msgs := make([]*xmsgtypes.MsgVoteOnObservedInboundTx, 0)
	for _, log := range receipt.Logs {
		event, err := h.Contract.ParseWithdrawalQueued(*log)
		if err == nil && event != nil {
			// sanity check tx event
			if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, tx.Hash, evm.TopicsPellWithdrawalQueued); err != nil {
				h.InBoundLogger.Error().Err(err).
					Str("tx_hash", tx.Hash).
					Msg("checkEvmTxLog error on WithdrawalQueued")
				return msgs, err
			}
			if msg := h.BuildInboundVoteMsg(event, common.HexToAddress(tx.From)); msg != nil {
				msgs = append(msgs, msg)
				return msgs, nil
			}
			// withdrawal queued event may have multiple events in the same tx
			// so not breaking here
		}
	}
	return msgs, nil
}

func (h *WithdrawalQueuedHandler) BuildInboundVoteMsg(
	event *delegationmanager.DelegationManagerWithdrawalQueued,
	sender common.Address,
) *xmsgtypes.MsgVoteOnObservedInboundTx {
	// compliance check
	if config.ContainRestrictedAddress(sender.Hex()) {
		compliance.PrintComplianceLog(h.InBoundLogger, h.ComplianceLogger,
			false, h.ChainId, event.Raw.TxHash.Hex(), sender.Hex(), "", "ERC20")
		return nil
	}
	inboundPellTx := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_WithdrawalQueued{
			WithdrawalQueued: &xmsgtypes.WithdrawalQueued{
				WithdrawalRoot: event.WithdrawalRoot[:],
				Withdrawal:     ConvertToWithdrawal(event.Withdrawal),
			},
		},
	}
	return pellcore.GetInBoundVoteMessage(
		sender.Hex(),
		h.ChainId,
		event.Withdrawal.Withdrawer.Hex(),
		event.Withdrawal.Staker.Hex(),
		h.CoreChainId,
		event.Raw.TxHash.Hex(),
		event.Raw.BlockNumber,
		0,
		h.SignerAddress,
		event.Raw.Index,
		inboundPellTx,
	)
}

func ConvertToWithdrawal(event delegationmanager.IDelegationManagerWithdrawal) *xmsgtypes.Withdrawal {
	strategies := make([]string, len(event.Strategies))
	for i, addr := range event.Strategies {
		strategies[i] = addr.Hex()
	}
	shares := make([]string, len(event.Shares))
	for i, share := range event.Shares {
		shares[i] = share.String()
	}
	return &xmsgtypes.Withdrawal{
		Staker:         event.Staker.Hex(),
		DelegatedTo:    event.DelegatedTo.Hex(),
		Withdrawer:     event.Withdrawer.Hex(),
		Nonce:          event.Nonce.String(),
		StartTimestamp: event.StartTimestamp,
		Strategies:     strategies,
		Shares:         shares,
	}
}
