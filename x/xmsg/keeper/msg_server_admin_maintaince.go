package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// TODO: remove this after the next upgrade
// InboundTxMaintenance handles the maintenance of inbound transaction data
// It deletes block proofs and their associated ballot records within a specified block height range
func (k msgServer) InboundTxMaintenance(ctx context.Context, msg *types.MsgInboundTxMaintenance) (*types.MsgInboundTxMaintenanceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Check if the signer has admin permissions
	if !k.GetAuthorityKeeper().IsAuthorized(sdkCtx, msg.Signer, authoritytypes.PolicyType_GROUP_ADMIN) {
		return &types.MsgInboundTxMaintenanceResponse{}, authoritytypes.ErrUnauthorized
	}

	// Iterate through the specified block height range
	for i := msg.FromBlockHeight; i <= msg.ToBlockHeight; i++ {
		// Get block proof for the current height
		bp, exist := k.GetBlockProof(sdkCtx, uint64(msg.ChainId), i)
		if !exist {
			continue
		}

		// Delete the block proof
		k.DeleteBlockProof(sdkCtx, uint64(msg.ChainId), i)

		// Generate ballot index from block proof
		ballotIndex := ballotIndexByBlockProof(&bp)

		// Delete the corresponding ballot record
		k.relayerKeeper.DeleteBallot(sdkCtx, ballotIndex)
	}

	return &types.MsgInboundTxMaintenanceResponse{}, nil
}

// ballotIndexByBlockProof generates a ballot index from a block proof
// by creating a temporary MsgVoteInboundBlock and calculating its digest
func ballotIndexByBlockProof(bp *types.BlockProof) string {
	msg := &types.MsgVoteInboundBlock{
		BlockProof: bp,
	}

	return msg.Digest()
}
