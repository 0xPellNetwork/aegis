package sample

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	"github.com/0xPellNetwork/aegis/pkg/proofs/ethereum"
	"github.com/0xPellNetwork/aegis/pkg/testdata"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
)

func BlockHeader(blockHash []byte) proofs.BlockHeader {
	return proofs.BlockHeader{
		Height:     42,
		Hash:       blockHash,
		ParentHash: Hash().Bytes(),
		ChainId:    42,
		Header:     proofs.HeaderData{},
	}
}

func ChainState(chainID int64) lightclienttypes.ChainState {
	return lightclienttypes.ChainState{
		ChainId:         chainID,
		LatestHeight:    42,
		EarliestHeight:  42,
		LatestBlockHash: Hash().Bytes(),
	}
}

func VerificationFlags() lightclienttypes.VerificationFlags {
	return lightclienttypes.VerificationFlags{
		EthTypeChainEnabled: true,
		BtcTypeChainEnabled: true,
	}
}

// Proof generates a proof and block header
// returns the proof, block header, block hash, tx index, chain id, and tx hash
func Proof(t *testing.T) (*proofs.Proof, proofs.BlockHeader, string, int64, int64, ethcommon.Hash) {
	header, err := testdata.ReadEthHeader()
	require.NoError(t, err)
	b, err := rlp.EncodeToBytes(&header)
	require.NoError(t, err)

	var txs ethtypes.Transactions
	for i := 0; i < testdata.TxsCount; i++ {
		tx, err := testdata.ReadEthTx(i)
		require.NoError(t, err)
		txs = append(txs, &tx)
	}
	txsTree := ethereum.NewTrie(txs)

	// choose 2 as the index of the tx to prove
	txIndex := 2
	proof, err := txsTree.GenerateProof(txIndex)
	require.NoError(t, err)

	chainID := chains.SepoliaChain().Id
	ethProof := proofs.NewEthereumProof(proof)
	ethHeader := proofs.NewEthereumHeader(b)
	blockHeader := proofs.BlockHeader{
		Height:     header.Number.Int64(),
		Hash:       header.Hash().Bytes(),
		ParentHash: header.ParentHash.Bytes(),
		ChainId:    chainID,
		Header:     ethHeader,
	}
	txHash := txs[txIndex].Hash()
	return ethProof, blockHeader, header.Hash().Hex(), int64(txIndex), chainID, txHash
}
