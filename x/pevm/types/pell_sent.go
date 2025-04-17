package types

import (
	"errors"
	"fmt"
	"math/big"
)

const PellSentDefaultDestinationGasLimit uint64 = 1000000

// PellSentParamType is an enum-like type for Pell events.
type PellSentParamType uint8

const (
	ReceiveCall PellSentParamType = iota
	RevertableCall
	Transfer
)

// String returns a string representation of PellSentParamType.
func (p PellSentParamType) String() string {
	switch p {
	case ReceiveCall:
		return "0"
	case RevertableCall:
		return "1"
	case Transfer:
		return "2"
	default:
		return "Unknown"
	}
}

// PellSentParamTypeFromString converts a string to PellSentParamType.
func PellSentParamTypeFromString(s string) (PellSentParamType, error) {
	switch s {
	case "0":
		return ReceiveCall, nil
	case "1":
		return RevertableCall, nil
	case "2":
		return Transfer, nil
	default:
		return 0, fmt.Errorf("invalid PellSentParamType string: %s", s)
	}
}

// ToPellSentParamType converts the given []byte to PellSentParamType.
// The bytes are interpreted as a big-endian integer.
func ToPellSentParamType(b []byte) (PellSentParamType, error) {
	// If the byte slice is empty, we can't parse a valid number
	if len(b) == 0 {
		return 0, errors.New("invalid PellSentParamType: empty bytes")
	}

	// Convert bytes to big.Int
	intVal := new(big.Int).SetBytes(b)
	v := intVal.Uint64()

	// Map numeric value to PellSentParamType
	switch v {
	case 0:
		return ReceiveCall, nil
	case 1:
		return RevertableCall, nil
	case 2:
		return Transfer, nil
	default:
		return 0, fmt.Errorf("invalid PellSentParamType: %d", v)
	}
}

// MethodName returns the corresponding method name for the PellSentParamType.
func (p PellSentParamType) MethodName() (string, error) {
	switch p {
	case ReceiveCall:
		return "receiveCall", nil
	case RevertableCall:
		return "onReceive", nil
	case Transfer:
		return "", nil
	default:
		return "", fmt.Errorf("invalid PellSentParamType: %d", p)
	}
}
