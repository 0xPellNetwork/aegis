package types

import (
	"encoding/hex"
)

// BytesToEthHex converts a byte slice to a ethereum hex string with a leading "0x" prefix.
func BytesToEthHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}
