package keeper_test

import (
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestKeeper_GetBlame(t *testing.T) {
	k, ctx, _, _ := keepertest.RelayerKeeper(t)
	var chainId int64 = 97
	var nonce uint64 = 101
	digest := sample.PellIndex(t)

	index := types.GetBlameIndex(chainId, nonce, digest, 123)

	k.SetBlame(ctx, types.Blame{
		Index:         index,
		FailureReason: "failed to join party",
		Nodes:         nil,
	})

	blameRecords, found := k.GetBlame(ctx, index)
	require.True(t, found)
	require.Equal(t, index, blameRecords.Index)
}

func TestKeeper_GetBlameByChainAndNonce(t *testing.T) {
	k, ctx, _, _ := keepertest.RelayerKeeper(t)
	var chainId int64 = 97
	var nonce uint64 = 101
	digest := sample.PellIndex(t)

	index := types.GetBlameIndex(chainId, nonce, digest, 123)

	k.SetBlame(ctx, types.Blame{
		Index:         index,
		FailureReason: "failed to join party",
		Nodes:         nil,
	})

	blameRecords, found := k.GetBlamesByChainAndNonce(ctx, chainId, int64(nonce))
	require.True(t, found)
	require.Equal(t, 1, len(blameRecords))
	require.Equal(t, index, blameRecords[0].Index)
}

func TestKeeper_BlameAll(t *testing.T) {
	t.Run("GetBlameRecord by limit ", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		blameList := sample.BlameRecordsList_pell(t, 10)
		for _, record := range blameList {
			k.SetBlame(ctx, record)
		}
		sort.SliceStable(blameList, func(i, j int) bool {
			return blameList[i].Index < blameList[j].Index
		})
		rst, pageRes, err := k.GetAllBlamePaginated(ctx, &query.PageRequest{Limit: 10, CountTotal: true})
		require.NoError(t, err)
		sort.SliceStable(rst, func(i, j int) bool {
			return rst[i].Index < rst[j].Index
		})
		require.Equal(t, blameList, rst)
		require.Equal(t, len(blameList), int(pageRes.Total))
	})
	t.Run("GetBlameRecord by offset ", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		blameList := sample.BlameRecordsList_pell(t, 20)
		offset := 10
		for _, record := range blameList {
			k.SetBlame(ctx, record)
		}
		sort.SliceStable(blameList, func(i, j int) bool {
			return blameList[i].Index < blameList[j].Index
		})
		rst, pageRes, err := k.GetAllBlamePaginated(ctx, &query.PageRequest{Offset: uint64(offset), CountTotal: true})
		require.NoError(t, err)
		sort.SliceStable(rst, func(i, j int) bool {
			return rst[i].Index < rst[j].Index
		})
		require.Subset(t, blameList, rst)
		require.Equal(t, len(blameList)-offset, len(rst))
		require.Equal(t, len(blameList), int(pageRes.Total))
	})
	t.Run("GetAllBlameRecord", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		blameList := sample.BlameRecordsList_pell(t, 100)
		for _, record := range blameList {
			k.SetBlame(ctx, record)
		}
		rst := k.GetAllBlame(ctx)
		sort.SliceStable(rst, func(i, j int) bool {
			return rst[i].Index < rst[j].Index
		})
		sort.SliceStable(blameList, func(i, j int) bool {
			return blameList[i].Index < blameList[j].Index
		})
		require.Equal(t, blameList, rst)
	})
	t.Run("Get no records if nothing is set", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		rst := k.GetAllBlame(ctx)
		require.Len(t, rst, 0)
	})
}
