package stub

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/pell-chain/pellcore/relayer/chains/interfaces"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// ----------------------------------------------------------------------------
// EVMClient
// ----------------------------------------------------------------------------
var _ interfaces.ChainClient = (*EVMClient)(nil)

// EVMClient is a mock of evm chain client for testing
type EVMClient struct {
	ChainParams observertypes.ChainParams
}

func NewEVMClient(chainParams *observertypes.ChainParams) *EVMClient {
	return &EVMClient{
		ChainParams: *chainParams,
	}
}

func (s *EVMClient) Start(ctx context.Context) {
}

func (s *EVMClient) Stop() {
}

func (s *EVMClient) IsOutboundProcessed(_ context.Context, _ *xmsgtypes.Xmsg, _ zerolog.Logger) (bool, bool, error) {
	return false, false, nil
}

func (s *EVMClient) SetChainParams(chainParams observertypes.ChainParams) {
	s.ChainParams = chainParams
}

func (s *EVMClient) GetChainParams() observertypes.ChainParams {
	return s.ChainParams
}

func (s *EVMClient) OutboundID(_ uint64) string {
	return ""
}

func (s *EVMClient) WatchIntxTracker(_ context.Context) error {
	return nil
}
