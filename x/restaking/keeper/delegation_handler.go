package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var _ xmsgtypes.EventHandler = DelegationHandler{}

type DelegationHandler struct {
	Keeper
}

// HandleEvent handles the event from the log
func (h DelegationHandler) HandleEvent(ctx sdk.Context, epochNum uint64, _ ethcommon.Address, logs []*ethtypes.Log, txOrigin string) ([]*xmsgtypes.CrossChainFee, error) {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.Keeper.Logger(ctx).Info("restaking DelegationShareEventHandler", "txOrigin", txOrigin, "xmsgIndex", inXmsgIndices)

	addr, err := h.GetContractAddress(ctx)
	if err != nil || addr == (ethcommon.Address{}) {
		return nil, err
	}

	for _, log := range logs {
		event, err := h.ParseEvent(addr, log)
		if err != nil {
			continue
		}

		h.Logger(ctx).Info("delegation handler parsed event",
			"event", event,
			"xmsgIndex", inXmsgIndices,
		)

		switch evt := event.(type) {
		case *pelldelegationmanager.PellDelegationManagerOperatorSharesIncreased:
			return nil, h.handleSharesIncreasedEvent(ctx, epochNum, evt)
		case *pelldelegationmanager.PellDelegationManagerOperatorSharesDecreased:
			return nil, h.handleSharesDecreasedEvent(ctx, epochNum, evt)
		}
	}

	return nil, nil
}

// get delegation manager contract address
func (h DelegationHandler) GetContractAddress(ctx sdk.Context) (ethcommon.Address, error) {
	return h.pevmKeeper.GetPellDelegationManagerProxyContractAddress(ctx)
}

// parse event from log
func (h DelegationHandler) ParseEvent(contractAddr ethcommon.Address, log *ethtypes.Log) (interface{}, error) {
	if len(log.Topics) == 0 {
		return nil, fmt.Errorf("ParseEvent: invalid log - no topics")
	}

	delegationManager, err := pelldelegationmanager.NewPellDelegationManagerFilterer(contractAddr, bind.ContractFilterer(nil))
	if err != nil {
		return nil, err
	}

	// try to parse OperatorSharesIncreased event
	if increasedEvent, err := delegationManager.ParseOperatorSharesIncreased(*log); err == nil {
		if increasedEvent.Raw.Address != contractAddr {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				increasedEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return increasedEvent, nil
	}

	// try to parse OperatorSharesDecreased event
	if decreasedEvent, err := delegationManager.ParseOperatorSharesDecreased(*log); err == nil {
		if decreasedEvent.Raw.Address != contractAddr {
			return nil, fmt.Errorf("ParseEvent: event address %s does not match delegation manager %s",
				decreasedEvent.Raw.Address.Hex(), contractAddr.Hex())
		}

		return decreasedEvent, nil
	}

	return nil, nil
}

// handle shares increased event
func (h DelegationHandler) handleSharesIncreasedEvent(ctx sdk.Context, epochNum uint64, event *pelldelegationmanager.PellDelegationManagerOperatorSharesIncreased) error {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.Keeper.Logger(ctx).Info("handleSharesIncreasedEvent", "event", *event, "epochNum", epochNum, "xmsgIndex", inXmsgIndices)

	operatorShares := h.GetOperatorShares(ctx, event.ChainId.Uint64(), event.Operator.Hex(), event.Strategy.Hex())
	if operatorShares == nil {
		operatorShares = &types.OperatorShares{
			ChainId:  event.ChainId.Uint64(),
			Operator: event.Operator.Hex(),
			Strategy: event.Strategy.Hex(),
			Shares:   sdkmath.NewInt(0),
		}
	}

	operatorShares.Shares = operatorShares.Shares.Add(sdkmath.NewIntFromBigInt(event.Shares))

	h.SetOperatorShares(ctx, event.ChainId.Uint64(), event.Operator.Hex(), event.Strategy.Hex(), operatorShares.Shares)

	// share change record
	h.storeChangedOperatorShares(ctx, epochNum, operatorShares)

	return nil
}

// store share change
func (h DelegationHandler) storeChangedOperatorShares(ctx sdk.Context, epochNum uint64, operatorShares *types.OperatorShares) {
	shareChange, exist := h.GetChangedOperatorSharesSnapshot(ctx, epochNum)
	if !exist {
		h.SetChangedOperatorSharesSnapshot(ctx, epochNum, []*types.OperatorShares{operatorShares})
		return
	}

	// Find and update existing share or append new one
	for _, share := range shareChange.OperatorShares {
		if share.ChainId == operatorShares.ChainId &&
			share.Operator == operatorShares.Operator &&
			share.Strategy == operatorShares.Strategy {
			share.Shares = operatorShares.Shares
			h.SetChangedOperatorSharesSnapshot(ctx, epochNum, shareChange.OperatorShares)
			return
		}
	}

	shareChange.OperatorShares = append(shareChange.OperatorShares, operatorShares)

	h.Logger(ctx).Info("store ChangedOperatorSharesSnapshot", "shareChange", shareChange, "epochNumber", epochNum)

	h.SetChangedOperatorSharesSnapshot(ctx, epochNum, shareChange.OperatorShares)
}

// handle shares decreased event
func (h DelegationHandler) handleSharesDecreasedEvent(ctx sdk.Context, epochNum uint64, event *pelldelegationmanager.PellDelegationManagerOperatorSharesDecreased) error {
	inXmsgIndices, _ := ctx.Value("inXmsgIndices").(string)
	h.Keeper.Logger(ctx).Info("handleSharesDecreasedEvent", "event", *event, "epochNum", epochNum, "xmsgIndex", inXmsgIndices)

	operatorShares := h.GetOperatorShares(ctx, event.ChainId.Uint64(), event.Operator.Hex(), event.Strategy.Hex())
	if operatorShares == nil {
		return fmt.Errorf("handleSharesDecreasedEvent: operator shares not found")
	}

	if operatorShares.Shares.LT(sdkmath.NewIntFromBigInt(event.Shares)) {
		return fmt.Errorf("handleSharesDecreasedEvent: operator shares less than event shares")
	}

	operatorShares.Shares = operatorShares.Shares.Sub(sdkmath.NewIntFromBigInt(event.Shares))

	h.SetOperatorShares(ctx, event.ChainId.Uint64(), event.Operator.Hex(), event.Strategy.Hex(), operatorShares.Shares)

	//  operator shares snapshot
	h.storeChangedOperatorShares(ctx, epochNum, operatorShares)
	return nil
}
