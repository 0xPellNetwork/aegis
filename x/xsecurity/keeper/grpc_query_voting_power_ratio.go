package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// QueryVotingPowerRatio queries the voting power ratio of the LST
func (k Keeper) QueryVotingPowerRatio(goCtx context.Context, req *types.QueryVotingPowerRatioRequest) (*types.QueryVotingPowerRatioResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	info, exist := k.GetLSTVotingPowerRatio(ctx)
	if !exist {
		return nil, errors.New("data not found")
	}

	return &types.QueryVotingPowerRatioResponse{
		Numerator:   info.Numerator.Uint64(),
		Denominator: info.Denominator.Uint64(),
	}, nil
}
