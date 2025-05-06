package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/keeper"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

var externalChainList = chains.ExternalChainList()

func TestMsgServer_RemoveChainParams(t *testing.T) {
	t.Run("can update chain params", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
			UsePevmMock:      true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		// mock the authority keeper for authorization
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)

		chain1 := externalChainList[0].Id
		chain2 := externalChainList[1].Id
		chain3 := externalChainList[2].Id

		// set admin
		admin := sample.AccAddress()

		// add chain params
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				sample.ChainParams_pell(chain1),
				sample.ChainParams_pell(chain2),
				sample.ChainParams_pell(chain3),
			},
		})

		// remove chain params
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)
		_, err := srv.RemoveChainParams(sdk.WrapSDKContext(ctx), &types.MsgRemoveChainParams{
			Signer:  admin,
			ChainId: chain2,
		})
		require.NoError(t, err)

		// check list has two chain params
		chainParamsList, found := k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 2)
		require.Equal(t, chain1, chainParamsList.ChainParams[0].ChainId)
		require.Equal(t, chain3, chainParamsList.ChainParams[1].ChainId)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)
		// remove chain params
		_, err = srv.RemoveChainParams(sdk.WrapSDKContext(ctx), &types.MsgRemoveChainParams{
			Signer:  admin,
			ChainId: chain1,
		})
		require.NoError(t, err)

		// check list has one chain params
		chainParamsList, found = k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 1)
		require.Equal(t, chain3, chainParamsList.ChainParams[0].ChainId)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// remove chain params
		_, err = srv.RemoveChainParams(sdk.WrapSDKContext(ctx), &types.MsgRemoveChainParams{
			Signer:  admin,
			ChainId: chain3,
		})
		require.NoError(t, err)

		// check list has no chain params
		chainParamsList, found = k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 0)
	})

	t.Run("cannot remove chain params if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		_, err := srv.RemoveChainParams(sdk.WrapSDKContext(ctx), &types.MsgRemoveChainParams{
			Signer:  admin,
			ChainId: externalChainList[0].Id,
		})
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
	})

	t.Run("cannot remove if chain ID not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		// set admin
		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// not found if no chain params
		_, found := k.GetChainParamsList(ctx)
		require.False(t, found)

		_, err := srv.RemoveChainParams(sdk.WrapSDKContext(ctx), &types.MsgRemoveChainParams{
			Signer:  admin,
			ChainId: externalChainList[0].Id,
		})
		require.ErrorIs(t, err, types.ErrChainParamsNotFound)

		// add chain params
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				sample.ChainParams_pell(externalChainList[0].Id),
				sample.ChainParams_pell(externalChainList[1].Id),
				sample.ChainParams_pell(externalChainList[2].Id),
			},
		})

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// not found if chain ID not in list
		_, err = srv.RemoveChainParams(sdk.WrapSDKContext(ctx), &types.MsgRemoveChainParams{
			Signer:  admin,
			ChainId: externalChainList[3].Id,
		})
		require.ErrorIs(t, err, types.ErrChainParamsNotFound)
	})
}
