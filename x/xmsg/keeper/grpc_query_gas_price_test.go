package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestGasPriceQuerySingle(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNGasPrice(k, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetGasPriceRequest
		response *types.QueryGasPriceResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetGasPriceRequest{Index: msgs[0].Index},
			response: &types.QueryGasPriceResponse{GasPrice: &msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetGasPriceRequest{Index: msgs[1].Index},
			response: &types.QueryGasPriceResponse{GasPrice: &msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetGasPriceRequest{Index: "1000000000"},
			err:     status.Error(codes.InvalidArgument, "not found"),
		},
		{
			desc: "InvalidRequest nil",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			desc:    "InvalidRequest index",
			request: &types.QueryGetGasPriceRequest{Index: "abc"},
			err:     fmt.Errorf("strconv.Atoi: parsing \"abc\": invalid syntax"),
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			response, err := k.GasPrice(wctx, tc.request)
			if tc.err != nil {
				require.Error(t, err)
			} else {
				require.Equal(t, tc.response, response)
			}
		})
	}
}

func TestGasPriceQueryPaginated(t *testing.T) {
	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNGasPrice(k, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllGasPriceRequest {
		return &types.QueryAllGasPriceRequest{
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
			resp, err := k.GasPriceAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			for j := i; j < len(msgs) && j < i+step; j++ {
				require.Equal(t, &msgs[j], resp.GasPrices[j-i])
			}
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := k.GasPriceAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			for j := i; j < len(msgs) && j < i+step; j++ {
				require.Equal(t, &msgs[j], resp.GasPrices[j-i])
			}
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := k.GasPriceAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := k.GasPriceAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
