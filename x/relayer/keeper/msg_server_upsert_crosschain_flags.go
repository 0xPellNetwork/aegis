package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// UpdateCrosschainFlags updates the crosschain related flags.
//
// Aurthorized: admin policy group 1 (except enabling/disabled
// inbounds/outbounds and gas price increase), admin policy group 2 (all).
func (k msgServer) UpsertCrosschainFlags(goCtx context.Context, msg *types.MsgUpsertCrosschainFlags) (*types.MsgUpsertCrosschainFlagsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, msg.GetRequiredPolicyType()) {
		return &types.MsgUpsertCrosschainFlagsResponse{}, authoritytypes.ErrUnauthorized
	}

	// check if the value exists
	flags, isFound := k.GetCrosschainFlags(ctx)
	if !isFound {
		flags = *types.DefaultCrosschainFlags()
	}

	// update values
	flags.IsInboundEnabled = msg.IsInboundEnabled
	flags.IsOutboundEnabled = msg.IsOutboundEnabled

	if msg.GasPriceIncreaseFlags != nil {
		flags.GasPriceIncreaseFlags = msg.GasPriceIncreaseFlags
	}

	if msg.BlockHeaderVerificationFlags != nil {
		flags.BlockHeaderVerificationFlags = msg.BlockHeaderVerificationFlags
	}

	k.SetCrosschainFlags(ctx, flags)

	err := ctx.EventManager().EmitTypedEvents(&types.EventCrosschainFlagsUpdated{
		MsgTypeUrl:                   sdk.MsgTypeURL(&types.MsgUpsertCrosschainFlags{}),
		IsInboundEnabled:             msg.IsInboundEnabled,
		IsOutboundEnabled:            msg.IsOutboundEnabled,
		GasPriceIncreaseFlags:        msg.GasPriceIncreaseFlags,
		BlockHeaderVerificationFlags: msg.BlockHeaderVerificationFlags,
		Signer:                       msg.Signer,
	})
	if err != nil {
		ctx.Logger().Error("Error emitting EventCrosschainFlagsUpdated :", err)
	}

	return &types.MsgUpsertCrosschainFlagsResponse{}, nil
}
