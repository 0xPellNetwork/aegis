package emissions

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/x/emissions/keeper"
	"github.com/0xPellNetwork/aegis/x/emissions/types"
)

func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	emissionPoolBalance := keeper.GetReservesFactor(ctx)
	blockRewards := types.BlockReward
	if blockRewards.GT(emissionPoolBalance) {
		ctx.Logger().Info(fmt.Sprintf("Block rewards %s are greater than emission pool balance %s", blockRewards.String(), emissionPoolBalance.String()))
		return
	}

	// Get the distribution of rewards
	params := keeper.GetParamsIfExists(ctx)
	validatorRewards, observerRewards, tssSignerRewards, tssGasReserve := types.GetRewardsDistributions(params)

	// TODO : Replace hardcoded slash amount with a parameter
	slashAmount, ok := sdkmath.NewIntFromString(types.ObserverSlashAmount)
	if !ok {
		ctx.Logger().Error(fmt.Sprintf("Error while parsing observer slash amount %s", types.ObserverSlashAmount))
		return
	}

	// Use a tmpCtx, which is a cache-wrapped context to avoid writing to the store
	// We commit only if all three distributions are successful, if not the funds stay in the emission pool
	tmpCtx, commit := ctx.CacheContext()
	err := DistributeValidatorRewards(tmpCtx, validatorRewards, keeper.GetBankKeeper(), keeper.GetFeeCollector())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error while distributing validator rewards %s", err))
		return
	}
	err = DistributeObserverRewards(tmpCtx, observerRewards, keeper, slashAmount)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error while distributing observer rewards %s", err))
		return
	}
	err = DistributeTssRewards(tmpCtx, tssSignerRewards, keeper.GetBankKeeper())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error while distributing tss signer rewards %s", err))
		return
	}
	err = DistributeTssGasReserve(tmpCtx, tssGasReserve, keeper.GetBankKeeper())
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error while distributing pevm rewards %s", err))
		return
	}
	commit()

	types.EmitValidatorEmissions(ctx, "", "",
		"",
		validatorRewards.String(),
		observerRewards.String(),
		tssSignerRewards.String())
}

// DistributeValidatorRewards distributes the rewards to validators who signed the block .
// The block proposer gets a bonus reward
// This function uses the distribution module of cosmos-sdk , by directly sending funds to the feecollector.
func DistributeValidatorRewards(ctx sdk.Context, amount sdkmath.Int, bankKeeper types.BankKeeper, feeCollector string) error {
	coin := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount))
	ctx.Logger().Info(fmt.Sprintf("Distributing Validator Rewards Total:%s To FeeCollector", amount.String()))
	return bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, feeCollector, coin)
}

// DistributeObserverRewards distributes the rewards to all observers who voted in any of the matured ballots
// The total rewards are distributed equally among all Successful votes
// NotVoted or Unsuccessful votes are slashed
// rewards given or slashed amounts are in apell
func DistributeObserverRewards(
	ctx sdk.Context,
	amount sdkmath.Int,
	keeper keeper.Keeper,
	slashAmount sdkmath.Int,
) error {

	rewardsDistributer := map[string]int64{}
	totalRewardsUnits := int64(0)
	err := keeper.GetBankKeeper().SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedObserverRewardsPool, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	if err != nil {
		return err
	}
	ballotIdentifiers := keeper.GetObserverKeeper().GetMaturedBallotList(ctx)
	// do not distribute rewards if no ballots are matured, the rewards can accumulate in the undistributed pool
	if len(ballotIdentifiers) == 0 {
		return nil
	}
	for _, ballotIdentifier := range ballotIdentifiers {
		ballot, found := keeper.GetObserverKeeper().GetBallot(ctx, ballotIdentifier)
		if !found {
			continue
		}
		totalRewardsUnits += ballot.BuildRewardsDistribution(rewardsDistributer)
	}
	rewardPerUnit := sdkmath.ZeroInt()
	if totalRewardsUnits > 0 && amount.IsPositive() {
		rewardPerUnit = amount.Quo(sdkmath.NewInt(totalRewardsUnits))
	}
	ctx.Logger().Debug(fmt.Sprintf("Total Rewards Units : %d , rewards per Unit %s ,number of ballots :%d", totalRewardsUnits, rewardPerUnit.String(), len(ballotIdentifiers)))
	sortedKeys := make([]string, 0, len(rewardsDistributer))
	for k := range rewardsDistributer {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var finalDistributionList []*types.RelayerEmission
	for _, key := range sortedKeys {
		observerAddress, err := sdk.AccAddressFromBech32(key)
		if err != nil {
			ctx.Logger().Error("Error while parsing observer address ", "error", err, "address", key)
			continue
		}
		// observerRewardUnits can be negative if the observer has been slashed
		// an observers earn 1 unit for a correct vote, and -1 unit for an incorrect vote
		observerRewardUnits := rewardsDistributer[key]

		if observerRewardUnits == 0 {
			finalDistributionList = append(finalDistributionList, &types.RelayerEmission{
				EmissionType:    types.EmissionType_SLASH,
				ObserverAddress: observerAddress.String(),
				Amount:          sdkmath.ZeroInt(),
			})
			continue
		}
		if observerRewardUnits < 0 {
			keeper.SlashObserverEmission(ctx, observerAddress.String(), slashAmount)
			finalDistributionList = append(finalDistributionList, &types.RelayerEmission{
				EmissionType:    types.EmissionType_SLASH,
				ObserverAddress: observerAddress.String(),
				Amount:          slashAmount,
			})
			continue
		}

		// Defensive check
		if rewardPerUnit.GT(sdkmath.ZeroInt()) {
			rewardAmount := rewardPerUnit.Mul(sdkmath.NewInt(observerRewardUnits))
			keeper.AddObserverEmission(ctx, observerAddress.String(), rewardAmount)
			finalDistributionList = append(finalDistributionList, &types.RelayerEmission{
				EmissionType:    types.EmissionType_REWARDS,
				ObserverAddress: observerAddress.String(),
				Amount:          rewardAmount,
			})
		}
	}
	types.EmitObserverEmissions(ctx, finalDistributionList)
	// TODO : Delete Ballots after distribution
	return nil
}

// DistributeTssRewards trasferes the allocated rewards to the Undistributed Tss Rewards Pool.
// This is done so that the reserves factor is properly calculated in the next block
func DistributeTssRewards(ctx sdk.Context, amount sdkmath.Int, bankKeeper types.BankKeeper) error {
	coin := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount))
	ctx.Logger().Info(fmt.Sprintf("Distributing Tss Rewards Total:%s To Tss module", amount.String()))
	return bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedTssRewardsPool, coin)
}

func DistributeTssGasReserve(ctx sdk.Context, amount sdkmath.Int, bankKeeper types.BankKeeper) error {
	coin := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount))
	ctx.Logger().Info(fmt.Sprintf("Distributing Tss Gas Reserve Total:%s To PEvm module", amount.String()))
	return bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedTssGasReservePool, coin)
}
