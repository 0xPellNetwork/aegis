package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func createNXmsgWithStatus(keeper *keeper.Keeper, ctx sdk.Context, n int, status types.XmsgStatus) []types.Xmsg {
	items := make([]types.Xmsg, n)
	for i := range items {
		items[i].Signer = "any"
		items[i].Index = fmt.Sprintf("%d-%d", i, status)
		items[i].XmsgStatus = &types.Status{
			Status:              status,
			StatusMessage:       "",
			LastUpdateTimestamp: 0,
		}
		items[i].InboundTxParams = &types.InboundTxParams{InboundTxHash: fmt.Sprint(i)}
		items[i].OutboundTxParams = []*types.OutboundTxParams{{}}

		keeper.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, items[i])
	}
	return items
}

// Keeper Tests
func createNXmsg(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Xmsg {
	items := make([]types.Xmsg, n)
	for i := range items {
		items[i].Signer = "any"
		items[i].InboundTxParams = &types.InboundTxParams{
			Sender:                       fmt.Sprint(i),
			SenderChainId:                int64(i),
			TxOrigin:                     fmt.Sprint(i),
			InboundTxHash:                fmt.Sprint(i),
			InboundTxBlockHeight:         uint64(i),
			InboundTxFinalizedPellHeight: uint64(i),
			InboundPellTx:                nil,
		}
		items[i].OutboundTxParams = []*types.OutboundTxParams{{
			Receiver:                 fmt.Sprint(i),
			ReceiverChainId:          int64(i),
			OutboundTxHash:           fmt.Sprint(i),
			OutboundTxTssNonce:       uint64(i),
			OutboundTxGasLimit:       uint64(i),
			OutboundTxGasPrice:       fmt.Sprint(i),
			OutboundTxBallotIndex:    fmt.Sprint(i),
			OutboundTxExternalHeight: uint64(i),
		}}
		items[i].XmsgStatus = &types.Status{
			Status:              types.XmsgStatus_PENDING_INBOUND,
			StatusMessage:       "any",
			LastUpdateTimestamp: 0,
		}

		items[i].Index = fmt.Sprint(i)

		keeper.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, items[i])
	}
	return items
}

func TestSends(t *testing.T) {
	sendsTest := []struct {
		TestName        string
		PendingInbound  int
		PendingOutbound int
		OutboundMined   int
		Confirmed       int
		PendingRevert   int
		Reverted        int
		Aborted         int
	}{
		{
			TestName:        "test pending",
			PendingInbound:  10,
			PendingOutbound: 10,
			Confirmed:       10,
			PendingRevert:   10,
			Aborted:         10,
			OutboundMined:   10,
			Reverted:        10,
		},
		{
			TestName:        "test pending random",
			PendingInbound:  rand.Intn(300-10) + 10,
			PendingOutbound: rand.Intn(300-10) + 10,
			Confirmed:       rand.Intn(300-10) + 10,
			PendingRevert:   rand.Intn(300-10) + 10,
			Aborted:         rand.Intn(300-10) + 10,
			OutboundMined:   rand.Intn(300-10) + 10,
			Reverted:        rand.Intn(300-10) + 10,
		},
	}
	for _, tt := range sendsTest {
		tt := tt
		t.Run(tt.TestName, func(t *testing.T) {
			keeper, ctx, _, zk := keepertest.XmsgKeeper(t)
			var sends []types.Xmsg
			zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
			sends = append(sends, createNXmsgWithStatus(keeper, ctx, tt.PendingInbound, types.XmsgStatus_PENDING_INBOUND)...)
			sends = append(sends, createNXmsgWithStatus(keeper, ctx, tt.PendingOutbound, types.XmsgStatus_PENDING_OUTBOUND)...)
			sends = append(sends, createNXmsgWithStatus(keeper, ctx, tt.PendingRevert, types.XmsgStatus_PENDING_REVERT)...)
			sends = append(sends, createNXmsgWithStatus(keeper, ctx, tt.Aborted, types.XmsgStatus_ABORTED)...)
			sends = append(sends, createNXmsgWithStatus(keeper, ctx, tt.OutboundMined, types.XmsgStatus_OUTBOUND_MINED)...)
			sends = append(sends, createNXmsgWithStatus(keeper, ctx, tt.Reverted, types.XmsgStatus_REVERTED)...)
			//require.Equal(t, tt.PendingOutbound, len(keeper.GetAllXmsgByStatuses(ctx, []types.XmsgStatus{types.XmsgStatus_PendingOutbound})))
			//require.Equal(t, tt.PendingInbound, len(keeper.GetAllXmsgByStatuses(ctx, []types.XmsgStatus{types.XmsgStatus_PendingInbound})))
			//require.Equal(t, tt.PendingOutbound+tt.PendingRevert, len(keeper.GetAllXmsgByStatuses(ctx, []types.XmsgStatus{types.XmsgStatus_PendingOutbound, types.XmsgStatus_PendingRevert})))
			require.Equal(t, len(sends), len(keeper.GetAllXmsg(ctx)))
			for _, s := range sends {
				send, found := keeper.GetXmsg(ctx, s.Index)
				require.True(t, found)
				require.Equal(t, s, send)
			}

		})
	}
}

func TestSendGetAll(t *testing.T) {
	keeper, ctx, _, zk := keepertest.XmsgKeeper(t)
	zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
	items := createNXmsg(keeper, ctx, 10)
	xmsg := keeper.GetAllXmsg(ctx)
	c := make([]types.Xmsg, len(xmsg))
	for i, val := range xmsg {
		c[i] = val
	}
	require.Equal(t, items, c)
}

// Querier Tests

func TestSendQuerySingle(t *testing.T) {
	keeper, ctx, _, zk := keepertest.XmsgKeeper(t)
	zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNXmsg(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetXmsgRequest
		response *types.QueryXmsgResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetXmsgRequest{Index: msgs[0].Index},
			response: &types.QueryXmsgResponse{Xmsg: &msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetXmsgRequest{Index: msgs[1].Index},
			response: &types.QueryXmsgResponse{Xmsg: &msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetXmsgRequest{Index: "missing"},
			err:     status.Error(codes.InvalidArgument, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Xmsg(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.Equal(t, tc.response, response)
			}
		})
	}
}

func TestSendQueryPaginated(t *testing.T) {
	keeper, ctx, _, zk := keepertest.XmsgKeeper(t)
	zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNXmsg(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllXmsgRequest {
		return &types.QueryAllXmsgRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.XmsgAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			for j := i; j < len(msgs) && j < i+step; j++ {
				require.Equal(t, &msgs[j], resp.Xmsgs[j-i])
			}
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.XmsgAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			for j := i; j < len(msgs) && j < i+step; j++ {
				require.Equal(t, &msgs[j], resp.Xmsgs[j-i])
			}
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.XmsgAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.XmsgAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func TestKeeper_RemoveXmsg(t *testing.T) {
	keeper, ctx, _, zk := keepertest.XmsgKeeper(t)
	zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
	txs := createNXmsg(keeper, ctx, 5)

	keeper.RemoveXmsg(ctx, txs[0].Index)
	txs = keeper.GetAllXmsg(ctx)
	require.Equal(t, 4, len(txs))
}
