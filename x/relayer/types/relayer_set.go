package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

func (m *RelayerSet) Len() int {
	return len(m.RelayerList)
}

func (m *RelayerSet) LenUint() uint64 {
	return uint64(len(m.RelayerList))
}

// Validate observer mapper contains an existing chain
func (m *RelayerSet) Validate() error {
	for _, observerAddress := range m.RelayerList {
		_, err := sdk.AccAddressFromBech32(observerAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckReceiveStatus(status chains.ReceiveStatus) error {
	switch status {
	case chains.ReceiveStatus_SUCCESS:
		return nil
	case chains.ReceiveStatus_FAILED:
		return nil
	default:
		return ErrInvalidStatus
	}
}
