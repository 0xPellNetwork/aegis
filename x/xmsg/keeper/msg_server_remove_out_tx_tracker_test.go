package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestMsgServer_RemoveFromOutTxTracker(t *testing.T) {
	t.Run("should error if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			ChainId: 1,
			Nonce:   1,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)

		msgServer := keeper.NewMsgServerImpl(*k)

		res, err := msgServer.RemoveFromOutTxTracker(ctx, &types.MsgRemoveFromOutTxTracker{
			Signer: admin,
		})
		require.Error(t, err)
		require.Empty(t, res)

		_, found := k.GetOutTxTracker(ctx, 1, 1)
		require.True(t, found)
	})

	t.Run("should remove if authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		k.SetOutTxTracker(ctx, types.OutTxTracker{
			ChainId: 1,
			Nonce:   1,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)

		msgServer := keeper.NewMsgServerImpl(*k)

		res, err := msgServer.RemoveFromOutTxTracker(ctx, &types.MsgRemoveFromOutTxTracker{
			Signer:  admin,
			ChainId: 1,
			Nonce:   1,
		})
		require.NoError(t, err)
		require.Empty(t, res)

		_, found := k.GetOutTxTracker(ctx, 1, 1)
		require.False(t, found)
	})
}
