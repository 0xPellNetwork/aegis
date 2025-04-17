package proofs

import (
	"bytes"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestTrueEthereumHeader(t *testing.T) {
	var header ethtypes.Header
	// read file into a byte slice
	file, err := os.Open("../testdata/eth_header_18495266.json")
	require.NoError(t, err)
	defer file.Close()
	headerBytes := make([]byte, 4096)
	n, err := file.Read(headerBytes)
	require.NoError(t, err)
	err = header.UnmarshalJSON(headerBytes[:n])
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = header.EncodeRLP(&buffer)
	require.NoError(t, err)

	headerData := NewEthereumHeader(buffer.Bytes())
	err = headerData.Validate(header.Hash().Bytes(), 1, 18495266)
	require.NoError(t, err)

	parentHash, err := headerData.ParentHash()
	require.NoError(t, err)
	require.Equal(t, header.ParentHash.Bytes(), parentHash)

	err = headerData.ValidateTimestamp(time.Now())
	require.NoError(t, err)
}

func TestFalseEthereumHeader(t *testing.T) {
	var header ethtypes.Header
	// read file into a byte slice
	file, err := os.Open("../testdata/eth_header_18495266.json")
	require.NoError(t, err)
	defer file.Close()
	headerBytes := make([]byte, 4096)
	n, err := file.Read(headerBytes)
	require.NoError(t, err)
	err = header.UnmarshalJSON(headerBytes[:n])
	require.NoError(t, err)
	hash := header.Hash()
	header.Number = big.NewInt(18495267)
	var buffer bytes.Buffer
	err = header.EncodeRLP(&buffer)
	require.NoError(t, err)

	headerData := NewEthereumHeader(buffer.Bytes())
	err = headerData.Validate(hash.Bytes(), 1, 18495267)
	require.Error(t, err)
}

func TestNonExistentHeaderType(t *testing.T) {
	headerData := HeaderData{}

	err := headerData.Validate([]byte{}, 18332, 0)
	require.EqualError(t, err, "unrecognized header type")

	parentHash, err := headerData.ParentHash()
	require.EqualError(t, err, "unrecognized header type")
	require.Nil(t, parentHash)

	err = headerData.ValidateTimestamp(time.Now())
	require.ErrorContains(t, err, "unrecognized header type")
}

func createBTCClient(t *testing.T) *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:18332",
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true,
		DisableTLS:   true,
		Params:       "testnet3",
	}
	client, err := rpcclient.New(connCfg, nil)
	require.NoError(t, err)
	return client
}

func copyHeader(header *wire.BlockHeader) *wire.BlockHeader {
	copyHeader := &wire.BlockHeader{
		Version:    header.Version,
		PrevBlock:  chainhash.Hash{},
		MerkleRoot: chainhash.Hash{},
		Timestamp:  header.Timestamp,
		Bits:       header.Bits,
		Nonce:      header.Nonce,
	}
	copy(copyHeader.PrevBlock[:], header.PrevBlock[:])
	copy(copyHeader.MerkleRoot[:], header.MerkleRoot[:])

	return copyHeader
}

func marshalHeader(t *testing.T, header *wire.BlockHeader) []byte {
	var headerBuf bytes.Buffer
	err := header.Serialize(&headerBuf)
	require.NoError(t, err)
	return headerBuf.Bytes()
}

func unmarshalHeader(t *testing.T, headerBytes []byte) *wire.BlockHeader {
	var header wire.BlockHeader
	err := header.Deserialize(bytes.NewReader(headerBytes))
	require.NoError(t, err)
	return &header
}
