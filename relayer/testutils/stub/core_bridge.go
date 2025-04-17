package stub

import (
	"context"
	"errors"
	"math/big"

	"cosmossdk.io/math"
	"github.com/rs/zerolog"
	"gitlab.com/thorchain/tss/go-tss/blame"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/keys"
	keyinterfaces "github.com/0xPellNetwork/aegis/relayer/keys/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/testutils"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	observerTypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const ErrMsgPaused = "pell core bridge is paused"

var _ interfaces.PellCoreBridger = &MockPellCoreBridge{}

type MockPellCoreBridge struct {
	paused bool
	chain  chains.Chain
}

func NewMockPellCoreBridge() *MockPellCoreBridge {
	chain, err := chains.PellChainFromChainID("pellchain_86-1")
	if err != nil {
		panic(err)
	}
	return &MockPellCoreBridge{
		paused: false,
		chain:  chain,
	}
}

func (z *MockPellCoreBridge) PostVoteInboundEvents(ctx context.Context, _, _ uint64, _ []*xmsgtypes.MsgVoteOnObservedInboundTx) (string, string, error) {
	if z.paused {
		return "", "", errors.New(ErrMsgPaused)
	}
	return "", "", nil
}

func (z *MockPellCoreBridge) PostVoteOutbound(ctx context.Context, _ string, _ string, _ uint64, _ uint64, _ *big.Int, _ uint64, _ chains.ReceiveStatus, _ string, _ chains.Chain, _ uint64) (string, string, error) {
	if z.paused {
		return "", "", errors.New(ErrMsgPaused)
	}
	return sample.Hash().Hex(), "", nil
}

func (z *MockPellCoreBridge) PostGasPrice(ctx context.Context, _ chains.Chain, _ uint64, _ string, _ uint64) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return "", nil
}

func (z *MockPellCoreBridge) PostVoteBlockHeader(ctx context.Context, _ int64, _ []byte, _ int64, _ proofs.HeaderData) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return "", nil
}

func (z *MockPellCoreBridge) GetBlockHeaderChainState(ctx context.Context, _ int64) (*lightclienttypes.QueryChainStateResponse, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}
	return &lightclienttypes.QueryChainStateResponse{}, nil
}

func (z *MockPellCoreBridge) PostBlameData(ctx context.Context, _ *blame.Blame, _ int64, _ string) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return "", nil
}

func (z *MockPellCoreBridge) PostAddTxHashToOutTxTracker(ctx context.Context, _ int64, _ uint64, _ string, _ *proofs.Proof, _ string, _ int64) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return "", nil
}

func (z *MockPellCoreBridge) PostVoteOnPellRecharge(ctx context.Context, chain chains.Chain, voteIndex uint64) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return "", nil
}

func (z *MockPellCoreBridge) PostVoteOnGasRecharge(ctx context.Context, chain chains.Chain, voteIndex uint64) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return "", nil
}

func (z *MockPellCoreBridge) Chain() chains.Chain {
	return z.chain
}

func (z *MockPellCoreBridge) GetLogger() *zerolog.Logger {
	return nil
}

func (z *MockPellCoreBridge) GetKeys() keyinterfaces.ObserverKeys {
	return &keys.Keys{}
}

func (z *MockPellCoreBridge) GetBlockHeight(ctx context.Context) (int64, error) {
	if z.paused {
		return 0, errors.New(ErrMsgPaused)
	}
	return 0, nil
}

func (z *MockPellCoreBridge) GetChainIndex(ctx context.Context, chainId int64) (*xmsgtypes.ChainIndex, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}

	return &xmsgtypes.ChainIndex{}, nil
}

func (z *MockPellCoreBridge) GetPellRechargeOperationIndex(ctx context.Context, chainId int64) (*xmsgtypes.PellRechargeOperationIndex, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}

	return &xmsgtypes.PellRechargeOperationIndex{}, nil
}

func (z *MockPellCoreBridge) GetGasRechargeOperationIndex(ctx context.Context, chainId int64) (*xmsgtypes.GasRechargeOperationIndex, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}

	return &xmsgtypes.GasRechargeOperationIndex{}, nil
}

func (z *MockPellCoreBridge) GetLastBlockHeightByChain(ctx context.Context, _ chains.Chain) (*xmsgtypes.LastBlockHeight, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}

	return &xmsgtypes.LastBlockHeight{}, nil
}

func (z *MockPellCoreBridge) GetRateLimiterFlags(ctx context.Context) (xmsgtypes.RateLimiterFlags, error) {
	if z.paused {
		return xmsgtypes.RateLimiterFlags{}, errors.New(ErrMsgPaused)
	}

	return xmsgtypes.RateLimiterFlags{}, nil
}

