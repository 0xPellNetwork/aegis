package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestMsgServer_AbortStuckXmsg(t *testing.T) {
	t.Run("can abort a xmsg in pending inbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		msgServer := xmsgkeeper.NewMsgServerImpl(*k)
		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// create a xmsg
		xmsg := sample.Xmsg_pell(t, "xmsg_index")
		xmsg.XmsgStatus = &xmsgtypes.Status{
			Status:        xmsgtypes.XmsgStatus_PENDING_INBOUND,
			StatusMessage: "pending inbound",
		}
		k.SetXmsg(ctx, *xmsg)

		// abort the xmsg
		_, err := msgServer.AbortStuckXmsg(ctx, &xmsgtypes.MsgAbortStuckXmsg{
			Signer:    admin,
			XmsgIndex: sample.GetXmsgIndicesFromString_pell("xmsg_index"),
		})

		require.NoError(t, err)
		xmsgFound, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("xmsg_index"))
		require.True(t, found)
		require.Equal(t, xmsgtypes.XmsgStatus_ABORTED, xmsgFound.XmsgStatus.Status)
		require.Equal(t, xmsgkeeper.AbortMessage, xmsgFound.XmsgStatus.StatusMessage)
	})

	t.Run("can abort a xmsg in pending outbound", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		msgServer := xmsgkeeper.NewMsgServerImpl(*k)
		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// create a xmsg
		xmsg := sample.Xmsg_pell(t, "xmsg_index")
		xmsg.XmsgStatus = &xmsgtypes.Status{
			Status:        xmsgtypes.XmsgStatus_PENDING_OUTBOUND,
			StatusMessage: "pending outbound",
		}
		k.SetXmsg(ctx, *xmsg)

		// abort the xmsg
		_, err := msgServer.AbortStuckXmsg(ctx, &xmsgtypes.MsgAbortStuckXmsg{
			Signer:    admin,
			XmsgIndex: sample.GetXmsgIndicesFromString_pell("xmsg_index"),
		})

		require.NoError(t, err)
		xmsgFound, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("xmsg_index"))
		require.True(t, found)
		require.Equal(t, xmsgtypes.XmsgStatus_ABORTED, xmsgFound.XmsgStatus.Status)
		require.Equal(t, xmsgkeeper.AbortMessage, xmsgFound.XmsgStatus.StatusMessage)
	})

	t.Run("can abort a xmsg in pending revert", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		msgServer := xmsgkeeper.NewMsgServerImpl(*k)
		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// create a xmsg
		xmsg := sample.Xmsg_pell(t, "xmsg_index")
		xmsg.XmsgStatus = &xmsgtypes.Status{
			Status:        xmsgtypes.XmsgStatus_PENDING_REVERT,
			StatusMessage: "pending revert",
		}
		k.SetXmsg(ctx, *xmsg)

		// abort the xmsg
		_, err := msgServer.AbortStuckXmsg(ctx, &xmsgtypes.MsgAbortStuckXmsg{
			Signer:    admin,
			XmsgIndex: sample.GetXmsgIndicesFromString_pell("xmsg_index"),
		})

		require.NoError(t, err)
		xmsgFound, found := k.GetXmsg(ctx, sample.GetXmsgIndicesFromString_pell("xmsg_index"))
		require.True(t, found)
		require.Equal(t, xmsgtypes.XmsgStatus_ABORTED, xmsgFound.XmsgStatus.Status)
		require.Equal(t, xmsgkeeper.AbortMessage, xmsgFound.XmsgStatus.StatusMessage)
	})

	t.Run("cannot abort a xmsg in pending outbound if not admin", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		msgServer := xmsgkeeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		// create a xmsg
		xmsg := sample.Xmsg_pell(t, "xmsg_index")
		xmsg.XmsgStatus = &xmsgtypes.Status{
			Status:        xmsgtypes.XmsgStatus_PENDING_OUTBOUND,
			StatusMessage: "pending outbound",
		}
		k.SetXmsg(ctx, *xmsg)

		// abort the xmsg
		_, err := msgServer.AbortStuckXmsg(ctx, &xmsgtypes.MsgAbortStuckXmsg{
			Signer:    admin,
			XmsgIndex: sample.GetXmsgIndicesFromString_pell("xmsg_index"),
		})
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
	})

	t.Run("cannot abort a xmsg if doesn't exist", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		msgServer := xmsgkeeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// abort the xmsg
		_, err := msgServer.AbortStuckXmsg(ctx, &xmsgtypes.MsgAbortStuckXmsg{
			Signer:    admin,
			XmsgIndex: sample.GetXmsgIndicesFromString_pell("xmsg_index"),
		})
		require.ErrorIs(t, err, xmsgtypes.ErrCannotFindXmsg)
	})

	t.Run("cannot abort a xmsg if not pending", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})
		msgServer := xmsgkeeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// create a xmsg
		xmsg := sample.Xmsg_pell(t, "xmsg_index")
		xmsg.XmsgStatus = &xmsgtypes.Status{
			Status:        xmsgtypes.XmsgStatus_OUTBOUND_MINED,
			StatusMessage: "outbound mined",
		}
		k.SetXmsg(ctx, *xmsg)

		// abort the xmsg
		_, err := msgServer.AbortStuckXmsg(ctx, &xmsgtypes.MsgAbortStuckXmsg{
			Signer:    admin,
			XmsgIndex: sample.GetXmsgIndicesFromString_pell("xmsg_index"),
		})
		require.ErrorIs(t, err, xmsgtypes.ErrStatusNotPending)
	})
}
