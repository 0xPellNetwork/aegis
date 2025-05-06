package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_IsFinalizedInbound(t *testing.T) {
	t.Run("check true for finalized inbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		intxHash := sample.Hash().String()
		chainID := sample.Chain(5).Id
		eventIndex := sample.EventIndex()
		k.AddFinalizedInbound(ctx, intxHash, chainID, eventIndex)
		require.True(t, k.IsFinalizedInbound(ctx, intxHash, chainID, eventIndex))
	})
	t.Run("check false for non-finalized inbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		intxHash := sample.Hash().String()
		chainID := sample.Chain(5).Id
		eventIndex := sample.EventIndex()
		require.False(t, k.IsFinalizedInbound(ctx, intxHash, chainID, eventIndex))
	})
	t.Run("check true for finalized inbound list", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		listSize := 1000
		txHashList := make([]string, listSize)
		chainIdList := make([]int64, listSize)
		eventIndexList := make([]uint64, listSize)
		for i := 0; i < listSize; i++ {
			txHashList[i] = sample.Hash().String()
			chainIdList[i] = sample.Chain(5).Id
			eventIndexList[i] = sample.EventIndex()
			k.AddFinalizedInbound(ctx, txHashList[i], chainIdList[i], eventIndexList[i])
		}
		for i := 0; i < listSize; i++ {
			require.True(t, k.IsFinalizedInbound(ctx, txHashList[i], chainIdList[i], eventIndexList[i]))
		}
	})
}

func TestKeeper_AddFinalizedInbound(t *testing.T) {
	t.Run("check add finalized inbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		intxHash := sample.Hash().String()
		chainID := sample.Chain(5).Id
		eventIndex := sample.EventIndex()
		k.AddFinalizedInbound(ctx, intxHash, chainID, eventIndex)
		require.True(t, k.IsFinalizedInbound(ctx, intxHash, chainID, eventIndex))
	})
}

func TestKeeper_GetAllFinalizedInbound(t *testing.T) {
	t.Run("check empty list", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		list := k.GetAllFinalizedInbound(ctx)
		require.Empty(t, list)
	})
	t.Run("check list", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		listSize := 1000
		txHashList := make([]string, listSize)
		chainIdList := make([]int64, listSize)
		eventIndexList := make([]uint64, listSize)
		for i := 0; i < listSize; i++ {
			txHashList[i] = sample.Hash().String()
			chainIdList[i] = sample.Chain(5).Id
			eventIndexList[i] = sample.EventIndex()
			k.AddFinalizedInbound(ctx, txHashList[i], chainIdList[i], eventIndexList[i])
		}
		list := k.GetAllFinalizedInbound(ctx)
		require.Equal(t, listSize, len(list))
		for i := 0; i < listSize; i++ {
			require.Contains(t, list, types.FinalizedInboundKey(txHashList[i], chainIdList[i], eventIndexList[i]))
		}
		require.NotContains(t, list, types.FinalizedInboundKey(sample.Hash().String(), sample.Chain(5).Id, sample.EventIndex()))
	})
}
