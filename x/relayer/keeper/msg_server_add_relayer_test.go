package keeper_test

import (
	"math"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestMsgServer_AddObserver(t *testing.T) {
	t.Run("should error if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)
		wctx := sdk.WrapSDKContext(ctx)

		srv := keeper.NewMsgServerImpl(*k)
		res, err := srv.AddObserver(wctx, &types.MsgAddObserver{
			Signer: admin,
		})
		require.Error(t, err)
		require.Equal(t, &types.MsgAddObserverResponse{}, res)
	})

	t.Run("should error if pub key not valid", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)
		wctx := sdk.WrapSDKContext(ctx)

		srv := keeper.NewMsgServerImpl(*k)
		res, err := srv.AddObserver(wctx, &types.MsgAddObserver{
			Signer:                  admin,
			PellclientGranteePubkey: "invalid",
		})
		require.Error(t, err)
		require.Equal(t, &types.MsgAddObserverResponse{}, res)
	})

	t.Run("should add if add node account only false", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)
		wctx := sdk.WrapSDKContext(ctx)

		_, found := k.GetLastObserverCount(ctx)
		require.False(t, found)
		srv := keeper.NewMsgServerImpl(*k)
		observerAddress := sdk.AccAddress(crypto.AddressHash([]byte("ObserverAddress")))
		res, err := srv.AddObserver(wctx, &types.MsgAddObserver{
			Signer:                  admin,
			PellclientGranteePubkey: sample.PubKeyString(),
			AddNodeAccountOnly:      false,
			ObserverAddress:         observerAddress.String(),
		})
		require.NoError(t, err)
		require.Equal(t, &types.MsgAddObserverResponse{}, res)

		loc, found := k.GetLastObserverCount(ctx)
		require.True(t, found)
		require.Equal(t, uint64(1), loc.Count)
	})

	t.Run("should add to node account if add node account only true", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		admin := sample.AccAddress()
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)
		wctx := sdk.WrapSDKContext(ctx)

		_, found := k.GetLastObserverCount(ctx)
		require.False(t, found)
		srv := keeper.NewMsgServerImpl(*k)
		observerAddress := sdk.AccAddress(crypto.AddressHash([]byte("ObserverAddress")))
		_, found = k.GetKeygen(ctx)
		require.False(t, found)
		_, found = k.GetNodeAccount(ctx, observerAddress.String())
		require.False(t, found)

		res, err := srv.AddObserver(wctx, &types.MsgAddObserver{
			Signer:                  admin,
			PellclientGranteePubkey: sample.PubKeyString(),
			AddNodeAccountOnly:      true,
			ObserverAddress:         observerAddress.String(),
		})
		require.NoError(t, err)
		require.Equal(t, &types.MsgAddObserverResponse{}, res)

		_, found = k.GetLastObserverCount(ctx)
		require.False(t, found)

		keygen, found := k.GetKeygen(ctx)
		require.True(t, found)
		require.Equal(t, types.Keygen{BlockNumber: math.MaxInt64}, keygen)

		_, found = k.GetNodeAccount(ctx, observerAddress.String())
		require.True(t, found)
	})
}
