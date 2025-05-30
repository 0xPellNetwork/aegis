package keeper

import (
	"context"
	"math"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/0xPellNetwork/aegis/pkg/crypto"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

// AddObserver adds an observer address to the observer set
func (k msgServer) AddObserver(goCtx context.Context, msg *types.MsgAddObserver) (*types.MsgAddObserverResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgAddObserverResponse{}, authoritytypes.ErrUnauthorized
	}

	pubkey, err := crypto.NewPubKey(msg.PellclientGranteePubkey)
	if err != nil {
		return &types.MsgAddObserverResponse{}, cosmoserrors.Wrap(sdkerrors.ErrInvalidPubKey, err.Error())
	}
	granteeAddress, err := crypto.GetAddressFromPubkeyString(msg.PellclientGranteePubkey)
	if err != nil {
		return &types.MsgAddObserverResponse{}, cosmoserrors.Wrap(sdkerrors.ErrInvalidPubKey, err.Error())
	}

	k.DisableInboundOnly(ctx)

	// AddNodeAccountOnly flag usage
	// True: adds observer into the Node Account list but returns without adding to the observer list
	// False: adds observer to the observer list, and not the node account list
	// Inbound is disabled in both cases and needs to be enabled manually using an admin TX
	if msg.AddNodeAccountOnly {
		pubkeySet := crypto.PubKeySet{Secp256k1: pubkey, Ed25519: ""}
		k.SetNodeAccount(ctx, types.NodeAccount{
			Operator:       msg.ObserverAddress,
			GranteeAddress: granteeAddress.String(),
			GranteePubkey:  &pubkeySet,
			NodeStatus:     types.NodeStatus_ACTIVE,
		})
		k.SetKeygen(ctx, types.Keygen{BlockNumber: math.MaxInt64})
		return &types.MsgAddObserverResponse{}, nil
	}

	k.AddObserverToSet(ctx, msg.ObserverAddress)
	observerSet, _ := k.GetObserverSet(ctx)

	k.SetLastObserverCount(ctx, &types.LastRelayerCount{Count: observerSet.LenUint()})
	EmitEventAddObserver(ctx, observerSet.LenUint(), msg.ObserverAddress, granteeAddress.String(), msg.PellclientGranteePubkey)

	return &types.MsgAddObserverResponse{}, nil
}
