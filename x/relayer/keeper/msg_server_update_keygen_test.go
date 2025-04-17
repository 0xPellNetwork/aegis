package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestMsgServer_UpdateKeygen(t *testing.T) {
	t.Run("should error if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		wctx := sdk.WrapSDKContext(ctx)

		srv := keeper.NewMsgServerImpl(*k)
		res, err := srv.UpdateKeygen(wctx, &types.MsgUpdateKeygen{
			Signer: admin,
		})
		require.Error(t, err)
		require.Equal(t, &types.MsgUpdateKeygenResponse{}, res)
	})

	t.Run("should error if keygen not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		wctx := sdk.WrapSDKContext(ctx)

		srv := keeper.NewMsgServerImpl(*k)
		res, err := srv.UpdateKeygen(wctx, &types.MsgUpdateKeygen{
			Signer: admin,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if msg block too low", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		wctx := sdk.WrapSDKContext(ctx)
		item := types.Keygen{
			BlockNumber: 10,
		}
		k.SetKeygen(ctx, item)
		srv := keeper.NewMsgServerImpl(*k)
		res, err := srv.UpdateKeygen(wctx, &types.MsgUpdateKeygen{
			Signer: admin,
			Block:  2,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should update", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		wctx := sdk.WrapSDKContext(ctx)
		item := types.Keygen{
			BlockNumber: 10,
		}
		k.SetKeygen(ctx, item)
		srv := keeper.NewMsgServerImpl(*k)

		granteePubKey := sample.PubKeySet()
		k.SetNodeAccount(ctx, types.NodeAccount{
			Operator:      "operator",
			GranteePubkey: granteePubKey,
		})

		res, err := srv.UpdateKeygen(wctx, &types.MsgUpdateKeygen{
			Signer: admin,
			Block:  ctx.BlockHeight() + 30,
		})
		require.NoError(t, err)
		require.Equal(t, &types.MsgUpdateKeygenResponse{}, res)

		keygen, found := k.GetKeygen(ctx)
		require.True(t, found)
		require.Equal(t, 1, len(keygen.GranteePubkeys))
		require.Equal(t, granteePubKey.Secp256k1.String(), keygen.GranteePubkeys[0])
		require.Equal(t, ctx.BlockHeight()+30, keygen.BlockNumber)
		require.Equal(t, types.KeygenStatus_PENDING, keygen.Status)
	})
}
