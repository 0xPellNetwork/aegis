package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// Max key1 value (occupy 6 bytes = 48 bits)
	MaxKey1 = uint64(1<<48 - 1)
)

// GenerateSystemTxId generates a system transaction hash with business type and two keys
// Format: 0x{system_tx_type(1 byte)}{key1(6 bytes)}{key2(1 byte)}0000000000000000000000000000000000000000
func GenerateSystemTxId(systemTxType uint8, key1 uint64, key2 uint8) string {
	if key1 > MaxKey1 {
		panic("key1 exceeds maximum value")
	}

	// Combine all fields:
	// - business type in highest byte (bits 56-63)
	// - key1 in middle bytes (bits 8-55)
	// - key2 in lowest byte (bits 0-7)
	combined := (uint64(systemTxType) << 56) | (key1 << 8) | uint64(key2)

	// Create the full 32 bytes (64 chars) hash
	// First 8 bytes for our data, remaining 24 bytes are zeros
	return fmt.Sprintf("0x%016x%048s", combined, "0")
}

// ParseSystemTxId parses a system transaction hash to get business type and two keys
func ParseSystemTxId(hash string) (systemTxType uint8, key1 uint64, key2 uint8, err error) {
	// Remove "0x" prefix if present
	hash = strings.TrimPrefix(hash, "0x")

	// Check if hash length is correct (64 characters for 32 bytes)
	if len(hash) != 64 {
		return 0, 0, 0, fmt.Errorf("invalid hash length")
	}

	// Check if remaining bytes are all zeros
	remainingBytes := hash[16:]
	for _, c := range remainingBytes {
		if c != '0' {
			return 0, 0, 0, fmt.Errorf("invalid hash format: remaining bytes must be zeros")
		}
	}

	// Parse the first 16 characters (8 bytes) where we store our data
	combined, err := strconv.ParseUint(hash[:16], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse hash: %v", err)
	}

	// Extract fields:
	systemTxType = uint8(combined >> 56) // highest byte
	key1 = (combined >> 8) & MaxKey1     // middle 6 bytes
	key2 = uint8(combined & 0xFF)        // lowest byte

	return systemTxType, key1, key2, nil
}