func (z *MockPellCoreBridge) GetRateLimiterInput(ctx context.Context, window int64) (*xmsgtypes.QueryRateLimiterInputResponse, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}

	return &xmsgtypes.QueryRateLimiterInputResponse{}, nil
}

func (z *MockPellCoreBridge) GetPellBlockHeight(ctx context.Context) (int64, error) {
	if z.paused {
		return 0, errors.New(ErrMsgPaused)
	}
	return 0, nil
}

func (z *MockPellCoreBridge) ListPendingXmsg(ctx context.Context, _ int64) ([]*xmsgtypes.Xmsg, uint64, error) {
	if z.paused {
		return nil, 0, errors.New(ErrMsgPaused)
	}
	return []*xmsgtypes.Xmsg{}, 0, nil
}

func (z *MockPellCoreBridge) ListPendingXmsgWithinRatelimit(ctx context.Context) (*xmsgtypes.QueryListPendingXmsgWithinRateLimitResponse, error) {
	if z.paused {
		return &xmsgtypes.QueryListPendingXmsgWithinRateLimitResponse{}, errors.New(ErrMsgPaused)
	}
	return &xmsgtypes.QueryListPendingXmsgWithinRateLimitResponse{}, nil
}

func (z *MockPellCoreBridge) GetPendingNoncesByChain(ctx context.Context, _ int64) (observerTypes.PendingNonces, error) {
	if z.paused {
		return observerTypes.PendingNonces{}, errors.New(ErrMsgPaused)
	}
	return observerTypes.PendingNonces{}, nil
}

func (z *MockPellCoreBridge) GetXmsgByNonce(ctx context.Context, _ int64, _ uint64) (*xmsgtypes.Xmsg, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}
	return &xmsgtypes.Xmsg{}, nil
}

func (z *MockPellCoreBridge) GetOutTxTracker(ctx context.Context, _ chains.Chain, _ uint64) (*xmsgtypes.OutTxTracker, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}
	return &xmsgtypes.OutTxTracker{}, nil
}

func (z *MockPellCoreBridge) GetAllOutTxTrackerByChain(ctx context.Context, _ int64, _ interfaces.Order) ([]xmsgtypes.OutTxTracker, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}
	return []xmsgtypes.OutTxTracker{}, nil
}

func (z *MockPellCoreBridge) GetCrosschainFlags(ctx context.Context) (observerTypes.CrosschainFlags, error) {
	if z.paused {
		return observerTypes.CrosschainFlags{}, errors.New(ErrMsgPaused)
	}
	return observerTypes.CrosschainFlags{}, nil
}

func (z *MockPellCoreBridge) GetObserverList(ctx context.Context) ([]string, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}
	return []string{}, nil
}

func (z *MockPellCoreBridge) GetKeyGen(ctx context.Context) (observerTypes.Keygen, error) {
	if z.paused {
		return observerTypes.Keygen{}, errors.New(ErrMsgPaused)
	}
	return observerTypes.Keygen{}, nil
}

func (z *MockPellCoreBridge) GetBTCTSSAddress(ctx context.Context, _ int64) (string, error) {
	if z.paused {
		return "", errors.New(ErrMsgPaused)
	}
	return testutils.TSSAddressBTCMainnet, nil
}

func (z *MockPellCoreBridge) GetInboundTrackersForChain(ctx context.Context, _ int64) ([]xmsgtypes.InTxTracker, error) {
	if z.paused {
		return nil, errors.New(ErrMsgPaused)
	}
	return []xmsgtypes.InTxTracker{}, nil
}

func (z *MockPellCoreBridge) GetPellHotKeyBalance(ctx context.Context) (math.Int, error) {
	if z.paused {
		return math.NewInt(0), errors.New(ErrMsgPaused)
	}
	return math.NewInt(0), nil
}

func (z *MockPellCoreBridge) Stop() {

}

func (z *MockPellCoreBridge) OnBeforeStop(callback func()) {

}

func (z *MockPellCoreBridge) PostVoteInboundBlock(
	ctx context.Context,
	gasLimit, retryLimit uint64,
	block *xmsgtypes.MsgVoteInboundBlock,
	events []*xmsgtypes.MsgVoteOnObservedInboundTx,
) ([]string, []string, error) {
	if z.paused {
		return nil, nil, errors.New(ErrMsgPaused)
	}
	return []string{}, []string{}, nil
}
