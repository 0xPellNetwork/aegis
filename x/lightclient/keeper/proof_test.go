package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/proofs"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/lightclient/types"
)

func TestKeeper_VerifyProof(t *testing.T) {
	t.Run("should error if verification flags not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		_, err := k.VerifyProof(ctx, &proofs.Proof{}, chains.SepoliaChain().Id, sample.Hash().String(), 1)
		require.ErrorIs(t, err, types.ErrVerificationFlagsNotFound)
	})

	t.Run("should error if verification not enabled for evm chain", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: false,
			BtcTypeChainEnabled: true,
		})

		_, err := k.VerifyProof(ctx, &proofs.Proof{}, chains.SepoliaChain().Id, sample.Hash().String(), 1)
		require.ErrorIs(t, err, types.ErrBlockHeaderVerificationDisabled)
	})

	t.Run("should error if block header-based verification not supported", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: true,
			BtcTypeChainEnabled: true,
		})

		_, err := k.VerifyProof(ctx, &proofs.Proof{}, 101, sample.Hash().String(), 1)
		require.ErrorIs(t, err, types.ErrChainNotSupported)
	})

	t.Run("should error if block header not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: true,
			BtcTypeChainEnabled: true,
		})

		_, err := k.VerifyProof(ctx, &proofs.Proof{}, chains.SepoliaChain().Id, sample.Hash().String(), 1)
		require.ErrorIs(t, err, types.ErrBlockHeaderNotFound)
	})

	t.Run("should fail if proof can't be verified", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		proof, blockHeader, blockHash, txIndex, chainID, _ := sample.Proof(t)

		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: true,
			BtcTypeChainEnabled: true,
		})

		k.SetBlockHeader(ctx, blockHeader)

		// providing wrong tx index
		_, err := k.VerifyProof(ctx, proof, chainID, blockHash, txIndex+1)
		require.ErrorIs(t, err, types.ErrProofVerificationFailed)
	})

	t.Run("can verify a proof", func(t *testing.T) {
		k, ctx, _, _ := keepertest.LightclientKeeper(t)

		proof, blockHeader, blockHash, txIndex, chainID, _ := sample.Proof(t)

		k.SetVerificationFlags(ctx, types.VerificationFlags{
			EthTypeChainEnabled: true,
			BtcTypeChainEnabled: true,
		})

		k.SetBlockHeader(ctx, blockHeader)

		txBytes, err := k.VerifyProof(ctx, proof, chainID, blockHash, txIndex)
		require.NoError(t, err)
		require.NotNil(t, txBytes)
	})
}
