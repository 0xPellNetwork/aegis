// staker_delegated.go
package handler

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v2/strategymanager.sol"
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

var _ interfaces.ChainEventHandler = &StakerDepositHandler{}

type StakerDepositHandler struct {
	EvmJSONRPC       interfaces.EVMJSONRPCClient
	ContractAddr     common.Address
	Contract         *strategymanager.StrategyManager
	ChainId          int64
	CoreChainId      int64
	SignerAddress    string
	InBoundLogger    zerolog.Logger
	ComplianceLogger zerolog.Logger
}

func (h *StakerDepositHandler) HandleBlocks(startBlock, toBlock uint64, eventStore *map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx) (uint64, error) {
	lastScannedBlock := startBlock
	// filter events
	iter, err := h.Contract.FilterDeposit(&bind.FilterOpts{
		Start:   startBlock,
		End:     &toBlock,
		Context: context.TODO(),
	})
	if err != nil {
		return lastScannedBlock - 1, err
	}

	guard := make(map[string]bool)
	for iter.Next() {
		event := iter.Event
		if err := evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, "", evm.TopicsPellStakerDeposited); err != nil {
			h.InBoundLogger.Error().Msgf("failed to validate StakerDeposit event")
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
			h.InBoundLogger.Error().Msgf("multiple StakerDeposit remote call events detected in tx %s", event.Raw.TxHash)
			continue
		}
		guard[event.Raw.TxHash.Hex()] = true

		h.InBoundLogger.Info().Msgf("StakerDeposit event detected in tx %s", event.Raw.TxHash)

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

func (h *StakerDepositHandler) CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error) {
	msgs := make([]*xmsgtypes.MsgVoteOnObservedInboundTx, 0)
	for _, log := range receipt.Logs {
		event, err := h.Contract.ParseDeposit(*log)
		if err == nil && event != nil {
			// sanity check tx event
			if err = evm.ValidateEvmTxLog(&event.Raw, h.ContractAddr, tx.Hash, evm.TopicsPellStakerDeposited); err != nil {
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

func (h *StakerDepositHandler) BuildInboundVoteMsg(
	event *strategymanager.StrategyManagerDeposit,
	sender common.Address,
) *xmsgtypes.MsgVoteOnObservedInboundTx {
	// compliance check
	if config.ContainRestrictedAddress(sender.Hex()) {
		compliance.PrintComplianceLog(h.InBoundLogger, h.ComplianceLogger,
			false, h.ChainId, event.Raw.TxHash.Hex(), sender.Hex(), "", "ERC20")
		return nil
	}

	inboundPellTx := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_StakerDeposited{
			StakerDeposited: &xmsgtypes.StakerDeposited{
				Staker:   event.Staker.Hex(),
				Token:    event.Token.Hex(),
				Strategy: event.Strategy.Hex(),
				Shares:   sdkmath.NewUintFromBigInt(event.Shares),
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
