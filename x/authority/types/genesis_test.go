package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/authority/types"
)

func TestGenesisState_Validate(t *testing.T) {
	setConfig(t)

	tests := []struct {
		name        string
		gs          *types.GenesisState
		errContains string
	}{
		{
			name:        "default is valid",
			gs:          types.DefaultGenesis(),
			errContains: "",
		},
		{
			name: "valid genesis",
			gs: &types.GenesisState{
				Policies: sample.Policies(),
			},
			errContains: "",
		},
		{
			name: "invalid if policies is invalid",
			gs: &types.GenesisState{
				Policies: types.Policies{
					Items: []*types.Policy{
						{
							Address:    "invalid",
							PolicyType: types.PolicyType_GROUP_EMERGENCY,
						},
					},
				},
			},
			errContains: "invalid address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gs.Validate()
			if tt.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
