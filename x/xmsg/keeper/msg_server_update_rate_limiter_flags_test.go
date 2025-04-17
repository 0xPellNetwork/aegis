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

func TestMsgServer_UpdateRateLimiterFlags(t *testing.T) {
	t.Run("can update rate limiter flags", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		_, found := k.GetRateLimiterFlags(ctx)
		require.False(t, found)

		flags := sample.RateLimiterFlags_pell()

		_, err := msgServer.UpdateRateLimiterFlags(ctx, types.NewMsgUpdateRateLimiterFlags(
			admin,
			flags,
		))
		require.NoError(t, err)

		storedFlags, found := k.GetRateLimiterFlags(ctx)
		require.True(t, found)
		require.Equal(t, flags, storedFlags)
	})

	t.Run("cannot update rate limiter flags if unauthorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)
		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		_, err := msgServer.UpdateRateLimiterFlags(ctx, types.NewMsgUpdateRateLimiterFlags(
			admin,
			sample.RateLimiterFlags_pell(),
		))
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
	})
}
