package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/gas"
	"github.com/pell-chain/pellcore/relayer/tss"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/pevm/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/keeper"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func setupTssMigrationParams(
	zk keepertest.PellKeepers,
	k *keeper.Keeper,
	ctx sdk.Context,
	chain chains.Chain,
	amount sdkmath.Uint,
	setNewTss bool,
	setCurrentTSS bool,
) (string, string) {
	zk.ObserverKeeper.SetCrosschainFlags(ctx, observertypes.CrosschainFlags{
		IsInboundEnabled:  false,
		IsOutboundEnabled: true,
	})

	zk.ObserverKeeper.SetChainParamsList(ctx, observertypes.ChainParamsList{
		ChainParams: []*observertypes.ChainParams{
			{
				ChainId:               chain.Id,
				BallotThreshold:       sdkmath.LegacyNewDec(0),
				MinObserverDelegation: sdkmath.LegacyOneDec(),
				IsSupported:           true,
			},
		},
	})

	currentTss := sample.Tss_pell()
	newTss := sample.Tss_pell()
	newTss.FinalizedPellHeight = currentTss.FinalizedPellHeight + 1
	newTss.KeygenPellHeight = currentTss.KeygenPellHeight + 1
	k.GetRelayerKeeper().SetTSS(ctx, currentTss)
	if setCurrentTSS {
		k.GetRelayerKeeper().SetTSSHistory(ctx, currentTss)
	}
	if setNewTss {
		k.GetRelayerKeeper().SetTSSHistory(ctx, newTss)
	}
	k.GetRelayerKeeper().SetPendingNonces(ctx, observertypes.PendingNonces{
		NonceLow:  1,
		NonceHigh: 1,
		ChainId:   chain.Id,
		Tss:       currentTss.TssPubkey,
	})
	k.SetGasPrice(ctx, xmsgtypes.GasPrice{
		Signer:      "",
		Index:       "",
		ChainId:     chain.Id,
		Signers:     nil,
		BlockNums:   nil,
		Prices:      []uint64{100000, 100000, 100000},
		MedianIndex: 1,
	})
	k.GetRelayerKeeper().SetChainNonces(ctx, observertypes.ChainNonces{
		Index:   chain.ChainName(),
		ChainId: chain.Id,
		Nonce:   1,
	})

	currentTssAddr, err := tss.GetTssAddrEVM(currentTss.TssPubkey)
	if err != nil {
		fmt.Println(err)
	}
	newTssAddr, err := tss.GetTssAddrEVM(newTss.TssPubkey)
	if err != nil {
		fmt.Println(err)
	}

	pellSent := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_PellSent{
			PellSent: &xmsgtypes.PellSent{
				TxOrigin:            "",
				Sender:              currentTssAddr.String(),
				ReceiverChainId:     chain.Id,
				Receiver:            newTssAddr.String(),
				Message:             "",
				PellParams:          types.Transfer.String(),
				PellValue:           amount,
				DestinationGasLimit: sdkmath.NewUint(1000000), // TODO: use const
			},
		},
	}

	senderChainID, err := chains.CosmosToEthChainID(ctx.ChainID())
	fmt.Println(senderChainID, err, ctx.ChainID())

	eventLog := &ethtypes.Log{}
	msg := xmsgtypes.NewMsgVoteOnObservedInboundTx(
		"",
		currentTssAddr.String(),
		int64(senderChainID),
		"",
		newTssAddr.String(),
		chain.Id,
		eventLog.TxHash.String(),
		eventLog.BlockNumber,
		0,
		eventLog.Index,
		pellSent,
	)

	xmsg, err := xmsgtypes.NewXmsg(ctx, *msg, currentTss.TssPubkey)
	if err != nil {
		fmt.Println(err)
	}

	return xmsg.Index, currentTss.TssPubkey
}

func TestKeeper_MigrateTSSFundsForChain(t *testing.T) {
	t.Run("test evm chain", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		indexString, _ := setupTssMigrationParams(zk, k, ctx, *chain, amount, true, true)
		gp, found := k.GetMedianGasPriceInUint(ctx, chain.Id)
		require.True(t, found)
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.NoError(t, err)
		xmsg, found := k.GetXmsg(ctx, indexString)
		require.True(t, found)
		multipliedValue, err := gas.MultiplyGasPrice(gp, xmsgtypes.TssMigrationGasMultiplierEVM)
		require.NoError(t, err)
		//t.Log(multipliedValue)
		require.Equal(t, multipliedValue.String(), xmsg.GetCurrentOutTxParam().OutboundTxGasPrice)
	})
}

