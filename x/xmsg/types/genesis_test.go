package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				OutTxTrackerList: []types.OutTxTracker{
					{
						Index: "0",
					},
					{
						Index: "1",
					},
				},
				InTxHashToXmsgList: []types.InTxHashToXmsg{
					{
						InTxHash: "0",
					},
					{
						InTxHash: "1",
					},
				},
				GasPriceList: []*types.GasPrice{
					sample.GasPrice_pell(t, "0"),
					sample.GasPrice_pell(t, "1"),
					sample.GasPrice_pell(t, "2"),
				},
				RateLimiterFlags: sample.RateLimiterFlags_pell(),
			},
			valid: true,
		},
		{
			desc: "duplicated outTxTracker",
			genState: &types.GenesisState{
				OutTxTrackerList: []types.OutTxTracker{
					{
						Index: "0",
					},
					{
						Index: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated outTxTracker",
			genState: &types.GenesisState{
				OutTxTrackerList: []types.OutTxTracker{
					{
						Index: "0",
					},
				},
				RateLimiterFlags: types.RateLimiterFlags{
					Enabled: true,
					Window:  -1,
				},
			},
			valid: false,
		},
		{
			desc: "duplicated inTxHashToXmsg",
			genState: &types.GenesisState{
				InTxHashToXmsgList: []types.InTxHashToXmsg{
					{
						InTxHash: "0",
					},
					{
						InTxHash: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated gasPriceList",
			genState: &types.GenesisState{
				GasPriceList: []*types.GasPrice{
					{
						Index: "1",
					},
					{
						Index: "1",
					},
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
