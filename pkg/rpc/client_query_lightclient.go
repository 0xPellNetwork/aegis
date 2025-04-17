package rpc

import (
	"context"

	"cosmossdk.io/errors"

	"github.com/pell-chain/pellcore/pkg/proofs"
	"github.com/pell-chain/pellcore/x/lightclient/types"
)

// GetVerificationFlags returns the enabled chains for block headers
func (c *Clients) GetVerificationFlags(ctx context.Context) (types.VerificationFlags, error) {
	resp, err := c.Lightclient.VerificationFlags(ctx, &types.QueryVerificationFlagsRequest{})
	if err != nil {
		return types.VerificationFlags{}, err
	}
	return resp.VerificationFlags, nil
}

// GetBlockHeaderChainState returns the block header chain state
func (c *Clients) GetBlockHeaderChainState(ctx context.Context, chainID int64) (*types.QueryChainStateResponse, error) {
	in := &types.QueryGetChainStateRequest{ChainId: chainID}

	resp, err := c.Lightclient.ChainState(ctx, in)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chain state")
	}
	return resp, nil
}

func (c *Clients) Prove(ctx context.Context, blockHash string, txHash string, txIndex int64, proof *proofs.Proof, chainID int64) (bool, error) {
	in := &types.QueryProveRequest{
		BlockHash: blockHash,
		TxIndex:   txIndex,
		Proof:     proof,
		ChainId:   chainID,
		TxHash:    txHash,
	}

	resp, err := c.Lightclient.Prove(ctx, in)
	if err != nil {
		return false, errors.Wrap(err, "failed to prove")
	}
	return resp.Valid, nil
}
