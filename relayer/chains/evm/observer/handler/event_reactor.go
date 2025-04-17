// event_handler.go
package handler

import (
	"fmt"

	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/connector/pellconnector.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/registryinteractor.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v2/strategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/onrik/ethrpc"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/base"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

type EVMEventReactor struct {
	eventRegistry []interfaces.ChainEventHandler
	chainParams   observertypes.ChainParams
	evmClient     interfaces.EVMRPCClient
	logger        *base.ObserverLogger
}

func (r *EVMEventReactor) RegisterEventHandler(handler interfaces.ChainEventHandler) {
	r.eventRegistry = append(r.eventRegistry, handler)
}

func (r *EVMEventReactor) HandleBlocks(startBlock, toBlock uint64) (map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx, uint64) {
	lastScannedBlock := startBlock
	eventStore := make(map[uint64][]*xmsgtypes.MsgVoteOnObservedInboundTx)
	for _, handler := range r.eventRegistry {
		scannedBlock, err := handler.HandleBlocks(startBlock, toBlock, &eventStore)
		if err != nil {
			r.logger.Inbound.Error().Err(err).
				Uint64("start_block", startBlock).
				Uint64("to_block", toBlock).
				Int64("chain_id", r.chainParams.ChainId).
				Str("handler", fmt.Sprintf("%T", handler)).
				Msg("failed to handle blocks")
			continue
		}

		if scannedBlock > lastScannedBlock {
			lastScannedBlock = scannedBlock
		}
	}
	return eventStore, lastScannedBlock
}

func (r *EVMEventReactor) CheckAndBuildInboundVoteMsg(tx *ethrpc.Transaction, receipt *ethtypes.Receipt, lastBlock uint64) ([]*xmsgtypes.MsgVoteOnObservedInboundTx, error) {
	allMsgs := make([]*xmsgtypes.MsgVoteOnObservedInboundTx, 0)
	// check confirmations
	if confirmed := r.hasEnoughConfirmations(receipt, lastBlock); !confirmed {
		return allMsgs, fmt.Errorf("intx %s has not been confirmed yet: receipt block %d", tx.Hash, receipt.BlockNumber.Uint64())
	}
	// check and build inbound vote msg
	for _, handler := range r.eventRegistry {
		msgs, err := handler.CheckAndBuildInboundVoteMsg(tx, receipt)
		if err != nil {
			r.logger.Inbound.Error().Err(err).
				Int64("chain_id", r.chainParams.ChainId).
				Msg("failed to check and build inbound vote msg")
			continue
		}

		if msgs != nil {
			allMsgs = append(allMsgs, msgs...)
		}
	}
	return allMsgs, nil
}

func (r *EVMEventReactor) hasEnoughConfirmations(receipt *ethtypes.Receipt, lastHeight uint64) bool {
	confHeight := receipt.BlockNumber.Uint64() + r.chainParams.ConfirmationCount
	return lastHeight >= confHeight
}

func (r *EVMEventReactor) getStrategyManagerContract() (common.Address, *strategymanager.StrategyManager, error) {
	addr := common.HexToAddress(r.chainParams.StrategyManagerContractAddress)
	contract, err := strategymanager.NewStrategyManager(addr, r.evmClient)
	return addr, contract, err
}

func (r *EVMEventReactor) getDelegationManagerContract() (common.Address, *delegationmanager.DelegationManager, error) {
	addr := common.HexToAddress(r.chainParams.DelegationManagerContractAddress)
	contract, err := delegationmanager.NewDelegationManager(addr, r.evmClient)
	return addr, contract, err
}

func (r *EVMEventReactor) getChainRegistryInteractorContract() (common.Address, *registryinteractor.RegistryInteractor, error) {
	addr := common.HexToAddress(r.chainParams.ChainRegistryInteractorContractAddress)
	contract, err := registryinteractor.NewRegistryInteractor(addr, r.evmClient)
	return addr, contract, err
}

func (r *EVMEventReactor) getConnectorContract() (common.Address, *pellconnector.PellConnector, error) {
	addr := common.HexToAddress(r.chainParams.ConnectorContractAddress)
	contract, err := pellconnector.NewPellConnector(addr, r.evmClient)
	return addr, contract, err
}

func NewEVMEventReactor(
	evmClient interfaces.EVMRPCClient, jsonRPC interfaces.EVMJSONRPCClient, chainParams observertypes.ChainParams,
	chain chains.Chain, coreClient interfaces.PellCoreBridger, logger *base.ObserverLogger) interfaces.IEVMEventReactor {

	reactor := &EVMEventReactor{}
	reactor.evmClient = evmClient
	reactor.chainParams = chainParams
	reactor.logger = logger

	strategyManagerAddr, strategyManager, _ := reactor.getStrategyManagerContract()
	delegationManagerAddr, delegationManager, _ := reactor.getDelegationManagerContract()
	chainId := chain.Id
	coreChainId := coreClient.Chain().Id
	signerAddr := coreClient.GetKeys().GetOperatorAddress().String()
	inboundLogger := logger.Inbound
	complianceLogger := logger.Compliance
	chainRegistryInteractorAddr, chainRegistryInteractor, _ := reactor.getChainRegistryInteractorContract()
	connectorContractAddr, connectorConnector, _ := reactor.getConnectorContract()

	reactor.RegisterEventHandler(&StakerDepositHandler{
		EvmJSONRPC:       jsonRPC,
		ContractAddr:     strategyManagerAddr,
		Contract:         strategyManager,
		ChainId:          chainId,
		CoreChainId:      coreChainId,
		SignerAddress:    signerAddr,
		InBoundLogger:    inboundLogger,
		ComplianceLogger: complianceLogger,
	})
	reactor.RegisterEventHandler(&StakerDelegatedHandler{
		EvmJSONRPC:       jsonRPC,
		ContractAddr:     delegationManagerAddr,
		Contract:         delegationManager,
		ChainId:          chainId,
		CoreChainId:      coreChainId,
		SignerAddress:    signerAddr,
		InBoundLogger:    inboundLogger,
		ComplianceLogger: complianceLogger,
	})
	reactor.RegisterEventHandler(&WithdrawalQueuedHandler{
		EvmJSONRPC:       jsonRPC,
		ContractAddr:     delegationManagerAddr,
		Contract:         delegationManager,
		ChainId:          chainId,
		CoreChainId:      coreChainId,
		SignerAddress:    signerAddr,
		InBoundLogger:    inboundLogger,
		ComplianceLogger: complianceLogger,
	})
	reactor.RegisterEventHandler(&StakerUndelegatedHandler{
		EvmJSONRPC:       jsonRPC,
		ContractAddr:     delegationManagerAddr,
		Contract:         delegationManager,
		ChainId:          chainId,
		CoreChainId:      coreChainId,
		SignerAddress:    signerAddr,
		InBoundLogger:    inboundLogger,
		ComplianceLogger: complianceLogger,
	})
	reactor.RegisterEventHandler(&RegisterChainDVSToPellHandler{
		EvmJSONRPC:       jsonRPC,
		ContractAddr:     chainRegistryInteractorAddr,
		Contract:         chainRegistryInteractor,
		ChainId:          chainId,
		CoreChainId:      coreChainId,
		SignerAddress:    signerAddr,
		InBoundLogger:    inboundLogger,
		ComplianceLogger: complianceLogger,
	})
	reactor.RegisterEventHandler(&PellSentHandler{
		EvmJSONRPC:       jsonRPC,
		ContractAddr:     connectorContractAddr,
		Contract:         connectorConnector,
		ChainId:          chainId,
		CoreChainId:      coreChainId,
		SignerAddress:    signerAddr,
		InBoundLogger:    inboundLogger,
		ComplianceLogger: complianceLogger,
	})
	return reactor
}
