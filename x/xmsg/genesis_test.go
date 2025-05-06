package xmsg_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/nullify"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	xmsg "github.com/0xPellNetwork/aegis/x/xmsg"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		OutTxTrackerList: []types.OutTxTracker{
			sample.OutTxTracker_pell(t, "0"),
			sample.OutTxTracker_pell(t, "1"),
			sample.OutTxTracker_pell(t, "2"),
		},
		InTxTrackerList: []types.InTxTracker{
			sample.InTxTracker_pell(t, "0"),
			sample.InTxTracker_pell(t, "1"),
			sample.InTxTracker_pell(t, "2"),
		},
		FinalizedInbounds: []string{
			sample.Hash().String(),
			sample.Hash().String(),
			sample.Hash().String(),
		},
		GasPriceList: []*types.GasPrice{
			sample.GasPrice_pell(t, "0"),
			sample.GasPrice_pell(t, "1"),
			sample.GasPrice_pell(t, "2"),
		},
		Xmsgs: []*types.Xmsg{
			sample.Xmsg_pell(t, "0"),
			sample.Xmsg_pell(t, "1"),
			sample.Xmsg_pell(t, "2"),
		},
		LastBlockHeightList: []*types.LastBlockHeight{
			sample.LastBlockHeight_pell(t, "0"),
			sample.LastBlockHeight_pell(t, "1"),
			sample.LastBlockHeight_pell(t, "2"),
		},
		InTxHashToXmsgList: []types.InTxHashToXmsg{
			sample.InTxHashToXmsg_pell(t, "0x0"),
			sample.InTxHashToXmsg_pell(t, "0x1"),
			sample.InTxHashToXmsg_pell(t, "0x2"),
		},
		RateLimiterFlags: sample.RateLimiterFlags_pell(),
	}

	// Init and export
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	xmsg.InitGenesis(ctx, *k, genesisState)
	got := xmsg.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	// Compare genesis after init and export
	nullify.Fill(&genesisState)
	nullify.Fill(got)
	require.Equal(t, genesisState, *got)
}
