package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/keeper"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestKeeper_XmsgListPending(t *testing.T) {
	t.Run("should fail for empty req", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		_, err := k.ListPendingXmsg(ctx, nil)
		require.ErrorContains(t, err, "invalid request")
	})

	t.Run("should use max limit if limit is too high", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		_, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{Limit: keeper.MaxPendingXmsgs + 1})
		require.ErrorContains(t, err, "tss not found")
	})

	t.Run("should fail if no TSS", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		_, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{Limit: 1})
		require.ErrorContains(t, err, "tss not found")
	})

	t.Run("should return empty list if no nonces", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)

		//  set TSS
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())

		_, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{Limit: 1})
		require.ErrorContains(t, err, "pending nonces not found")
	})

	t.Run("can retrieve pending xmsg in range", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)
		xmsgs := createXmsgWithNonceRange(t, ctx, *k, 1000, 2000, chainID, tss, zk)

		res, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{ChainId: chainID, Limit: 100})
		require.NoError(t, err)
		require.Equal(t, 100, len(res.Xmsg))
		require.EqualValues(t, xmsgs[0:100], res.Xmsg)
		require.EqualValues(t, uint64(1000), res.TotalPending)

		res, err = k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{ChainId: chainID})
		require.NoError(t, err)
		require.Equal(t, keeper.MaxPendingXmsgs, len(res.Xmsg))
		require.EqualValues(t, xmsgs[0:keeper.MaxPendingXmsgs], res.Xmsg)
		require.EqualValues(t, uint64(1000), res.TotalPending)
	})

	t.Run("can retrieve pending xmsg with range smaller than max", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)
		xmsgs := createXmsgWithNonceRange(t, ctx, *k, 1000, 1100, chainID, tss, zk)

		res, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{ChainId: chainID})
		require.NoError(t, err)
		require.Equal(t, 100, len(res.Xmsg))
		require.EqualValues(t, xmsgs, res.Xmsg)
		require.EqualValues(t, uint64(100), res.TotalPending)
	})

	t.Run("can retrieve pending xmsg with pending xmsg below nonce low", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)
		xmsgs := createXmsgWithNonceRange(t, ctx, *k, 1000, 2000, chainID, tss, zk)

		// set some xmsgs as pending below nonce
		xmsg1, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("1337-940"))
		require.True(t, found)
		xmsg1.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, xmsg1)

		xmsg2, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("1337-955"))
		require.True(t, found)
		xmsg2.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, xmsg2)

		res, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{ChainId: chainID, Limit: 100})
		require.NoError(t, err)
		require.Equal(t, 100, len(res.Xmsg))

		expectedXmsgs := append([]*types.Xmsg{&xmsg1, &xmsg2}, xmsgs[0:98]...)
		require.EqualValues(t, expectedXmsgs, res.Xmsg)

		// pending nonce + 2
		require.EqualValues(t, uint64(1002), res.TotalPending)
	})

	t.Run("error if some before low nonce are missing", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)
		xmsgs := createXmsgWithNonceRange(t, ctx, *k, 1000, 2000, chainID, tss, zk)

		// set some xmsgs as pending below nonce
		xmsg1, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("1337-940"))
		require.True(t, found)
		xmsg1.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, xmsg1)

		xmsg2, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("1337-955"))
		require.True(t, found)
		xmsg2.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, xmsg2)

		res, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{ChainId: chainID, Limit: 100})
		require.NoError(t, err)
		require.Equal(t, 100, len(res.Xmsg))

		expectedXmsgs := append([]*types.Xmsg{&xmsg1, &xmsg2}, xmsgs[0:98]...)
		require.EqualValues(t, expectedXmsgs, res.Xmsg)

		// pending nonce + 2
		require.EqualValues(t, uint64(1002), res.TotalPending)
	})
}

func TestKeeper_XmsgByNonce(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		res, err := k.XmsgByNonce(ctx, nil)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		res, err := k.XmsgByNonce(ctx, &types.QueryGetXmsgByNonceRequest{
			ChainId: 1,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if nonce to xmsg not found", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)

		res, err := k.XmsgByNonce(ctx, &types.QueryGetXmsgByNonceRequest{
			ChainId: chainID,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if crosschain tx not found", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)
		nonce := 1000
		xmsg := sample.Xmsg_pell(t, fmt.Sprintf("%d-%d", chainID, nonce))

		zk.ObserverKeeper.SetNonceToXmsg(ctx, observertypes.NonceToXmsg{
			ChainId:   chainID,
			Nonce:     int64(nonce),
			XmsgIndex: xmsg.Index,
			Tss:       tss.TssPubkey,
		})

		res, err := k.XmsgByNonce(ctx, &types.QueryGetXmsgByNonceRequest{
			ChainId: chainID,
			Nonce:   uint64(nonce),
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should return if crosschain tx found", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		chainID := getValidEthChainID()
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)
		nonce := 1000
		xmsg := sample.Xmsg_pell(t, fmt.Sprintf("%d-%d", chainID, nonce))

		zk.ObserverKeeper.SetNonceToXmsg(ctx, observertypes.NonceToXmsg{
			ChainId:   chainID,
			Nonce:     int64(nonce),
			XmsgIndex: xmsg.Index,
			Tss:       tss.TssPubkey,
		})
		k.SetXmsg(ctx, *xmsg)

		res, err := k.XmsgByNonce(ctx, &types.QueryGetXmsgByNonceRequest{
			ChainId: chainID,
			Nonce:   uint64(nonce),
		})
		require.NoError(t, err)
		require.Equal(t, xmsg, res.Xmsg)
	})
}
