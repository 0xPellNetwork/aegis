package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultPolicyAddress is the default value for policy address
	DefaultPolicyAddress = "pell14r8nqy53kuruf7pp6aau3d8029ncxnwer54weg"
)

// DefaultPolicies returns the default value for policies
func DefaultPolicies() Policies {
	return Policies{
		Items: []*Policy{
			{
				Address:    DefaultPolicyAddress,
				PolicyType: PolicyType_GROUP_EMERGENCY,
			},
			{
				Address:    DefaultPolicyAddress,
				PolicyType: PolicyType_GROUP_OPERATIONAL,
			},
			{
				Address:    DefaultPolicyAddress,
				PolicyType: PolicyType_GROUP_ADMIN,
			},
		},
	}
}

// Validate performs basic validation of policies
func (p Policies) Validate() error {
	policyTypeMap := make(map[PolicyType]bool)

	// for each policy, check address, policy type, and ensure no duplicate policy types
	for _, policy := range p.Items {
		_, err := sdk.AccAddressFromBech32(policy.Address)
		if err != nil {
			return fmt.Errorf("invalid address: %s", err)
		}

		if policy.PolicyType != PolicyType_GROUP_EMERGENCY && policy.PolicyType != PolicyType_GROUP_ADMIN && policy.PolicyType != PolicyType_GROUP_OPERATIONAL {
			return fmt.Errorf("invalid policy type: %s", policy.PolicyType)
		}

		if policyTypeMap[policy.PolicyType] {
			return fmt.Errorf("duplicate policy type: %s", policy.PolicyType)
		}
		policyTypeMap[policy.PolicyType] = true
	}

	return nil
}
