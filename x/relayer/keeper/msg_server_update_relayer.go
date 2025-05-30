package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// UpdateObserver handles updating an observer address
// Authorized: admin policy (admin update), old observer address (if the
// reason is that the observer was tombstoned).
func (k msgServer) UpdateObserver(goCtx context.Context, msg *types.MsgUpdateObserver) (*types.MsgUpdateObserverResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ok, err := k.CheckUpdateReason(ctx, msg)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrUpdateObserver, err.Error())
	}
	if !ok {
		return nil, errorsmod.Wrap(types.ErrUpdateObserver, fmt.Sprintf("Unable to update observer with update reason : %s", msg.UpdateReason))
	}

	// We do not use IsNonTombstonedObserver here because we want to allow tombstoned observers to be updated
	if !k.IsAddressPartOfObserverSet(ctx, msg.OldObserverAddress) {
		return nil, errorsmod.Wrap(types.ErrNotObserver, fmt.Sprintf("Observer address is not authorized : %s", msg.OldObserverAddress))
	}

	err = k.IsValidator(ctx, msg.NewObserverAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrUpdateObserver, err.Error())
	}

	// Update all mappers so that ballots can be created for the new observer address
	err = k.UpdateObserverAddress(ctx, msg.OldObserverAddress, msg.NewObserverAddress)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrUpdateObserver, err.Error())
	}

	// Update the node account with the new operator address
	nodeAccount, found := k.GetNodeAccount(ctx, msg.OldObserverAddress)
	if !found {
		return nil, errorsmod.Wrap(types.ErrNodeAccountNotFound, fmt.Sprintf("Observer node account not found : %s", msg.OldObserverAddress))
	}
	newNodeAccount := nodeAccount
	newNodeAccount.Operator = msg.NewObserverAddress

	// Remove an old node account, so that number of node accounts remains the same as the number of observers in the system
	k.RemoveNodeAccount(ctx, msg.OldObserverAddress)
	k.SetNodeAccount(ctx, newNodeAccount)

	// Check LastBlockObserver count just to be safe
	observerSet, found := k.GetObserverSet(ctx)
	if !found {
		return nil, errorsmod.Wrap(types.ErrObserverSetNotFound, fmt.Sprintf("Observer set not found"))
	}
	totalObserverCountCurrentBlock := observerSet.LenUint()
	lastBlockCount, found := k.GetLastObserverCount(ctx)
	if !found {
		return nil, errorsmod.Wrap(types.ErrLastObserverCountNotFound, fmt.Sprintf("Observer count not found"))
	}
	if lastBlockCount.Count != totalObserverCountCurrentBlock {
		return nil, errorsmod.Wrap(types.ErrUpdateObserver, fmt.Sprintf("Observer count mismatch"))
	}
	return &types.MsgUpdateObserverResponse{}, nil
}

func (k Keeper) CheckUpdateReason(ctx sdk.Context, msg *types.MsgUpdateObserver) (bool, error) {
	switch msg.UpdateReason {
	case types.RelayerUpdateReason_TOMBSTONED:
		{
			if msg.Signer != msg.OldObserverAddress {
				return false, errorsmod.Wrap(types.ErrUpdateObserver, fmt.Sprintf("Creator address and old observer address need to be same for updating tombstoned observer"))
			}
			return k.IsOperatorTombstoned(ctx, msg.Signer)
		}
	case types.RelayerUpdateReason_ADMIN_UPDATE:
		{
			// Operational policy is required to update an observer for admin update
			if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
				return false, authoritytypes.ErrUnauthorized
			}
			return true, nil
		}
	}
	return false, nil
}

func UpdateRelayerList(list []string, oldObserverAddresss, newObserverAddress string) {
	for i, observer := range list {
		if observer == oldObserverAddresss {
			list[i] = newObserverAddress
		}
	}
}
