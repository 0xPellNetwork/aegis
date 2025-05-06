package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// UpdateTssAddress updates the TSS address.
func (k msgServer) UpdateTssAddress(goCtx context.Context, msg *types.MsgUpdateTssAddress) (*types.MsgUpdateTssAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if authorized
	// TODO : Add a new policy type for updating the TSS address
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return nil, errorsmod.Wrap(authoritytypes.ErrUnauthorized, "Update can only be executed by the correct policy account")
	}

	currentTss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, errorsmod.Wrap(types.ErrUnableToUpdateTss, "cannot find current TSS")
	}
	if currentTss.TssPubkey == msg.TssPubkey {
		return nil, errorsmod.Wrap(types.ErrUnableToUpdateTss, "no new tss address has been generated")
	}

	tss, ok := k.relayerKeeper.CheckIfTssPubkeyHasBeenGenerated(ctx, msg.TssPubkey)
	if !ok {
		return nil, errorsmod.Wrap(types.ErrUnableToUpdateTss, "tss pubkey has not been generated")
	}

	tssMigrators := k.relayerKeeper.GetAllTssFundMigrators(ctx)
	// Each connected chain should have its own tss migrator
	if len(k.relayerKeeper.GetSupportedChains(ctx)) != len(tssMigrators) {
		return nil, errorsmod.Wrap(types.ErrUnableToUpdateTss, "cannot update tss address not enough migrations have been created and completed")
	}

	// GetAllTssFundMigrators would return the migrators created for the current migration
	// if any of the migrations is still pending we should not allow the tss address to be updated
	// we can wait for all migrations to complete before updating; this includes btc and eth chains.
	for _, tssMigrator := range tssMigrators {
		migratorTx, found := k.GetXmsg(ctx, tssMigrator.MigrationXmsgIndex)
		if !found {
			return nil, errorsmod.Wrap(types.ErrUnableToUpdateTss, "migration cross chain tx not found")
		}
		if migratorTx.XmsgStatus.Status != types.XmsgStatus_OUTBOUND_MINED {
			return nil, errorsmod.Wrapf(types.ErrUnableToUpdateTss,
				"cannot update tss address while there are pending migrations , current status of migration xmsg : %s ", migratorTx.XmsgStatus.Status.String())
		}

	}

	k.GetRelayerKeeper().SetTssAndUpdateNonce(ctx, tss)

	// Remove all migrators once the tss address has been updated successfully,
	// A new set of migrators will be created when the next migration is triggered
	k.relayerKeeper.RemoveAllExistingMigrators(ctx)

	return &types.MsgUpdateTssAddressResponse{}, nil
}
