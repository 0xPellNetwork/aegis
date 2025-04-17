package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/keeper"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestMsgServer_UpdateTssAddress(t *testing.T) {
	t.Run("should fail if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, false)

		msgServer := keeper.NewMsgServerImpl(*k)

		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: "",
		})
		require.Error(t, err)
	})

	t.Run("should fail if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)

		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: "",
		})
		require.Error(t, err)
	})

	t.Run("successfully update tss address", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		tssOld := sample.Tss_pell()
		tssNew := sample.Tss_pell()
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssOld)
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssNew)
		k.GetRelayerKeeper().SetTSS(ctx, tssOld)
		for _, chain := range k.GetRelayerKeeper().GetSupportedChains(ctx) {
			index := chain.ChainName() + "_migration_tx_index"
			k.GetRelayerKeeper().SetFundMigrator(ctx, types.TssFundMigratorInfo{
				ChainId:            chain.Id,
				MigrationXmsgIndex: sample.GetXmsgIndicesFromString_pell(index),
			})
			xmsg := sample.Xmsg_pell(t, index)
			xmsg.XmsgStatus.Status = xmsgtypes.XmsgStatus_OUTBOUND_MINED
			k.SetXmsg(ctx, *xmsg)
		}
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: tssNew.TssPubkey,
		})
		require.NoError(t, err)
		tss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		require.Equal(t, tssNew, tss)
		migrators := k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)
		require.Equal(t, 0, len(migrators))
	})

	t.Run("new tss has not been added to tss history", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		tssOld := sample.Tss_pell()
		tssNew := sample.Tss_pell()
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssOld)
		k.GetRelayerKeeper().SetTSS(ctx, tssOld)
		for _, chain := range k.GetRelayerKeeper().GetSupportedChains(ctx) {
			index := chain.ChainName() + "_migration_tx_index"
			k.GetRelayerKeeper().SetFundMigrator(ctx, types.TssFundMigratorInfo{
				ChainId:            chain.Id,
				MigrationXmsgIndex: sample.GetXmsgIndicesFromString_pell(index),
			})
			xmsg := sample.Xmsg_pell(t, index)
			xmsg.XmsgStatus.Status = xmsgtypes.XmsgStatus_OUTBOUND_MINED
			k.SetXmsg(ctx, *xmsg)
		}
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: tssNew.TssPubkey,
		})
		require.ErrorContains(t, err, "tss pubkey has not been generated")
		require.ErrorIs(t, err, xmsgtypes.ErrUnableToUpdateTss)
		tss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		require.Equal(t, tssOld, tss)
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
	})

	t.Run("old tss pubkey provided", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		tssOld := sample.Tss_pell()
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssOld)
		k.GetRelayerKeeper().SetTSS(ctx, tssOld)
		for _, chain := range k.GetRelayerKeeper().GetSupportedChains(ctx) {
			index := chain.ChainName() + "_migration_tx_index"
			k.GetRelayerKeeper().SetFundMigrator(ctx, types.TssFundMigratorInfo{
				ChainId:            chain.Id,
				MigrationXmsgIndex: sample.GetXmsgIndicesFromString_pell(index),
			})
			xmsg := sample.Xmsg_pell(t, index)
			xmsg.XmsgStatus.Status = xmsgtypes.XmsgStatus_OUTBOUND_MINED
			k.SetXmsg(ctx, *xmsg)
		}
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: tssOld.TssPubkey,
		})
		require.ErrorContains(t, err, "no new tss address has been generated")
		require.ErrorIs(t, err, xmsgtypes.ErrUnableToUpdateTss)
		tss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		require.Equal(t, tssOld, tss)
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
	})

	t.Run("unable to update tss when not enough migrators are present", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		tssOld := sample.Tss_pell()
		tssNew := sample.Tss_pell()

		k.GetRelayerKeeper().SetTSSHistory(ctx, tssOld)
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssNew)
		k.GetRelayerKeeper().SetTSS(ctx, tssOld)
		setSupportedChain(ctx, zk, getValidEthChainIDWithIndex(t, 0), getValidEthChainIDWithIndex(t, 1))

		// set a single migrator while there are 2 supported chains
		chain := k.GetRelayerKeeper().GetSupportedChains(ctx)[0]
		index := chain.ChainName() + "_migration_tx_index"
		k.GetRelayerKeeper().SetFundMigrator(ctx, types.TssFundMigratorInfo{
			ChainId:            chain.Id,
			MigrationXmsgIndex: sample.GetXmsgIndicesFromString_pell(index),
		})
		xmsg := sample.Xmsg_pell(t, index)
		xmsg.XmsgStatus.Status = xmsgtypes.XmsgStatus_OUTBOUND_MINED
		k.SetXmsg(ctx, *xmsg)

		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), 1)
		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: tssNew.TssPubkey,
		})
		require.ErrorContains(t, err, "cannot update tss address not enough migrations have been created and completed")
		require.ErrorIs(t, err, xmsgtypes.ErrUnableToUpdateTss)
		tss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		require.Equal(t, tssOld, tss)
		migrators := k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)
		require.Equal(t, 1, len(migrators))
	})

	t.Run("unable to update tss when pending xmsg is present", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		tssOld := sample.Tss_pell()
		tssNew := sample.Tss_pell()

		k.GetRelayerKeeper().SetTSSHistory(ctx, tssOld)
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssNew)
		k.GetRelayerKeeper().SetTSS(ctx, tssOld)
		setSupportedChain(ctx, zk, getValidEthChainIDWithIndex(t, 0), getValidEthChainIDWithIndex(t, 1))

		for _, chain := range k.GetRelayerKeeper().GetSupportedChains(ctx) {
			index := chain.ChainName() + "_migration_tx_index"
			k.GetRelayerKeeper().SetFundMigrator(ctx, types.TssFundMigratorInfo{
				ChainId:            chain.Id,
				MigrationXmsgIndex: sample.GetXmsgIndicesFromString_pell(index),
			})
			xmsg := sample.Xmsg_pell(t, index)
			xmsg.XmsgStatus.Status = xmsgtypes.XmsgStatus_PENDING_OUTBOUND
			k.SetXmsg(ctx, *xmsg)
		}
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: tssNew.TssPubkey,
		})
		require.ErrorContains(t, err, "cannot update tss address while there are pending migrations")
		require.ErrorIs(t, err, xmsgtypes.ErrUnableToUpdateTss)
		tss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		require.Equal(t, tssOld, tss)
		migrators := k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)
		require.Equal(t, len(k.GetRelayerKeeper().GetSupportedChains(ctx)), len(migrators))
	})

	t.Run("unable to update tss xmsg is not present", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		tssOld := sample.Tss_pell()
		tssNew := sample.Tss_pell()

		k.GetRelayerKeeper().SetTSSHistory(ctx, tssOld)
		k.GetRelayerKeeper().SetTSSHistory(ctx, tssNew)
		k.GetRelayerKeeper().SetTSS(ctx, tssOld)
		setSupportedChain(ctx, zk, getValidEthChainIDWithIndex(t, 0), getValidEthChainIDWithIndex(t, 1))

		for _, chain := range k.GetRelayerKeeper().GetSupportedChains(ctx) {
			index := chain.ChainName() + "_migration_tx_index"
			k.GetRelayerKeeper().SetFundMigrator(ctx, types.TssFundMigratorInfo{
				ChainId:            chain.Id,
				MigrationXmsgIndex: sample.GetXmsgIndicesFromString_pell(index),
			})
		}
		require.Equal(t, len(k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)), len(k.GetRelayerKeeper().GetSupportedChains(ctx)))
		_, err := msgServer.UpdateTssAddress(ctx, &xmsgtypes.MsgUpdateTssAddress{
			Signer:    admin,
			TssPubkey: tssNew.TssPubkey,
		})
		require.ErrorContains(t, err, "migration cross chain tx not found")
		require.ErrorIs(t, err, xmsgtypes.ErrUnableToUpdateTss)
		tss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		require.Equal(t, tssOld, tss)
		migrators := k.GetRelayerKeeper().GetAllTssFundMigrators(ctx)
		require.Equal(t, len(k.GetRelayerKeeper().GetSupportedChains(ctx)), len(migrators))
	})
}
