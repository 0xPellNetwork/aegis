package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemTxId(t *testing.T) {
	const SystemTxTypeSyncDelegationShares = uint8(1)
	testCases := []struct {
		name         string
		systemTxType uint8
		key1         uint64
		key2         uint8
		expectPanic  bool
		expectHash   string
	}{
		{
			name:         "normal case",
			systemTxType: SystemTxTypeSyncDelegationShares,
			key1:         0x123456,
			key2:         0xAB,
			expectHash:   "0x01000000123456ab000000000000000000000000000000000000000000000000",
		},
		{
			name:         "all zeros",
			systemTxType: 0,
			key1:         0,
			key2:         0,
			expectHash:   "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:         "max values within limits",
			systemTxType: 0xFF,
			key1:         MaxKey1,
			key2:         0xFF,
			expectHash:   "0xffffffffffffffff000000000000000000000000000000000000000000000000",
		},
		{
			name:         "exceed max key1",
			systemTxType: SystemTxTypeSyncDelegationShares,
			key1:         MaxKey1 + 1,
			key2:         0,
			expectPanic:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				require.Panics(t, func() {
					GenerateSystemTxId(tc.systemTxType, tc.key1, tc.key2)
				})
				return
			}

			hash := GenerateSystemTxId(tc.systemTxType, tc.key1, tc.key2)
			require.Equal(t, tc.expectHash, hash)

			// Verify parsing
			systemTxType, key1, key2, err := ParseSystemTxId(hash)
			require.NoError(t, err)
			require.Equal(t, tc.systemTxType, systemTxType)
			require.Equal(t, tc.key1, key1)
			require.Equal(t, tc.key2, key2)
		})
	}
}

func TestParseSystemTxHash_Errors(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "invalid hash length",
			hash:    "0x123",
			wantErr: true,
		},
		{
			name:    "non-zero remaining bytes",
			hash:    "0x0100123456ab0000000000000000000000000000000000000000000000000001",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			hash:    "0xXX00123456ab0000000000000000000000000000000000000000000000000000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing hash: %s", tt.hash)
			_, _, _, err := ParseSystemTxId(tt.hash)
			if tt.wantErr {
				require.Error(t, err, "Expected an error for hash: %s", tt.hash)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidHashWithoutPrefix(t *testing.T) {
	hash := "0100123456ab0000000000000000000000000000000000000000000000000000"
	systemTxType, key1, key2, err := ParseSystemTxId(hash)
	require.NoError(t, err)

	// Generate hash with the parsed values and compare
	generatedHash := GenerateSystemTxId(systemTxType, key1, key2)
	require.Equal(t, "0x"+hash, generatedHash)
}
