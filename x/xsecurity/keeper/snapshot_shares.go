package keeper

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	xsecuritytypes "github.com/pell-chain/pellcore/x/xsecurity/types"
)

// SnapshotSharesFromRestakingModuleByEpoch is a function to snapshot shares from restaking module by epoch
func (k Keeper) SnapshotSharesFromRestakingModuleByEpoch(ctx sdk.Context) error {
	// Query shares from restaking module
	allShares := k.restakingKeeper.GetAllShares(ctx)

	// Filter shares by operator list
	shares, err := k.FilterSharesInfoByOperatorList(ctx, allShares)
	if err != nil {
		k.Logger(ctx).Error("filter shares by operator list", "error", err)
		return err
	}

	// Calculate weighted shares
	weightedShares, err := k.CalcWeightedShares(ctx, shares)
	if err != nil {
		k.Logger(ctx).Error("calculate weighted shares", "error", err)
		return err
	}

	// Store the weighted shares
	k.SetOperatorWeightedShareList(ctx, weightedShares)
	return nil
}

// FilterSharesInfoByOperatorList is a function to filter shares by operator list
func (k Keeper) FilterSharesInfoByOperatorList(ctx sdk.Context, allShares []*restakingtypes.OperatorShares) ([]*restakingtypes.OperatorShares, error) {
	var shares []*restakingtypes.OperatorShares

	// Get operator registration list
	list, exist := k.GetOperatorRegistrationList(ctx)
	if !exist {
		return nil, errors.New("operator registration list not found")
	}

	// Filter shares by operator list
	for _, operator := range list.OperatorRegistrations {
		for _, share := range allShares {
			if operator.OperatorAddress == share.Operator {
				shares = append(shares, share)
			}
		}
	}
	return shares, nil
}

// CalcWeightedShares is a function to calculate weighted shares
func (k Keeper) CalcWeightedShares(ctx sdk.Context, shares []*restakingtypes.OperatorShares) (*xsecuritytypes.LSTOperatorWeightedShareList, error) {
	groupInfo, exist := k.GetGroupInfo(ctx)
	if !exist {
		return nil, errors.New("group info not found")
	}

	validatorList, exist := k.GetOperatorRegistrationList(ctx)
	if !exist {
		return nil, errors.New("operator registration list not found")
	}

	var weightedShares []*xsecuritytypes.LSTOperatorWeightedShare
	// Calculate weighted shares
	for _, share := range shares {
		multiplier, err := k.queryPoolMultiplier(groupInfo, share.Strategy)
		if err != nil {
			continue
		}

		validatorAddress, err := k.queryValidatorAddress(validatorList, share.Operator)
		if err != nil {
			continue
		}

		// weighted share = share.Shares * multiplier / WeightingDivisor
		weightedShare := share.Shares.Mul(sdkmath.NewIntFromUint64(multiplier)).Quo(xsecuritytypes.WeightingDivisor)

		// create LSTOperatorWeightedShare data
		data := xsecuritytypes.LSTOperatorWeightedShare{
			OperatorAddress: share.Operator,
			ValidatorAddr:   validatorAddress,
			WeightedShare:   weightedShare,
		}
		weightedShares = append(weightedShares, &data)
	}
	return &xsecuritytypes.LSTOperatorWeightedShareList{OperatorWeightedShares: weightedShares}, nil
}

// queryPoolMultiplier is a function to query pool multiplier, due to the restaking module shares info not contain the multiplier info
func (k Keeper) queryPoolMultiplier(groupInfo *xsecuritytypes.LSTGroupInfo, pool string) (uint64, error) {
	for _, poolInfo := range groupInfo.PoolParams {
		if poolInfo.Pool == pool {
			return poolInfo.Multiplier, nil
		}
	}

	return 0, errors.New("pool multiplier not found")
}

// queryValidatorAddress is a function to query validator address
func (k Keeper) queryValidatorAddress(validatorList *xsecuritytypes.LSTOperatorRegistrationList, operator string) (string, error) {
	for _, validator := range validatorList.OperatorRegistrations {
		if operator == validator.OperatorAddress {
			return validator.ValidatorAddress, nil
		}
	}

	return "", errors.New("validator address not found")
}

// GetChangedSharesByEpoch is a function to get changed shares by epoch
func (k Keeper) GetChangedSharesByEpoch(ctx sdk.Context) ([]*xsecuritytypes.LSTOperatorWeightedShare, error) {
	var shares []*xsecuritytypes.LSTOperatorWeightedShare

	// Get the last round operator weighted share list
	oldList, exist := k.GetLastRoundOperatorWeightedShareList(ctx)
	if !exist {
		oldList = &xsecuritytypes.LSTOperatorWeightedShareList{}
	}

	// Get the current operator weighted share list
	newList, exist := k.GetOperatorWeightedShareList(ctx)
	if !exist {
		newList = &xsecuritytypes.LSTOperatorWeightedShareList{}
	}

	// Compare the two lists and get the differences
	diffShares := DiffOperatorWeightedSharesWith(oldList, newList)
	if len(diffShares) == 0 {
		return shares, nil
	}

	// Store the new list as the last round shares
	k.SetLastRoundOperatorWeightedShareList(ctx, newList)

	return diffShares, nil
}

// DiffOperatorWeightedSharesWith compares two operator weighted share lists and returns the differences
func DiffOperatorWeightedSharesWith(
	oldList,
	newList *xsecuritytypes.LSTOperatorWeightedShareList,
) []*xsecuritytypes.LSTOperatorWeightedShare {
	// diffShares will store all items that are either added, modified, or deleted.
	var diffShares []*xsecuritytypes.LSTOperatorWeightedShare

	// Create a map from the old list for quick lookup by operator address
	oldSharesMap := make(map[string]*xsecuritytypes.LSTOperatorWeightedShare)
	for _, oldShare := range oldList.OperatorWeightedShares {
		oldSharesMap[oldShare.OperatorAddress] = oldShare
	}

	// Create a map from the new list for quick lookup by operator address
	newSharesMap := make(map[string]*xsecuritytypes.LSTOperatorWeightedShare)
	for _, newShare := range newList.OperatorWeightedShares {
		newSharesMap[newShare.OperatorAddress] = newShare
	}

	// 1. Check for added or modified operators in newList
	for _, newShare := range newList.OperatorWeightedShares {
		oldShare, exists := oldSharesMap[newShare.OperatorAddress]
		if !exists {
			// This operator did not exist in oldList: "Added"
			diffShares = append(diffShares, newShare)
		} else if !oldShare.WeightedShare.Equal(newShare.WeightedShare) {
			// The operator exists but the WeightedShare is different: "Modified"
			diffShares = append(diffShares, newShare)
		}
	}

	// 2. Check for deleted operators (exists in oldList but not in newList)
	for _, oldShare := range oldList.OperatorWeightedShares {
		if _, exists := newSharesMap[oldShare.OperatorAddress]; !exists {
			// This operator no longer appears in newList: "Deleted"
			// We set WeightedShare to zero to indicate removal
			diffShares = append(diffShares, &xsecuritytypes.LSTOperatorWeightedShare{
				OperatorAddress: oldShare.OperatorAddress,
				WeightedShare:   sdkmath.ZeroInt(), // or whatever zero type is appropriate
			})
		}
	}

	return diffShares
}
