package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestKeeper_GetNonceToXmsg(t *testing.T) {
	t.Run("Get nonce to xmsg", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		nonceToXmsgList := sample.NonceToXmsgList_pell(t, "sample", 1)
		for _, n := range nonceToXmsgList {
			k.SetNonceToXmsg(ctx, n)
		}
		for _, n := range nonceToXmsgList {
			rst, found := k.GetNonceToXmsg(ctx, n.Tss, n.ChainId, n.Nonce)
			require.True(t, found)
			require.Equal(t, n, rst)
		}

		for _, n := range nonceToXmsgList {
			k.RemoveNonceToXmsg(ctx, n)
		}
		for _, n := range nonceToXmsgList {
			_, found := k.GetNonceToXmsg(ctx, n.Tss, n.ChainId, n.Nonce)
			require.False(t, found)
		}
	})
	t.Run("test nonce to xmsg", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		k.SetNonceToXmsg(ctx, types.NonceToXmsg{
			ChainId:   1337,
			Nonce:     0,
			XmsgIndex: "0x705b88814b2a049e75b591fd80595c53f3bd9ddfb67ad06aa6965ed91023ee9a",
			Tss:       "pellpub1addwnpepq0akz8ene4z2mg3tghamr0m5eg3eeuqtjcfamkh5ecetua9u0pcyvjeyerd",
		})
		_, found := k.GetNonceToXmsg(ctx, "pellpub1addwnpepq0akz8ene4z2mg3tghamr0m5eg3eeuqtjcfamkh5ecetua9u0pcyvjeyerd", 1337, 0)
		require.True(t, found)
	})
	t.Run("Get nonce to xmsg not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		nonceToXmsgList := sample.NonceToXmsgList_pell(t, "sample", 1)
		for _, n := range nonceToXmsgList {
			k.SetNonceToXmsg(ctx, n)
		}
		_, found := k.GetNonceToXmsg(ctx, "not_found", 1, 1)
		require.False(t, found)
	})
	t.Run("Get all nonce to xmsg", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		nonceToXmsgList := sample.NonceToXmsgList_pell(t, "sample", 10)
		for _, n := range nonceToXmsgList {
			k.SetNonceToXmsg(ctx, n)
		}
		rst := k.GetAllNonceToXmsg(ctx)
		require.Equal(t, nonceToXmsgList, rst)
	})
}
