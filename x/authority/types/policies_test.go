package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/app"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/authority/types"
)

// setConfig sets the global config to use pell chain's bech32 prefixes
func setConfig(t *testing.T) {
	defer func(t *testing.T) {
		if r := recover(); r != nil {
			t.Log("config is already sealed", r)
		}
	}(t)
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.Seal()
}

func TestPolicies_Validate(t *testing.T) {
	setConfig(t)
	// use table driven tests to test the validation of policies
	tests := []struct {
		name        string
		policies    types.Policies
		errContains string
	}{
		{
			name:        "empty is valid",
			policies:    types.Policies{},
			errContains: "",
		},
		{
			name:        "default is valid",
			policies:    types.DefaultPolicies(),
			errContains: "",
		},
		{
			name: "policies with all group",
			policies: types.Policies{
				Items: []*types.Policy{
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_EMERGENCY,
					},
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_ADMIN,
					},
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_OPERATIONAL,
					},
				},
			},
			errContains: "",
		},
		{
			name: "valid if a policy type is not existing",
			policies: types.Policies{
				Items: []*types.Policy{
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_EMERGENCY,
					},
				},
			},
			errContains: "",
		},
		{
			name: "invalid if address is invalid",
			policies: types.Policies{
				Items: []*types.Policy{
					{
						Address:    "invalid",
						PolicyType: types.PolicyType_GROUP_EMERGENCY,
					},
				},
			},
			errContains: "invalid address",
		},
		{
			name: "invalid if policy type is invalid",
			policies: types.Policies{
				Items: []*types.Policy{
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType(1000),
					},
				},
			},
			errContains: "invalid policy type",
		},
		{
			name: "invalid if duplicated policy type",
			policies: types.Policies{
				Items: []*types.Policy{
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_EMERGENCY,
					},
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_ADMIN,
					},
					{
						Address:    sample.AccAddress(),
						PolicyType: types.PolicyType_GROUP_EMERGENCY,
					},
				},
			},
			errContains: "duplicate policy type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policies.Validate()
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
