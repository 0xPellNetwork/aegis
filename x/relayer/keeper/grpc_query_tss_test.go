package keeper_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestTSSQuerySingle(t *testing.T) {
	k, ctx, _, _ := keepertest.RelayerKeeper(t)
	tss := sample.Tss_pell()
	wctx := sdk.WrapSDKContext(ctx)

	for _, tc := range []struct {
		desc           string
		request        *types.QueryGetTSSRequest
		response       *types.QueryTSSResponse
		skipSettingTss bool
		err            error
	}{
		{
			desc:           "Skip setting tss",
			request:        &types.QueryGetTSSRequest{},
			skipSettingTss: true,
			err:            status.Error(codes.InvalidArgument, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			desc:     "Should return tss",
			request:  &types.QueryGetTSSRequest{},
			response: &types.QueryTSSResponse{Tss: tss},
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			if !tc.skipSettingTss {
				k.SetTSS(ctx, tss)
			}
			response, err := k.TSS(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.Equal(t, tc.response, response)
			}
		})
	}
}

func TestTSSQueryHistory(t *testing.T) {
	keeper, ctx, _, _ := keepertest.RelayerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	for _, tc := range []struct {
		desc          string
		tssCount      int
		foundPrevious bool
		err           error
	}{
		{
			desc:          "1 Tss addresses",
			tssCount:      1,
			foundPrevious: false,
			err:           nil,
		},
		{
			desc:          "10 Tss addresses",
			tssCount:      10,
			foundPrevious: true,
			err:           nil,
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			tssList := sample.TssList_pell(tc.tssCount)
			for _, tss := range tssList {
				keeper.SetTSS(ctx, tss)
				keeper.SetTSSHistory(ctx, tss)
			}
			request := &types.QueryTssHistoryRequest{}
			response, err := keeper.TssHistory(wctx, request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.Equal(t, len(tssList), len(response.TssList))
				prevTss, found := keeper.GetPreviousTSS(ctx)
				require.Equal(t, tc.foundPrevious, found)
				if found {
					require.Equal(t, tssList[len(tssList)-2], prevTss)
				}
			}
		})
	}
}

func TestKeeper_GetTssAddress(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetTssAddress(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetTssAddress(wctx, &types.QueryGetTssAddressRequest{
			BitcoinChainId: 1,
		})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if invalid chain id", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		tss := sample.Tss_pell()
		k.SetTSS(ctx, tss)

		// because there is no limit for invalid chain id now
		res, err := k.GetTssAddress(wctx, &types.QueryGetTssAddressRequest{
			BitcoinChainId: 9876,
		})
		require.NotNil(t, res)
		require.Nil(t, err)
	})

}

func TestKeeper_GetTssAddressByFinalizedHeight(t *testing.T) {
	t.Run("should error if req is nil", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetTssAddressByFinalizedHeight(wctx, nil)
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		res, err := k.GetTssAddressByFinalizedHeight(wctx, &types.QueryGetTssAddressByFinalizedHeightRequest{
			BitcoinChainId: 1,
		})
		require.Nil(t, res)
		require.Error(t, err)
	})

	t.Run("should error if invalid chain id", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		wctx := sdk.WrapSDKContext(ctx)

		tssList := sample.TssList_pell(100)
		r := rand.Intn((len(tssList)-1)-0) + 0
		for _, tss := range tssList {
			k.SetTSSHistory(ctx, tss)
		}

		res, err := k.GetTssAddressByFinalizedHeight(wctx, &types.QueryGetTssAddressByFinalizedHeightRequest{
			BitcoinChainId:      987,
			FinalizedPellHeight: tssList[r].FinalizedPellHeight,
		})
		require.NotNil(t, res)
		require.Nil(t, err)
	})

}
