package types

import (
	"fmt"
	"strconv"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

func (m OutboundTxParams) GetGasPrice() (uint64, error) {
	gasPrice, err := strconv.ParseUint(m.OutboundTxGasPrice, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse xmsg gas price %s: %s", m.OutboundTxGasPrice, err.Error())
	}

	return gasPrice, nil
}

func (m OutboundTxParams) Validate() error {
	if m.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty")
	}
	if _, exist := chains.GetChainByChainId(m.ReceiverChainId); !exist {
		return fmt.Errorf("invalid receiver chain id %d", m.ReceiverChainId)
	}

	// Disabled checks
	// TODO: Improve the checks, move the validation call to a new place and reenable
	//if err := ValidateAddressForChain(m.Receiver, m.ReceiverChainId); err != nil {
	//	return err
	//}
	//if m.BallotIndex != "" {
	//
	//	if err := ValidateXmsgIndices(m.BallotIndex); err != nil {
	//		return errors.Wrap(err, "invalid outbound tx ballot index")
	//	}
	//}
	//if m.Hash != "" {
	//	if err := ValidateHashForChain(m.Hash, m.ReceiverChainId); err != nil {
	//		return errors.Wrap(err, "invalid outbound tx hash")
	//	}
	//}

	return nil
}
