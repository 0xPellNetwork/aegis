package keeper

import (
	cosmoserror "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

// VerifyProof verifies the merkle proof for a given chain and block header
// It returns the transaction bytes if the proof is valid
func (k Keeper) VerifyProof(ctx sdk.Context, proof *proofs.Proof, chainID int64, blockHash string, txIndex int64) ([]byte, error) {
	// check verification flags are set
	if err := k.CheckVerificationFlagsEnabled(ctx, chainID); err != nil {
		return nil, err
	}

	// get block header from the store
	hashBytes, err := chains.StringToHash(chainID, blockHash)
	if err != nil {
		return nil, cosmoserror.Wrapf(types.ErrInvalidBlockHash, "block hash %s conversion failed %s", blockHash, err.Error())
	}
	res, found := k.GetBlockHeader(ctx, hashBytes)
	if !found {
		return nil, cosmoserror.Wrapf(types.ErrBlockHeaderNotFound, "block header not found %s", blockHash)
	}

	// verify merkle proof
	txBytes, err := proof.Verify(res.Header, int(txIndex))
	if err != nil {
		return nil, cosmoserror.Wrapf(types.ErrProofVerificationFailed, "failed to verify merkle proof: %s", err.Error())
	}
	return txBytes, nil
}
