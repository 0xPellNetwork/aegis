package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

func TestMergeMultipleShares(t *testing.T) {
	tests := []struct {
		name       string
		sharesList [][]*types.OperatorShares
		want       []*types.OperatorShares
	}{
		{
			name:       "empty list",
			sharesList: nil,
			want:       nil,
		},
		{
			name: "single shares array",
			sharesList: [][]*types.OperatorShares{
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(100)},
				},
			},
			want: []*types.OperatorShares{
				{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(100)},
			},
		},
		{
			name: "merge with duplicates",
			sharesList: [][]*types.OperatorShares{
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(100)},
					{ChainId: 1, Operator: "op2", Strategy: "s1", Shares: math.NewInt(200)},
				},
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(300)},
					{ChainId: 1, Operator: "op3", Strategy: "s1", Shares: math.NewInt(400)},
				},
			},
			want: []*types.OperatorShares{
				{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(300)},
				{ChainId: 1, Operator: "op2", Strategy: "s1", Shares: math.NewInt(200)},
				{ChainId: 1, Operator: "op3", Strategy: "s1", Shares: math.NewInt(400)},
			},
		},
		{
			name: "multiple chains and strategies",
			sharesList: [][]*types.OperatorShares{
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(100)},
					{ChainId: 2, Operator: "op1", Strategy: "s1", Shares: math.NewInt(200)},
				},
				{
					{ChainId: 1, Operator: "op1", Strategy: "s2", Shares: math.NewInt(300)},
					{ChainId: 2, Operator: "op1", Strategy: "s1", Shares: math.NewInt(400)},
				},
			},
			want: []*types.OperatorShares{
				{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(100)},
				{ChainId: 1, Operator: "op1", Strategy: "s2", Shares: math.NewInt(300)},
				{ChainId: 2, Operator: "op1", Strategy: "s1", Shares: math.NewInt(400)},
			},
		},
		{
			name: "three arrays with duplicates",
			sharesList: [][]*types.OperatorShares{
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(100)},
				},
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(200)},
				},
				{
					{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(300)},
				},
			},
			want: []*types.OperatorShares{
				{ChainId: 1, Operator: "op1", Strategy: "s1", Shares: math.NewInt(300)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectLatestOperatorSharesFromSnapshots(tt.sharesList)

			// verify length
			require.Equal(t, len(tt.want), len(got), "length mismatch")

			// verify each element
			for i := 0; i < len(got); i++ {
				require.Equal(t, tt.want[i].ChainId, got[i].ChainId, "ChainId mismatch at index %d", i)
				require.Equal(t, tt.want[i].Operator, got[i].Operator, "Operator mismatch at index %d", i)
				require.Equal(t, tt.want[i].Strategy, got[i].Strategy, "Strategy mismatch at index %d", i)
				require.True(t, tt.want[i].Shares.Equal(got[i].Shares), "Shares mismatch at index %d", i)
			}
		})
	}
}

// test compare operator shares
func TestCompareOperator(t *testing.T) {
	tests := []struct {
		name string
		a    *types.OperatorShares
		b    *types.OperatorShares
		want int
	}{
		{
			name: "equal operators",
			a:    &types.OperatorShares{ChainId: 1, Operator: "op1", Strategy: "s1"},
			b:    &types.OperatorShares{ChainId: 1, Operator: "op1", Strategy: "s1"},
			want: 0,
		},
		{
			name: "different chain ids",
			a:    &types.OperatorShares{ChainId: 1, Operator: "op1", Strategy: "s1"},
			b:    &types.OperatorShares{ChainId: 2, Operator: "op1", Strategy: "s1"},
			want: -1,
		},
		{
			name: "different operators",
			a:    &types.OperatorShares{ChainId: 1, Operator: "op1", Strategy: "s1"},
			b:    &types.OperatorShares{ChainId: 1, Operator: "op2", Strategy: "s1"},
			want: -1,
		},
		{
			name: "different strategies",
			a:    &types.OperatorShares{ChainId: 1, Operator: "op1", Strategy: "s1"},
			b:    &types.OperatorShares{ChainId: 1, Operator: "op1", Strategy: "s2"},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareOperator(tt.a, tt.b)
			require.Equal(t, tt.want, got)
		})
	}
}
