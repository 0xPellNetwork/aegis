package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"

	"github.com/pell-chain/pellcore/pkg/chains"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for pellObserver module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(observerParams []*RelayerParams, adminParams []*AdminPolicy, ballotMaturityBlocks int64) Params {
	return Params{
		ObserverParams:       observerParams,
		AdminPolicy:          adminParams,
		BallotMaturityBlocks: ballotMaturityBlocks,
	}
}

// DefaultParams returns a default set of parameters.
// privnet chains are supported by default for testing purposes
// custom params must be provided in genesis for other networks
func DefaultParams() Params {
	chains := chains.FindChains(func(c chains.Chain) bool { return c.NetworkType == chains.NetWorkType_PRIVNET })
	observerParams := make([]*RelayerParams, len(chains))
	for i, chain := range chains {
		observerParams[i] = &RelayerParams{
			IsSupported:          true,
			Chain:                &chain,
			BallotThreshold:      sdkmath.LegacyMustNewDecFromStr("0.66"),
			MinRelayerDelegation: sdkmath.LegacyMustNewDecFromStr("1000000000000000000000"), // 1000 PELL
		}
	}
	return NewParams(observerParams, DefaultAdminPolicy(), 100)
}

func DefaultAdminPolicy() []*AdminPolicy {
	return []*AdminPolicy{
		{
			PolicyType: PolicyType_GROUP1,
			Address:    GroupID1Address,
		},
		{
			PolicyType: PolicyType_GROUP2,
			Address:    GroupID1Address,
		},
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPrefix(ObserverParamsKey), &p.ObserverParams, validateVotingThresholds),
		paramtypes.NewParamSetPair(KeyPrefix(AdminPolicyParamsKey), &p.AdminPolicy, validateAdminPolicy),
		paramtypes.NewParamSetPair(KeyPrefix(BallotMaturityBlocksParamsKey), &p.BallotMaturityBlocks, validateBallotMaturityBlocks),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, err := yaml.Marshal(p)
	if err != nil {
		return ""
	}
	return string(out)
}

// Deprecated: observer params are now stored in core params
func validateVotingThresholds(i interface{}) error {
	v, ok := i.([]*RelayerParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, threshold := range v {
		if threshold.BallotThreshold.GT(sdkmath.LegacyOneDec()) {
			return ErrParamsThreshold
		}
	}
	return nil
}

func validateAdminPolicy(i interface{}) error {
	_, ok := i.([]*AdminPolicy)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBallotMaturityBlocks(i interface{}) error {
	_, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
