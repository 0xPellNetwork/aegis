package types

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance

func NewParams() Params {
	return Params{
		MaxBondFactor:               "1.25",
		MinBondFactor:               "0.75",
		AvgBlockTime:                "6.00",
		TargetBondRatio:             "00.67",
		ValidatorEmissionPercentage: "00.50",
		ObserverEmissionPercentage:  "00.15",
		TssSignerEmissionPercentage: "00.25",
		DurationFactorConstant:      "0.001877876953694702",
		TssGasEmissionPercentage:    "00.10",

		// ObserverSlashAmount is currently disabled
		// TODO: enable this param
		ObserverSlashAmount: sdkmath.Int{},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPrefix(ParamMaxBondFactor), &p.MaxBondFactor, validateMaxBondFactor),
		paramtypes.NewParamSetPair(KeyPrefix(ParamMinBondFactor), &p.MinBondFactor, validateMinBondFactor),
		paramtypes.NewParamSetPair(KeyPrefix(ParamAvgBlockTime), &p.AvgBlockTime, validateAvgBlockTime),
		paramtypes.NewParamSetPair(KeyPrefix(ParamTargetBondRatio), &p.TargetBondRatio, validateTargetBondRatio),
		paramtypes.NewParamSetPair(KeyPrefix(ParamValidatorEmissionPercentage), &p.ValidatorEmissionPercentage, validateValidatorEmissionPercentage),
		paramtypes.NewParamSetPair(KeyPrefix(ParamObserverEmissionPercentage), &p.ObserverEmissionPercentage, validateObserverEmissionPercentage),
		paramtypes.NewParamSetPair(KeyPrefix(ParamTssSignerEmissionPercentage), &p.TssSignerEmissionPercentage, validateTssEmissionPercentage),
		paramtypes.NewParamSetPair(KeyPrefix(ParamDurationFactorConstant), &p.DurationFactorConstant, validateDurationFactorConstant),
		paramtypes.NewParamSetPair(KeyPrefix(ParamTssGasEmissionPercentage), &p.TssGasEmissionPercentage, validateTssGasEmissionPercentage),

		// TODO: enable this param
		//paramtypes.NewParamSetPair(KeyPrefix(ParamObserverSlashAmount), &p.ObserverSlashAmount, validateObserverSlashAmount),
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

func validateDurationFactorConstant(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateMaxBondFactor(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	decMaxBond := sdkmath.LegacyMustNewDecFromStr(v)
	if decMaxBond.GT(sdkmath.LegacyMustNewDecFromStr("1.25")) {
		return fmt.Errorf("max bond factor cannot be higher that 0.25")
	}
	return nil
}

func validateMinBondFactor(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	decMaxBond := sdkmath.LegacyMustNewDecFromStr(v)
	if decMaxBond.LT(sdkmath.LegacyMustNewDecFromStr("0.75")) {
		return fmt.Errorf("min bond factor cannot be lower that 0.75")
	}
	return nil
}

func validateAvgBlockTime(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	blocktime, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("invalid block time: %T", i)
	}
	if blocktime <= 0 {
		return fmt.Errorf("block time cannot be less than or equal to 0")
	}
	return nil
}

func validateTargetBondRatio(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	sdkmath.LegacyZeroDec()
	decMaxBond := sdkmath.LegacyMustNewDecFromStr(v)
	if decMaxBond.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("target bond ratio cannot be more than 100 percent")
	}
	if decMaxBond.LT(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("target bond ratio cannot be less than 0 percent")
	}
	return nil
}

func validateValidatorEmissionPercentage(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	dec := sdkmath.LegacyMustNewDecFromStr(v)
	if dec.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("validator emission percentage cannot be more than 100 percent")
	}
	if dec.LT(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("validator emission percentage cannot be less than 0 percent")
	}
	return nil
}

func validateObserverEmissionPercentage(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	dec := sdkmath.LegacyMustNewDecFromStr(v)
	if dec.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("observer emission percentage cannot be more than 100 percent")
	}
	if dec.LT(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("observer emission percentage cannot be less than 0 percent")
	}
	return nil
}

func validateTssEmissionPercentage(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	dec := sdkmath.LegacyMustNewDecFromStr(v)
	if dec.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("tss emission percentage cannot be more than 100 percent")
	}
	if dec.LT(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("tss emission percentage cannot be less than 0 percent")
	}
	return nil
}

func validateTssGasEmissionPercentage(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	dec := sdkmath.LegacyMustNewDecFromStr(v)
	if dec.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("tss gas emission percentage cannot be more than 100 percent")
	}
	if dec.LT(sdkmath.LegacyZeroDec()) {
		return fmt.Errorf("tss gas emission percentage cannot be less than 0 percent")
	}
	return nil
}
