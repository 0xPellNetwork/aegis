package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestMsgServer_ResetChainNonces(t *testing.T) {
	t.Run("cannot reset chain nonces if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)
		chainId := chains.GoerliLocalnetChain().Id

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		_, err := srv.ResetChainNonces(sdk.WrapSDKContext(ctx), &types.MsgResetChainNonces{
			Signer:         admin,
			ChainId:        chainId,
			ChainNonceLow:  1,
			ChainNonceHigh: 5,
		})
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
	})

	t.Run("cannot reset chain nonces if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		chainId := chains.GoerliLocalnetChain().Id
		_, err := srv.ResetChainNonces(sdk.WrapSDKContext(ctx), &types.MsgResetChainNonces{
			Signer:         admin,
			ChainId:        chainId,
			ChainNonceLow:  1,
			ChainNonceHigh: 5,
		})
		require.ErrorIs(t, err, types.ErrTssNotFound)
	})

	t.Run("cannot reset chain nonces if chain not supported", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)
		tss := sample.Tss_pell()
		k.SetTSS(ctx, tss)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		_, err := srv.ResetChainNonces(sdk.WrapSDKContext(ctx), &types.MsgResetChainNonces{
			Signer:         admin,
			ChainId:        999,
			ChainNonceLow:  1,
			ChainNonceHigh: 5,
		})
		require.ErrorIs(t, err, types.ErrSupportedChains)
	})

	t.Run("can reset chain nonces", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)
		tss := sample.Tss_pell()
		k.SetTSS(ctx, tss)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		chainId := chains.GoerliLocalnetChain().Id
		index := chains.GoerliLocalnetChain().ChainName()

		// check existing chain nonces
		_, found := k.GetChainNonces(ctx, index)
		require.False(t, found)
		_, found = k.GetPendingNonces(ctx, tss.TssPubkey, chainId)
		require.False(t, found)

		// reset chain nonces
		nonceLow := 1
		nonceHigh := 5
		_, err := srv.ResetChainNonces(sdk.WrapSDKContext(ctx), &types.MsgResetChainNonces{
			Signer:         admin,
			ChainId:        chainId,
			ChainNonceLow:  int64(nonceLow),
			ChainNonceHigh: int64(nonceHigh),
		})
		require.NoError(t, err)

		// check updated chain nonces
		chainNonces, found := k.GetChainNonces(ctx, index)
		require.True(t, found)
		require.Equal(t, chainId, chainNonces.ChainId)
		require.Equal(t, index, chainNonces.Index)
		require.Equal(t, uint64(nonceHigh), chainNonces.Nonce)

		pendingNonces, found := k.GetPendingNonces(ctx, tss.TssPubkey, chainId)
		require.True(t, found)
		require.Equal(t, chainId, pendingNonces.ChainId)
		require.Equal(t, tss.TssPubkey, pendingNonces.Tss)
		require.Equal(t, int64(nonceLow), pendingNonces.NonceLow)
		require.Equal(t, int64(nonceHigh), pendingNonces.NonceHigh)

		// reset nonces back to 0
		_, err = srv.ResetChainNonces(sdk.WrapSDKContext(ctx), &types.MsgResetChainNonces{
			Signer:         admin,
			ChainId:        chainId,
			ChainNonceLow:  0,
			ChainNonceHigh: 0,
		})
		require.NoError(t, err)

		// check updated chain nonces
		chainNonces, found = k.GetChainNonces(ctx, index)
		require.True(t, found)
		require.Equal(t, chainId, chainNonces.ChainId)
		require.Equal(t, index, chainNonces.Index)
		require.Equal(t, uint64(0), chainNonces.Nonce)

		pendingNonces, found = k.GetPendingNonces(ctx, tss.TssPubkey, chainId)
		require.True(t, found)
		require.Equal(t, chainId, pendingNonces.ChainId)
		require.Equal(t, tss.TssPubkey, pendingNonces.Tss)
		require.Equal(t, int64(0), pendingNonces.NonceLow)
		require.Equal(t, int64(0), pendingNonces.NonceHigh)
	})
}
