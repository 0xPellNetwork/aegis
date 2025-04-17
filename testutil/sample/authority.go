package sample

import authoritytypes "github.com/pell-chain/pellcore/x/authority/types"

func Policies() authoritytypes.Policies {
	return authoritytypes.Policies{
		Items: []*authoritytypes.Policy{
			{
				Address:    AccAddress(),
				PolicyType: authoritytypes.PolicyType_GROUP_EMERGENCY,
			},
			{
				Address:    AccAddress(),
				PolicyType: authoritytypes.PolicyType_GROUP_ADMIN,
			},
			{
				Address:    AccAddress(),
				PolicyType: authoritytypes.PolicyType_GROUP_OPERATIONAL,
			},
		},
	}
}
