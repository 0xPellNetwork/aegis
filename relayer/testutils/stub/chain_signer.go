package stub

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/outtxprocessor"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// ----------------------------------------------------------------------------
// EVMSigner
// ----------------------------------------------------------------------------
var _ interfaces.ChainSigner = (*EVMSigner)(nil)

// EVMSigner is a mock of evm chain signer for testing
type EVMSigner struct {
	Chain                chains.Chain
	PellConnectorAddress ethcommon.Address
}

func NewEVMSigner(
	chain chains.Chain,
	pellConnectorAddress ethcommon.Address,
) *EVMSigner {
	return &EVMSigner{
		Chain:                chain,
		PellConnectorAddress: pellConnectorAddress,
	}
}

func (s *EVMSigner) TryProcessOutTx(
	_ context.Context,
	_ *xmsgtypes.Xmsg,
	_ *outtxprocessor.Processor,
	_ string,
	_ interfaces.ChainClient,
	_ interfaces.PellCoreBridger,
	_ uint64,
) {
}

func (s *EVMSigner) SetPellConnectorAddress(address ethcommon.Address) {
	s.PellConnectorAddress = address
}

func (s *EVMSigner) GetPellConnectorAddress() ethcommon.Address {
	return s.PellConnectorAddress
}

// ----------------------------------------------------------------------------
// BTCSigner
// ----------------------------------------------------------------------------
var _ interfaces.ChainSigner = (*BTCSigner)(nil)

// BTCSigner is a mock of bitcoin chain signer for testing
type BTCSigner struct {
}

func NewBTCSigner() *BTCSigner {
	return &BTCSigner{}
}

func (s *BTCSigner) TryProcessOutTx(
	_ context.Context,
	_ *xmsgtypes.Xmsg,
	_ *outtxprocessor.Processor,
	_ string,
	_ interfaces.ChainClient,
	_ interfaces.PellCoreBridger,
	_ uint64,
) {
}

func (s *BTCSigner) SetPellConnectorAddress(_ ethcommon.Address) {
}

func (s *BTCSigner) GetPellConnectorAddress() ethcommon.Address {
	return ethcommon.Address{}
}
