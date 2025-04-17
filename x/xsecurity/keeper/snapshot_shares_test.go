package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// TestDiffOperatorWeightedSharesWith_NoDifferences tests that no differences are returned when oldList and newList are identical.
func TestDiffOperatorWeightedSharesWith_NoDifferences(t *testing.T) {
	oldList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
			{
				OperatorAddress: "op2",
				WeightedShare:   sdkmath.NewInt(200),
			},
		},
	}
	newList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
			{
				OperatorAddress: "op2",
				WeightedShare:   sdkmath.NewInt(200),
			},
		},
	}

	diffs := DiffOperatorWeightedSharesWith(oldList, newList)
	require.Empty(t, diffs, "when oldList and newList are the same, diffs should be empty")
}

// TestDiffOperatorWeightedSharesWith_AddedOperator tests the scenario where a new operator is added in newList.
func TestDiffOperatorWeightedSharesWith_AddedOperator(t *testing.T) {
	oldList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
		},
	}
	newList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
			{
				OperatorAddress: "op2", // Added operator
				WeightedShare:   sdkmath.NewInt(200),
			},
		},
	}

	diffs := DiffOperatorWeightedSharesWith(oldList, newList)
	require.Len(t, diffs, 1, "should detect one added operator")
	require.Equal(t, "op2", diffs[0].OperatorAddress)
	require.True(t, diffs[0].WeightedShare.Equal(sdkmath.NewInt(200)))
}

// TestDiffOperatorWeightedSharesWith_ModifiedOperator tests the scenario where an operator's WeightedShare is modified in newList.
func TestDiffOperatorWeightedSharesWith_ModifiedOperator(t *testing.T) {
	oldList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
		},
	}
	newList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(150), // Modified value
			},
		},
	}

	diffs := DiffOperatorWeightedSharesWith(oldList, newList)
	require.Len(t, diffs, 1, "should detect one modified operator")
	require.Equal(t, "op1", diffs[0].OperatorAddress)
	require.True(t, diffs[0].WeightedShare.Equal(sdkmath.NewInt(150)))
}

// TestDiffOperatorWeightedSharesWith_DeletedOperator tests the scenario where an operator present in oldList is deleted in newList.
func TestDiffOperatorWeightedSharesWith_DeletedOperator(t *testing.T) {
	oldList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
			{
				OperatorAddress: "op2",
				WeightedShare:   sdkmath.NewInt(200),
			},
		},
	}
	newList := &types.LSTOperatorWeightedShareList{
		OperatorWeightedShares: []*types.LSTOperatorWeightedShare{
			{
				OperatorAddress: "op1",
				WeightedShare:   sdkmath.NewInt(100),
			},
		},
	}

	diffs := DiffOperatorWeightedSharesWith(oldList, newList)
	require.Len(t, diffs, 1, "should detect one deleted operator")
	require.Equal(t, "op2", diffs[0].OperatorAddress)
	// The deleted operator's WeightedShare should be set to zero
	require.True(t, diffs[0].WeightedShare.Equal(sdkmath.ZeroInt()))
}
