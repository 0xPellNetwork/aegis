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

func TestMsgServer_UpsertChainParams(t *testing.T) {
	t.Run("can update chain params", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
			UsePevmMock:      true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		chain1 := externalChainList[0].Id
		chain2 := externalChainList[1].Id
		chain3 := externalChainList[2].Id

		// set admin
		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)

		// check list initially empty
		_, found := k.GetChainParamsList(ctx)
		require.False(t, found)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// a new chain params can be added
		chainParams1 := sample.ChainParams_pell(chain1)
		_, err := srv.UpsertChainParams(sdk.WrapSDKContext(ctx), &types.MsgUpsertChainParams{
			Signer:      admin,
			ChainParams: chainParams1,
		})
		require.NoError(t, err)

		// check list has one chain params
		chainParamsList, found := k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 1)
		require.Equal(t, chainParams1, chainParamsList.ChainParams[0])

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// a new chian params can be added
		chainParams2 := sample.ChainParams_pell(chain2)
		_, err = srv.UpsertChainParams(sdk.WrapSDKContext(ctx), &types.MsgUpsertChainParams{
			Signer:      admin,
			ChainParams: chainParams2,
		})
		require.NoError(t, err)

		// check list has two chain params
		chainParamsList, found = k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 2)
		require.Equal(t, chainParams1, chainParamsList.ChainParams[0])
		require.Equal(t, chainParams2, chainParamsList.ChainParams[1])

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// a new chain params can be added
		chainParams3 := sample.ChainParams_pell(chain3)
		_, err = srv.UpsertChainParams(sdk.WrapSDKContext(ctx), &types.MsgUpsertChainParams{
			Signer:      admin,
			ChainParams: chainParams3,
		})
		require.NoError(t, err)

		// check list has three chain params
		chainParamsList, found = k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 3)
		require.Equal(t, chainParams1, chainParamsList.ChainParams[0])
		require.Equal(t, chainParams2, chainParamsList.ChainParams[1])
		require.Equal(t, chainParams3, chainParamsList.ChainParams[2])

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// chain params can be updated
		chainParams2.ConfirmationCount = chainParams2.ConfirmationCount + 1
		_, err = srv.UpsertChainParams(sdk.WrapSDKContext(ctx), &types.MsgUpsertChainParams{
			Signer:      admin,
			ChainParams: chainParams2,
		})
		require.NoError(t, err)

		// check list has three chain params
		chainParamsList, found = k.GetChainParamsList(ctx)
		require.True(t, found)
		require.Len(t, chainParamsList.ChainParams, 3)
		require.Equal(t, chainParams1, chainParamsList.ChainParams[0])
		require.Equal(t, chainParams2, chainParamsList.ChainParams[1])
		require.Equal(t, chainParams3, chainParamsList.ChainParams[2])
	})

	t.Run("cannot update chain params if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		_, err := srv.UpsertChainParams(sdk.WrapSDKContext(ctx), &types.MsgUpsertChainParams{
			Signer:      admin,
			ChainParams: sample.ChainParams_pell(externalChainList[0].Id),
		})
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
	})
}
