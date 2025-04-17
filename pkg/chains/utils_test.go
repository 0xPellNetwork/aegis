package chains

import (
	"encoding/hex"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestHashToString(t *testing.T) {
	evmChainId := int64(5)
	unknownChainId := int64(3)
	mockEthBlockHash := []byte("0xc2339489a45f8976d45482ad6fa08751a1eae91f92d60645521ca0aff2422639")

	tests := []struct {
		name      string
		chainID   int64
		blockHash []byte
		expect    string
		wantErr   bool
	}{
		{"evm chain", evmChainId, mockEthBlockHash, hex.EncodeToString(mockEthBlockHash), false},
		{"unknown chain", unknownChainId, mockEthBlockHash, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HashToString(tt.chainID, tt.blockHash)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestStringToHash(t *testing.T) {
	evmChainId := int64(5)
	unknownChainId := int64(3)

	tests := []struct {
		name    string
		chainID int64
		hash    string
		expect  []byte
		wantErr bool
	}{
		{"evm chain", evmChainId, "95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5", ethcommon.HexToHash("95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5").Bytes(), false},
		{"unknown chain", unknownChainId, "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringToHash(tt.chainID, tt.hash)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestParseAddressAndData(t *testing.T) {
	expectedShortMsgResult, err := hex.DecodeString("1a2b3c4d5e6f708192a3b4c5d6e7f808")
	require.NoError(t, err)
	tests := []struct {
		name       string
		message    string
		expectAddr ethcommon.Address
		expectData []byte
		wantErr    bool
	}{
		{"valid msg", "95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5", ethcommon.HexToAddress("95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"), []byte{}, false},
		{"empty msg", "", ethcommon.Address{}, nil, false},
		{"invalid hex", "invalidHex", ethcommon.Address{}, nil, true},
		{"short msg", "1a2b3c4d5e6f708192a3b4c5d6e7f808", ethcommon.Address{}, expectedShortMsgResult, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, data, err := ParseAddressAndData(tt.message)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectAddr, addr)
				require.Equal(t, tt.expectData, data)
			}
		})
	}
}
