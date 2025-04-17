package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ProcessEpoch processes each epoch for the xsecurity module
// When conditions are met, calculates and returns validator updates
// 1. Gather the current state of native staking validators
// 2. Calculate the total native validator voting power
// 3. Derive the total LST voting power based on a configured ratio
// 4. Distribute the LST voting power among registered operators proportionally to their shares
// 5. Generate validator updates that combine native staking power with LST-derived power
func (k Keeper) ProcessEpoch(goctx context.Context) ([]types.ValidatorUpdate, error) {
	ctx := sdk.UnwrapSDKContext(goctx)
	logger := k.Logger(ctx)

	// 1. Check if LST staking is enabled
	enabled, exist := k.GetLSTStakingEnabled(ctx)
	if !exist || !enabled.Enabled {
		logger.Info("LST staking not enabled")
		return nil, nil
	}

	// 2. Check if we should create a shares snapshot
	blocksPerEpoch, _ := k.GetBlocksPerEpoch(ctx)
	if blocksPerEpoch == 0 {
		logger.Error("failed to get blocks per epoch or blocks per epoch is set to an invalid value", "error", "data not found")
		return nil, nil
	}

	blockHeight := ctx.BlockHeight()
	// Only process at epoch boundaries
	if blockHeight%int64(blocksPerEpoch) != 0 {
		return nil, nil
	}

	// ignore if epoch number is not found
	epochNumber, _ := k.GetEpochNumber(ctx)
	defer k.SetEpochNumber(ctx, epochNumber+1)

	logger.Info("processing epoch", "blockHeight", blockHeight, "blocksPerEpoch", blocksPerEpoch, "currentEpochNumber", epochNumber)

	// 3. Create restaking module shares snapshot
	if err := k.SnapshotSharesFromRestakingModuleByEpoch(ctx); err != nil {
		logger.Error("failed to snapshot shares from restaking module", "error", err)
		return nil, nil // Don't return error to avoid preventing new block creation
	}

	// 4. Get changed shares
	changedShares, err := k.GetChangedSharesByEpoch(ctx)
	if err != nil {
		logger.Error("failed to get changed operator shares", "error", err)
		return nil, nil
	}

	// 5. Get native validator information
	nativeValidators, err := k.stakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		logger.Error("failed to get native validators", "error", err)
		return nil, nil
	}

	// 6. Calculate current native voting power
	nativeTotalVotingPower, err := k.CalcNativeVotingPower(ctx, nativeValidators)
	if err != nil {
		logger.Error("failed to calculate native voting power", "error", err)
		return nil, nil
	}

	// 7. Check if recalculation is needed
	lastNativeVotingPower, _ := k.GetLastNativeVotingPower(ctx)
	if len(changedShares) == 0 && nativeTotalVotingPower.Equal(math.NewInt(lastNativeVotingPower)) {
		logger.Info("no need to recalculate validator updates")
		return nil, nil
	}

	// 8. Update latest native voting power
	k.SetLastNativeVotingPower(ctx, nativeTotalVotingPower.Int64())

	// 9. Calculate validator updates
	updates, err := k.CalcValidatorUpdatesByLSTTokenShares(ctx, nativeTotalVotingPower)
	if err != nil {
		logger.Error("failed to calculate validator updates from LST token shares", "error", err)
		return nil, nil
	}

	logger.Info("calculated validator updates based on LST token shares", "updateCount", len(updates))
	return updates, nil
}

