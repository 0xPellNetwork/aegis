package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/pkg/coin"
	"github.com/0xPellNetwork/aegis/x/emissions/types"
)

func (k Keeper) GetBlockRewardComponents(ctx sdk.Context) (sdkmath.LegacyDec, sdkmath.LegacyDec, sdkmath.LegacyDec) {
	reservesFactor := k.GetReservesFactor(ctx)

	if reservesFactor.LTE(sdkmath.LegacyZeroDec()) {
		return sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()
	}
	bondFactor := k.GetBondFactor(ctx, k.GetStakingKeeper())
	durationFactor := k.GetDurationFactor(ctx)
	return reservesFactor, bondFactor, durationFactor
}

func (k Keeper) GetBondFactor(ctx sdk.Context, stakingKeeper types.StakingKeeper) sdkmath.LegacyDec {
	targetBondRatio := sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).TargetBondRatio)
	maxBondFactor := sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).MaxBondFactor)
	minBondFactor := sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).MinBondFactor)

	currentBondedRatio, err := stakingKeeper.BondedRatio(ctx)
	// Bond factor ranges between minBondFactor (0.75) to maxBondFactor (1.25)
	if err != nil || currentBondedRatio.IsZero() {
		return sdkmath.LegacyZeroDec()
	}
	bondFactor := targetBondRatio.Quo(currentBondedRatio)
	if bondFactor.GT(maxBondFactor) {
		return maxBondFactor
	}
	if bondFactor.LT(minBondFactor) {
		return minBondFactor
	}
	return bondFactor
}

func (k Keeper) GetDurationFactor(ctx sdk.Context) sdkmath.LegacyDec {
	avgBlockTime := sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).AvgBlockTime)
	NumberOfBlocksInAMonth := sdkmath.LegacyNewDec(types.SecsInMonth).Quo(avgBlockTime)
	monthFactor := sdkmath.LegacyNewDec(ctx.BlockHeight()).Quo(NumberOfBlocksInAMonth)
	logValueDec := sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).DurationFactorConstant)
	// month * log(1 + 0.02 / 12)
	fractionNumerator := monthFactor.Mul(logValueDec)
	// (month * log(1 + 0.02 / 12) ) + 1
	fractionDenominator := fractionNumerator.Add(sdkmath.LegacyOneDec())

	// (month * log(1 + 0.02 / 12)) / (month * log(1 + 0.02 / 12) ) + 1
	if fractionDenominator.IsZero() {
		return sdkmath.LegacyOneDec()
	}
	if fractionNumerator.IsZero() {
		return sdkmath.LegacyZeroDec()
	}
	return fractionNumerator.Quo(fractionDenominator)
}

func (k Keeper) GetReservesFactor(ctx sdk.Context) sdkmath.LegacyDec {
	reserveAmount := k.GetBankKeeper().GetBalance(ctx, types.EmissionsModuleAddress, config.BaseDenom)
	return sdkmath.LegacyNewDecFromInt(reserveAmount.Amount)
}

func (k Keeper) GetFixedBlockRewards() (sdkmath.LegacyDec, error) {
	return CalculateFixedValidatorRewards(types.AvgBlockTime)
}

func CalculateFixedValidatorRewards(avgBlockTimeString string) (sdkmath.LegacyDec, error) {
	apellAmountTotalRewards, err := coin.GetApellDecFromAmountInPell(types.BlockRewardsInPell)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}
	avgBlockTime, err := sdkmath.LegacyNewDecFromStr(avgBlockTimeString)
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}
	numberOfBlocksInAMonth := sdkmath.LegacyNewDec(types.SecsInMonth).Quo(avgBlockTime)
	numberOfBlocksTotal := numberOfBlocksInAMonth.Mul(sdkmath.LegacyNewDec(12)).Mul(sdkmath.LegacyNewDec(types.EmissionScheduledYears))
	constantRewardPerBlock := apellAmountTotalRewards.Quo(numberOfBlocksTotal)
	return constantRewardPerBlock, nil
}
