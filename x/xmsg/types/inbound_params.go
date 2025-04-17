package types

import (
	"fmt"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

func (m InboundTxParams) Validate() error {
	if m.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}

	if _, exist := chains.GetChainByChainId(m.SenderChainId); !exist {
		return fmt.Errorf("invalid sender chain id %d", m.SenderChainId)
	}

	// Disabled checks
	// TODO: Improve the checks, move the validation call to a new place and reenable
	//if err := ValidateAddressForChain(m.Sender, m.SenderChainId) err != nil {
	//	return err
	//}
	//if m.TxOrigin != "" {
	//	errTxOrigin := ValidateAddressForChain(m.TxOrigin, m.SenderChainId)
	//	if errTxOrigin != nil {
	//		return errTxOrigin
	//	}
	//}
	//if err := ValidateHashForChain(m.ObservedHash, m.SenderChainId); err != nil {
	//	return errors.Wrap(err, "invalid inbound tx observed hash")
	//}
	//if m.BallotIndex != "" {
	//	if err := ValidateXmsgIndices(m.BallotIndex); err != nil {
	//		return errors.Wrap(err, "invalid inbound tx ballot index")
	//	}
	//}
	return nil
}