// CalcValidatorUpdatesByLSTTokenShares calculates validator updates based on LST token shares
// If the total voting power in ctx.VoteInfo() is 5000, then the total voting power of LST is 2500.
// Query the specific staking weight information in OperatorShares and allocate the 2500 voting power
// proportionally to the validators corresponding to the operators.
// Update the validator content and return it to the end blocker's ctx.
func (k Keeper) CalcValidatorUpdatesByLSTTokenShares(ctx sdk.Context, nativeVotingPower *math.Int) ([]types.ValidatorUpdate, error) {
	logger := k.Logger(ctx)
	updates := make([]types.ValidatorUpdate, 0)

	// 1. Calculate total voting power
	totalLSTVotingPower, err := k.calcLSTVotingPower(ctx, nativeVotingPower)
	if err != nil {
		logger.Error("failed to calculate total voting power", "error", err)
		return updates, nil
	}

	// 2. Get operator list
	operatorList, exist := k.GetOperatorWeightedShareList(ctx)
	if !exist || len(operatorList.OperatorWeightedShares) == 0 {
		return updates, nil
	}

	// 3. Calculate total shares
	totalShares := math.ZeroInt()
	for _, operator := range operatorList.OperatorWeightedShares {
		totalShares = totalShares.Add(operator.WeightedShare)
	}

	if totalShares.IsZero() {
		logger.Info("total shares is zero, no LST voting power to allocate")
		return updates, nil
	}

	// 4. Calculate voting power for each operator
	for _, operator := range operatorList.OperatorWeightedShares {
		// voting power = total LST voting power * operator share / total shares
		votingPower := totalLSTVotingPower.Mul(operator.WeightedShare).Quo(totalShares)

		// 5. Get validator information
		valAddr, err := sdk.ValAddressFromBech32(operator.ValidatorAddr)
		if err != nil {
			logger.Error("failed to get validator address", "error", err, "validator", operator.ValidatorAddr)
			continue // Skip validator with error
		}

		validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			logger.Error("failed to get validator info", "error", err, "validator", operator.ValidatorAddr)
			continue // Skip validator with error
		}

		// 6. Get validator public key
		pubKey, err := validator.CmtConsPublicKey()
		if err != nil {
			logger.Error("failed to get validator public key", "error", err, "validator", operator.ValidatorAddr)
			continue // Skip validator with error
		}

		// 7. Get validator's native token voting power
		nativePower := sdk.TokensToConsensusPower(validator.BondedTokens(), k.stakingKeeper.PowerReduction(ctx))
		lstPower := votingPower.Int64()

		logger.Debug("validator voting power details",
			"operator", operator.OperatorAddress,
			"validator", validator.GetOperator(),
			"nativePower", nativePower,
			"lstPower", lstPower,
			"totalPower", nativePower+lstPower)

		// 8. Create validator update
		updates = append(updates, types.ValidatorUpdate{
			PubKey: pubKey,
			Power:  nativePower + lstPower,
		})
	}

	return updates, nil
}

// CalcNativeVotingPower calculates the total voting power of native validators
// totalNativeVotingPower = sum(validator.ConsensusPower(powerReduction))
func (k Keeper) CalcNativeVotingPower(ctx sdk.Context, nativeValidators []stakingtypes.Validator) (*math.Int, error) {
	powerReduction := k.stakingKeeper.PowerReduction(ctx)
	totalNativeVotingPower := math.ZeroInt()
	for _, validator := range nativeValidators {
		power := validator.ConsensusPower(powerReduction)
		totalNativeVotingPower = totalNativeVotingPower.Add(math.NewInt(power))
	}

	return &totalNativeVotingPower, nil
}

// calcLSTVotingPower calculates the total voting power of LST token
// lstVotingPower = nativeVotingPower * ratio.Numerator / ratio.Denominator
func (k Keeper) calcLSTVotingPower(ctx sdk.Context, nativeVotingPower *math.Int) (*math.Int, error) {
	// Get LST token voting power ratio
	ratio, exist := k.GetLSTVotingPowerRatio(ctx)
	if !exist {
		return nil, errors.New("LST voting power ratio not found")
	}

	//  Calculate total LST voting power
	// totalNativeVotingPower * ratio.Numerator / ratio.Denominator
	totalLSTVotingPower := nativeVotingPower.
		Mul(ratio.Numerator).
		Quo(ratio.Denominator)

	return &totalLSTVotingPower, nil
}
