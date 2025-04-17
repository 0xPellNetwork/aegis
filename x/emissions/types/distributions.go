package types

import (
	sdkmath "cosmossdk.io/math"
)

// GetRewardsDistributions returns the current distribution of rewards
// for validators, observers and TSS signers
// If the percentage is not set, it returns 0
func GetRewardsDistributions(params Params) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	// Fetch the validator rewards, use 0 if the percentage is not set
	validatorRewards := sdkmath.ZeroInt()
	validatorRewardsDec, err := sdkmath.LegacyNewDecFromStr(params.ValidatorEmissionPercentage)
	if err == nil {
		validatorRewards = validatorRewardsDec.Mul(BlockReward).TruncateInt()
	}

	// Fetch the observer rewards, use 0 if the percentage is not set
	observerRewards := sdkmath.ZeroInt()
	observerRewardsDec, err := sdkmath.LegacyNewDecFromStr(params.ObserverEmissionPercentage)
	if err == nil {
		observerRewards = observerRewardsDec.Mul(BlockReward).TruncateInt()
	}

	// Fetch the TSS signer rewards, use 0 if the percentage is not set
	tssSignerRewards := sdkmath.NewInt(0)
	tssSignerRewardsDec, err := sdkmath.LegacyNewDecFromStr(params.TssSignerEmissionPercentage)
	if err == nil {
		tssSignerRewards = tssSignerRewardsDec.Mul(BlockReward).TruncateInt()
	}

	// Fetch the TSS gas reserve rewards, use 0 if the percentage is not set
	tssGasReserve := sdkmath.NewInt(0)
	tssGasReserveDec, err := sdkmath.LegacyNewDecFromStr(params.TssGasEmissionPercentage)
	if err == nil {
		tssGasReserve = tssGasReserveDec.Mul(BlockReward).TruncateInt()
	}

	return validatorRewards, observerRewards, tssSignerRewards, tssGasReserve
}
