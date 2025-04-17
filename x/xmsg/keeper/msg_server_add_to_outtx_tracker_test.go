package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/proofs"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/keeper"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func getEthereumChainID() int64 {
	return 5 // Goerli
}

// TODO: Add a test case with proof and Bitcoin chain

func TestMsgServer_AddToOutTxTracker(t *testing.T) {
	t.Run("admin can add tracker", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)

		chainID := getEthereumChainID()
		hash := sample.Hash().Hex()

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    hash,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     0,
		})
		require.NoError(t, err)
		tracker, found := k.GetOutTxTracker(ctx, chainID, 0)
		require.True(t, found)
		require.Equal(t, hash, tracker.HashLists[0].TxHash)
	})

	t.Run("observer can add tracker", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(true)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)

		chainID := getEthereumChainID()
		hash := sample.Hash().Hex()

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    hash,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     0,
		})
		require.NoError(t, err)
		tracker, found := k.GetOutTxTracker(ctx, chainID, 0)
		require.True(t, found)
		require.Equal(t, hash, tracker.HashLists[0].TxHash)
	})

	t.Run("can add hash to existing tracker", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)

		chainID := getEthereumChainID()
		existinghHash := sample.Hash().Hex()
		newHash := sample.Hash().Hex()

		k.SetOutTxTracker(ctx, types.OutTxTracker{
			ChainId: chainID,
			Nonce:   42,
			HashLists: []*types.TxHashList{
				{
					TxHash: existinghHash,
				},
			},
		})

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    newHash,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.NoError(t, err)
		tracker, found := k.GetOutTxTracker(ctx, chainID, 42)
		require.True(t, found)
		require.Len(t, tracker.HashLists, 2)
		require.EqualValues(t, existinghHash, tracker.HashLists[0].TxHash)
		require.EqualValues(t, newHash, tracker.HashLists[1].TxHash)
	})

	t.Run("should return early if xmsg not pending", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})

		// set xmsg status to outbound mined
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_OUTBOUND_MINED, false)

		chainID := getEthereumChainID()

		res, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    sample.Hash().Hex(),
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     0,
		})
		require.NoError(t, err)
		require.Equal(t, &types.MsgAddToOutTxTrackerResponse{IsRemoved: true}, res)

		// check if tracker is removed
		_, found := k.GetOutTxTracker(ctx, chainID, 0)
		require.False(t, found)
	})

	t.Run("should error for unsupported chain", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(nil)

		chainID := getEthereumChainID()

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    sample.Hash().Hex(),
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     0,
		})
		require.ErrorIs(t, err, observertypes.ErrSupportedChains)
	})

	t.Run("should error if no XmsgByNonce", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		observerMock := keepertest.GetXmsgObserverMock(t, k)

		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, true)

		chainID := getEthereumChainID()

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    sample.Hash().Hex(),
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     0,
		})
		require.ErrorIs(t, err, types.ErrCannotFindXmsg)
	})

	t.Run("should fail if max tracker hashes reached", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)

		hashes := make([]*types.TxHashList, keeper.MaxOutTxTrackerHashes)
		for i := 0; i < keeper.MaxOutTxTrackerHashes; i++ {
			hashes[i] = &types.TxHashList{
				TxHash: sample.Hash().Hex(),
			}
		}

		chainID := getEthereumChainID()
		newHash := sample.Hash().Hex()

		k.SetOutTxTracker(ctx, types.OutTxTracker{
			ChainId:   chainID,
			Nonce:     42,
			HashLists: hashes,
		})

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    newHash,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.ErrorIs(t, err, types.ErrMaxTxOutTrackerHashesReached)
	})

	t.Run("no hash added if already exist", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)

		chainID := getEthereumChainID()
		existinghHash := sample.Hash().Hex()

		k.SetOutTxTracker(ctx, types.OutTxTracker{
			ChainId: chainID,
			Nonce:   42,
			HashLists: []*types.TxHashList{
				{
					TxHash: existinghHash,
				},
			},
		})

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    existinghHash,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.NoError(t, err)
		tracker, found := k.GetOutTxTracker(ctx, chainID, 42)
		require.True(t, found)
		require.Len(t, tracker.HashLists, 1)
		require.EqualValues(t, existinghHash, tracker.HashLists[0].TxHash)
	})

	t.Run("can add tracker with proof", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseObserverMock:    true,
			UseLightclientMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		chainID := getEthereumChainID()
		ethTx, ethTxBytes, tssAddress := sample.EthTxSigned(t, chainID, sample.EthAddress(), 42)
		txHash := ethTx.Hash().Hex()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(&observertypes.QueryGetTssAddressResponse{
			Eth: tssAddress.Hex(),
		}, nil)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ethTxBytes, nil)

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.NoError(t, err)
		tracker, found := k.GetOutTxTracker(ctx, chainID, 42)
		require.True(t, found)
		require.EqualValues(t, txHash, tracker.HashLists[0].TxHash)
		require.True(t, tracker.HashLists[0].Proved)
	})

	t.Run("adding existing hash with proof make it proven", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseObserverMock:    true,
			UseLightclientMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		chainID := getEthereumChainID()
		ethTx, ethTxBytes, tssAddress := sample.EthTxSigned(t, chainID, sample.EthAddress(), 42)
		txHash := ethTx.Hash().Hex()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(&observertypes.QueryGetTssAddressResponse{
			Eth: tssAddress.Hex(),
		}, nil)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ethTxBytes, nil)

		k.SetOutTxTracker(ctx, types.OutTxTracker{
			ChainId: chainID,
			Nonce:   42,
			HashLists: []*types.TxHashList{
				{
					TxHash: sample.Hash().Hex(),
					Proved: false,
				},
				{
					TxHash: txHash,
					Proved: false,
				},
			},
		})

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.NoError(t, err)
		tracker, found := k.GetOutTxTracker(ctx, chainID, 42)
		require.True(t, found)
		require.Len(t, tracker.HashLists, 2)
		require.EqualValues(t, txHash, tracker.HashLists[1].TxHash)
		require.True(t, tracker.HashLists[1].Proved)
	})

	t.Run("should fail if verify proof fail", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseObserverMock:    true,
			UseLightclientMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		chainID := getEthereumChainID()
		ethTx, ethTxBytes, _ := sample.EthTxSigned(t, chainID, sample.EthAddress(), 42)
		txHash := ethTx.Hash().Hex()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ethTxBytes, errors.New("error"))

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.ErrorIs(t, err, types.ErrProofVerificationFail)
	})

	t.Run("should fail if no tss when adding hash with proof", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseObserverMock:    true,
			UseLightclientMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		chainID := getEthereumChainID()
		ethTx, ethTxBytes, tssAddress := sample.EthTxSigned(t, chainID, sample.EthAddress(), 42)
		txHash := ethTx.Hash().Hex()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ethTxBytes, nil)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(&observertypes.QueryGetTssAddressResponse{
			Eth: tssAddress.Hex(),
		}, errors.New("error"))

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.ErrorIs(t, err, observertypes.ErrTssNotFound)
	})

	t.Run("should fail if body verification fail with proof", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseObserverMock:    true,
			UseLightclientMock: true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		chainID := getEthereumChainID()
		ethTx, _, tssAddress := sample.EthTxSigned(t, chainID, sample.EthAddress(), 42)
		txHash := ethTx.Hash().Hex()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		keepertest.MockXmsgByNonce_pell(t, ctx, *k, observerMock, types.XmsgStatus_PENDING_OUTBOUND, false)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(&observertypes.QueryGetTssAddressResponse{
			Eth: tssAddress.Hex(),
		}, nil)

		// makes VerifyProof returning an invalid hash
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sample.Bytes(), nil)

		_, err := msgServer.AddToOutTxTracker(ctx, &types.MsgAddToOutTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
			Nonce:     42,
		})
		require.ErrorIs(t, err, types.ErrTxBodyVerificationFail)
	})
}
