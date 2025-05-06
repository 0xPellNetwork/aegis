package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestGetRevertGasLimit(t *testing.T) {
	t.Run("should return 0 if no inbound tx params", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)

		gasLimit, err := k.GetRevertGasLimit(ctx, types.Xmsg{})
		require.NoError(t, err)
		require.Equal(t, uint64(0), gasLimit)
	})
}

func TestGetAbortedAmount(t *testing.T) {

	t.Run("should return the zero if no amounts are present", func(t *testing.T) {
		xmsg := types.Xmsg{}
		a := xmsgkeeper.GetAbortedAmount(xmsg)
		require.Equal(t, sdkmath.ZeroUint(), a)
	})
}

func Test_IsPending(t *testing.T) {
	tt := []struct {
		status   types.XmsgStatus
		expected bool
	}{
		{types.XmsgStatus_PENDING_INBOUND, false},
		{types.XmsgStatus_PENDING_OUTBOUND, true},
		{types.XmsgStatus_PENDING_REVERT, true},
		{types.XmsgStatus_REVERTED, false},
		{types.XmsgStatus_ABORTED, false},
		{types.XmsgStatus_OUTBOUND_MINED, false},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("status %s", tc.status), func(t *testing.T) {
			require.Equal(t, tc.expected, xmsgkeeper.IsPending(&types.Xmsg{XmsgStatus: &types.Status{Status: tc.status}}))
		})
	}
}

func TestKeeper_UpdateNonce(t *testing.T) {
	t.Run("should error if supported chain is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(nil)

		err := k.UpdateNonce(ctx, 5, nil)
		require.Error(t, err)
	})

	t.Run("should error if chain nonces not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{
			Id: 5,
		})
		observerMock.On("GetChainNonces", mock.Anything, mock.Anything).Return(observertypes.ChainNonces{}, false)
		xmsg := types.Xmsg{
			InboundTxParams:  &types.InboundTxParams{},
			OutboundTxParams: []*types.OutboundTxParams{{}},
		}
		err := k.UpdateNonce(ctx, 5, &xmsg)
		require.Error(t, err)
	})

	t.Run("should error if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{
			Id: 5,
		})
		observerMock.On("GetChainNonces", mock.Anything, mock.Anything).Return(observertypes.ChainNonces{
			Nonce: 100,
		}, true)
		observerMock.On("GetTSS", mock.Anything).Return(observertypes.TSS{}, false)
		xmsg := types.Xmsg{
			InboundTxParams:  &types.InboundTxParams{},
			OutboundTxParams: []*types.OutboundTxParams{{}},
		}
		err := k.UpdateNonce(ctx, 5, &xmsg)
		require.Error(t, err)
		require.Equal(t, uint64(100), xmsg.GetCurrentOutTxParam().OutboundTxTssNonce)
	})

	t.Run("should error if pending nonces not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{
			Id: 5,
		})
		observerMock.On("GetChainNonces", mock.Anything, mock.Anything).Return(observertypes.ChainNonces{
			Nonce: 100,
		}, true)
		observerMock.On("GetTSS", mock.Anything).Return(observertypes.TSS{}, true)
		observerMock.On("GetPendingNonces", mock.Anything, mock.Anything, mock.Anything).Return(observertypes.PendingNonces{}, false)

		xmsg := types.Xmsg{
			InboundTxParams:  &types.InboundTxParams{},
			OutboundTxParams: []*types.OutboundTxParams{{}},
		}
		err := k.UpdateNonce(ctx, 5, &xmsg)
		require.Error(t, err)
	})

	t.Run("should error if nonces not equal", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{

			Id: 5,
		})
		observerMock.On("GetChainNonces", mock.Anything, mock.Anything).Return(observertypes.ChainNonces{
			Nonce: 100,
		}, true)
		observerMock.On("GetTSS", mock.Anything).Return(observertypes.TSS{}, true)
		observerMock.On("GetPendingNonces", mock.Anything, mock.Anything, mock.Anything).Return(observertypes.PendingNonces{
			NonceHigh: 99,
		}, true)

		xmsg := types.Xmsg{
			InboundTxParams:  &types.InboundTxParams{},
			OutboundTxParams: []*types.OutboundTxParams{{}},
		}
		err := k.UpdateNonce(ctx, 5, &xmsg)
		require.Error(t, err)
	})

	t.Run("should update nonces", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{

			Id: 5,
		})
		observerMock.On("GetChainNonces", mock.Anything, mock.Anything).Return(observertypes.ChainNonces{
			Nonce: 100,
		}, true)
		observerMock.On("GetTSS", mock.Anything).Return(observertypes.TSS{}, true)
		observerMock.On("GetPendingNonces", mock.Anything, mock.Anything, mock.Anything).Return(observertypes.PendingNonces{
			NonceHigh: 100,
		}, true)

		observerMock.On("SetChainNonces", mock.Anything, mock.Anything).Once()
		observerMock.On("SetPendingNonces", mock.Anything, mock.Anything).Once()

		xmsg := types.Xmsg{
			InboundTxParams:  &types.InboundTxParams{},
			OutboundTxParams: []*types.OutboundTxParams{{}},
		}
		err := k.UpdateNonce(ctx, 5, &xmsg)
		require.NoError(t, err)
	})
}