func TestMsgServer_MigrateTssFunds(t *testing.T) {
	t.Run("should error if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, false)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("should error if inbound enabled", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("IsInboundEnabled", mock.Anything).Return(true)
		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("should error if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("IsInboundEnabled", mock.Anything).Return(false)
		observerMock.On("GetTSS", mock.Anything).Return(observertypes.TSS{}, false)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("should error if tss history empty", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("IsInboundEnabled", mock.Anything).Return(false)
		observerMock.On("GetTSS", mock.Anything).Return(sample.Tss_pell(), true)
		observerMock.On("GetAllTSS", mock.Anything).Return([]observertypes.TSS{})

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("should error if no new tss generated", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("IsInboundEnabled", mock.Anything).Return(false)
		tss := sample.Tss_pell()
		observerMock.On("GetTSS", mock.Anything).Return(tss, true)
		observerMock.On("GetAllTSS", mock.Anything).Return([]observertypes.TSS{tss})

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("should error if current tss is the latest", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("IsInboundEnabled", mock.Anything).Return(false)
		tss1 := sample.Tss_pell()
		tss1.FinalizedPellHeight = 2
		tss2 := sample.Tss_pell()
		tss2.FinalizedPellHeight = 1
		observerMock.On("GetTSS", mock.Anything).Return(tss1, true)
		observerMock.On("GetAllTSS", mock.Anything).Return([]observertypes.TSS{tss2})

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("should error if pending nonces not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("IsInboundEnabled", mock.Anything).Return(false)
		tss1 := sample.Tss_pell()
		tss1.FinalizedPellHeight = 2
		tss2 := sample.Tss_pell()
		tss2.FinalizedPellHeight = 3
		observerMock.On("GetTSS", mock.Anything).Return(tss1, true)
		observerMock.On("GetAllTSS", mock.Anything).Return([]observertypes.TSS{tss2})
		observerMock.On("GetPendingNonces", mock.Anything, mock.Anything, mock.Anything).Return(observertypes.PendingNonces{}, false)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.Error(t, err)
	})

	t.Run("successfully create tss migration xmsg", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		indexString, _ := setupTssMigrationParams(zk, k, ctx, *chain, amount, true, true)
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.NoError(t, err)
		_, found := k.GetXmsg(ctx, indexString)
		require.True(t, found)

		//Todo: check pelldata
	})

	t.Run("unable to migrate funds if new TSS is not created ", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		indexString, _ := setupTssMigrationParams(zk, k, ctx, *chain, amount, false, true)
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.ErrorContains(t, err, "no new tss address has been generated")
		hash := crypto.Keccak256Hash([]byte(indexString))
		index := hash.Hex()
		_, found := k.GetXmsg(ctx, index)
		require.False(t, found)
	})

	t.Run("unable to migrate funds when nonce low does not match nonce high", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		indexString, tssPubkey := setupTssMigrationParams(zk, k, ctx, *chain, amount, true, true)
		k.GetRelayerKeeper().SetPendingNonces(ctx, observertypes.PendingNonces{
			NonceLow:  1,
			NonceHigh: 10,
			ChainId:   chain.Id,
			Tss:       tssPubkey,
		})
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.ErrorIs(t, err, xmsgtypes.ErrCannotMigrateTssFunds)
		require.ErrorContains(t, err, "cannot migrate funds when there are pending nonces")
		hash := crypto.Keccak256Hash([]byte(indexString))
		index := hash.Hex()
		_, found := k.GetXmsg(ctx, index)
		require.False(t, found)
	})

	t.Run("unable to migrate funds when a pending xmsg is presnt in migration info", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		indexString, tssPubkey := setupTssMigrationParams(zk, k, ctx, *chain, amount, true, true)
		k.GetRelayerKeeper().SetPendingNonces(ctx, observertypes.PendingNonces{
			NonceLow:  1,
			NonceHigh: 1,
			ChainId:   chain.Id,
			Tss:       tssPubkey,
		})
		existingXmsg := sample.Xmsg_pell(t, "sample_index")
		existingXmsg.XmsgStatus.Status = xmsgtypes.XmsgStatus_PENDING_OUTBOUND
		k.SetXmsg(ctx, *existingXmsg)
		k.GetRelayerKeeper().SetFundMigrator(ctx, observertypes.TssFundMigratorInfo{
			ChainId:            chain.Id,
			MigrationXmsgIndex: existingXmsg.Index,
		})
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.ErrorIs(t, err, xmsgtypes.ErrCannotMigrateTssFunds)
		require.ErrorContains(t, err, "cannot migrate funds while there are pending migrations")
		hash := crypto.Keccak256Hash([]byte(indexString))
		index := hash.Hex()
		_, found := k.GetXmsg(ctx, index)
		require.False(t, found)
		_, found = k.GetXmsg(ctx, existingXmsg.Index)
		require.True(t, found)
	})

	t.Run("unable to migrate funds if current TSS is not present in TSSHistory and no new TSS has been generated", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
		})

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_ADMIN, true)

		msgServer := keeper.NewMsgServerImpl(*k)
		chain := getValidEthChain()
		amount := sdkmath.NewUintFromString("10000000000000000000")
		indexString, _ := setupTssMigrationParams(zk, k, ctx, *chain, amount, false, false)
		currentTss, found := k.GetRelayerKeeper().GetTSS(ctx)
		require.True(t, found)
		newTss := sample.Tss_pell()
		newTss.FinalizedPellHeight = currentTss.FinalizedPellHeight - 10
		newTss.KeygenPellHeight = currentTss.KeygenPellHeight - 10
		k.GetRelayerKeeper().SetTSSHistory(ctx, newTss)
		_, err := msgServer.MigrateTssFunds(ctx, &xmsgtypes.MsgMigrateTssFunds{
			Signer:  admin,
			ChainId: chain.Id,
			Amount:  amount,
		})
		require.ErrorIs(t, err, xmsgtypes.ErrCannotMigrateTssFunds)
		require.ErrorContains(t, err, "current tss is the latest")
		hash := crypto.Keccak256Hash([]byte(indexString))
		index := hash.Hex()
		_, found = k.GetXmsg(ctx, index)
		require.False(t, found)
	})
}
