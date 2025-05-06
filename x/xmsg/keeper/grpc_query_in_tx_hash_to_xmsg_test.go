package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/nullify"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestInTxHashToXmsgQuerySingle(t *testing.T) {
	keeper, ctx, _, _ := keepertest.XmsgKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNInTxHashToXmsg(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetInTxHashToXmsgRequest
		response *types.QueryInTxHashToXmsgResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetInTxHashToXmsgRequest{
				InTxHash: msgs[0].InTxHash,
			},
			response: &types.QueryInTxHashToXmsgResponse{InTxHashToXmsg: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetInTxHashToXmsgRequest{
				InTxHash: msgs[1].InTxHash,
			},
			response: &types.QueryInTxHashToXmsgResponse{InTxHashToXmsg: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetInTxHashToXmsgRequest{
				InTxHash: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.InTxHashToXmsg(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestInTxHashToXmsgQueryPaginated(t *testing.T) {
	keeper, ctx, _, _ := keepertest.XmsgKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNInTxHashToXmsg(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllInTxHashToXmsgRequest {
		return &types.QueryAllInTxHashToXmsgRequest{
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
			resp, err := keeper.InTxHashToXmsgAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.InTxHashToXmsg), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.InTxHashToXmsg),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.InTxHashToXmsgAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.InTxHashToXmsg), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.InTxHashToXmsg),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.InTxHashToXmsgAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.InTxHashToXmsg),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.InTxHashToXmsgAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}

func createInTxHashToXmsgWithXmsgs(keeper *xmsgkeeper.Keeper, ctx sdk.Context) ([]types.Xmsg,
	types.InTxHashToXmsg) {
	xmsgs := make([]types.Xmsg, 5)
	for i := range xmsgs {
		xmsgs[i].Signer = "any"
		xmsgs[i].Index = fmt.Sprintf("0x123%d", i)
		xmsgs[i].InboundTxParams = &types.InboundTxParams{InboundTxHash: fmt.Sprint(i)}
		xmsgs[i].XmsgStatus = &types.Status{Status: types.XmsgStatus_PENDING_INBOUND}
		keeper.SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx, xmsgs[i])
	}

	var inTxHashToXmsg types.InTxHashToXmsg
	inTxHashToXmsg.InTxHash = fmt.Sprintf("0xabc")
	for i := range xmsgs {
		inTxHashToXmsg.XmsgIndices = append(inTxHashToXmsg.XmsgIndices, xmsgs[i].Index)
	}
	keeper.SetInTxHashToXmsg(ctx, inTxHashToXmsg)

	return xmsgs, inTxHashToXmsg
}

func TestKeeper_InTxHashToXmsgDataQuery(t *testing.T) {
	keeper, ctx, _, zk := keepertest.XmsgKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
	t.Run("can query all xmsgs data with in tx hash", func(t *testing.T) {
		xmsgs, inTxHashToXmsg := createInTxHashToXmsgWithXmsgs(keeper, ctx)
		req := &types.QueryInTxHashToXmsgDataRequest{
			InTxHash: inTxHashToXmsg.InTxHash,
		}
		res, err := keeper.InTxHashToXmsgData(wctx, req)
		require.NoError(t, err)
		require.Equal(t, len(xmsgs), len(res.Xmsgs))
		for i := range xmsgs {
			require.Equal(t, nullify.Fill(xmsgs[i]), nullify.Fill(res.Xmsgs[i]))
		}
	})
	t.Run("in tx hash not found", func(t *testing.T) {
		req := &types.QueryInTxHashToXmsgDataRequest{
			InTxHash: "notfound",
		}
		_, err := keeper.InTxHashToXmsgData(wctx, req)
		require.ErrorIs(t, err, status.Error(codes.NotFound, "not found"))
	})
	t.Run("xmsg not indexed return internal error", func(t *testing.T) {
		keeper.SetInTxHashToXmsg(ctx, types.InTxHashToXmsg{
			InTxHash:    "noxmsg",
			XmsgIndices: []string{"notfound"},
		})

		req := &types.QueryInTxHashToXmsgDataRequest{
			InTxHash: "noxmsg",
		}
		_, err := keeper.InTxHashToXmsgData(wctx, req)
		require.ErrorIs(t, err, status.Error(codes.Internal, "xmsg indexed notfound doesn't exist"))
	})
}
