package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/xmsg/keeper"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestMsgServer_VoteGasPrice(t *testing.T) {
	t.Run("should error if unsupported chain", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(nil)

		msgServer := keeper.NewMsgServerImpl(*k)

		res, err := msgServer.VoteGasPrice(ctx, &types.MsgVoteGasPrice{
			ChainId: 5,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if not non tombstoned observer", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)

		msgServer := keeper.NewMsgServerImpl(*k)

		res, err := msgServer.VoteGasPrice(ctx, &types.MsgVoteGasPrice{
			ChainId: 5,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should not error if gas price found and msg.Signer in signers", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{
			Id: 5,
		})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(true)

		msgServer := keeper.NewMsgServerImpl(*k)

		creator := sample.AccAddress()
		k.SetGasPrice(ctx, types.GasPrice{
			Signer:    creator,
			ChainId:   5,
			Signers:   []string{creator},
			BlockNums: []uint64{1},
			Prices:    []uint64{1},
		})

		res, err := msgServer.VoteGasPrice(ctx, &types.MsgVoteGasPrice{
			Signer:      creator,
			ChainId:     5,
			BlockNumber: 2,
			Price:       2,
		})
		require.NoError(t, err)
		require.Empty(t, res)
		gp, found := k.GetGasPrice(ctx, 5)
		require.True(t, found)
		require.Equal(t, types.GasPrice{
			Signer:      creator,
			Index:       "",
			ChainId:     5,
			Signers:     []string{creator},
			BlockNums:   []uint64{2},
			Prices:      []uint64{2},
			MedianIndex: 0,
		}, gp)
	})

	t.Run("should not error if gas price found and msg.Signer not in signers", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{
			Id: 5,
		})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(true)

		msgServer := keeper.NewMsgServerImpl(*k)

		creator := sample.AccAddress()
		k.SetGasPrice(ctx, types.GasPrice{
			Signer:    creator,
			ChainId:   5,
			BlockNums: []uint64{1},
			Prices:    []uint64{1},
		})

		res, err := msgServer.VoteGasPrice(ctx, &types.MsgVoteGasPrice{
			Signer:      creator,
			ChainId:     5,
			BlockNumber: 2,
			Price:       2,
		})
		require.NoError(t, err)
		require.Empty(t, res)
		gp, found := k.GetGasPrice(ctx, 5)
		require.True(t, found)
		require.Equal(t, types.GasPrice{
			Signer:      creator,
			Index:       "",
			ChainId:     5,
			Signers:     []string{creator},
			BlockNums:   []uint64{1, 2},
			Prices:      []uint64{1, 2},
			MedianIndex: 1,
		}, gp)
	})
}
