package xmsg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// InitGenesis initializes the xmsg module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {

	// Set all the outTxTracker
	for _, elem := range genState.OutTxTrackerList {
		k.SetOutTxTracker(ctx, elem)
	}

	// Set all the inTxTracker
	for _, elem := range genState.InTxTrackerList {
		k.SetInTxTracker(ctx, elem)
	}

	// Set all the inTxHashToXmsg
	for _, elem := range genState.InTxHashToXmsgList {
		k.SetInTxHashToXmsg(ctx, elem)
	}

	// Set all the gasPrice
	for _, elem := range genState.GasPriceList {
		if elem != nil {
			k.SetGasPrice(ctx, *elem)
		}
	}

	// Set all the chain nonces

	// Set all the last block heights
	for _, elem := range genState.LastBlockHeightList {
		if elem != nil {
			k.SetLastBlockHeight(ctx, *elem)
		}
	}

	// Set all the cross-chain txs
	for _, elem := range genState.Xmsgs {
		if elem != nil {
			k.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, *elem)
		}
	}
	for _, elem := range genState.FinalizedInbounds {
		k.SetFinalizedInbound(ctx, elem)
	}

	k.SetRateLimiterFlags(ctx, genState.RateLimiterFlags)
}

// ExportGenesis returns the xmsg module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var genesis types.GenesisState

	genesis.OutTxTrackerList = k.GetAllOutTxTracker(ctx)
	genesis.InTxHashToXmsgList = k.GetAllInTxHashToXmsg(ctx)
	genesis.InTxTrackerList = k.GetAllInTxTracker(ctx)

	// Get all gas prices
	gasPriceList := k.GetAllGasPrice(ctx)
	for _, elem := range gasPriceList {
		elem := elem
		genesis.GasPriceList = append(genesis.GasPriceList, &elem)
	}

	// Get all last block heights
	lastBlockHeightList := k.GetAllLastBlockHeight(ctx)
	for _, elem := range lastBlockHeightList {
		elem := elem
		genesis.LastBlockHeightList = append(genesis.LastBlockHeightList, &elem)
	}

	// Get all send
	sendList := k.GetAllXmsg(ctx)
	for _, elem := range sendList {
		elem := elem
		genesis.Xmsgs = append(genesis.Xmsgs, &elem)
	}

	genesis.FinalizedInbounds = k.GetAllFinalizedInbound(ctx)

	rateLimiterFlags, found := k.GetRateLimiterFlags(ctx)
	if found {
		genesis.RateLimiterFlags = rateLimiterFlags
	}

	return &genesis
}
