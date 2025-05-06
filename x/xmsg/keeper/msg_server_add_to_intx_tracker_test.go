package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/coin"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestMsgServer_AddToInTxTracker(t *testing.T) {
	t.Run("fail normal user submit without proof", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		nonAdmin := sample.AccAddress()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, nonAdmin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)

		txHash := "string"
		chainID := getValidEthChainID()

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    nonAdmin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
		})
		require.ErrorIs(t, err, authoritytypes.ErrUnauthorized)
		_, found := k.GetInTxTracker(ctx, chainID, txHash)
		require.False(t, found)
	})

	t.Run("fail for unsupported chain id", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(nil)

		txHash := "string"
		chainID := getValidEthChainID()

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    sample.AccAddress(),
			ChainId:   chainID + 1,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
		})
		require.ErrorIs(t, err, relayertypes.ErrSupportedChains)
		_, found := k.GetInTxTracker(ctx, chainID, txHash)
		require.False(t, found)
	})

	t.Run("admin add tx tracker", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
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

		txHash := "string"
		chainID := getValidEthChainID()
		setSupportedChain(ctx, zk, chainID)

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
		})
		require.NoError(t, err)
		_, found := k.GetInTxTracker(ctx, chainID, txHash)
		require.True(t, found)
	})

	t.Run("observer add tx tracker", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock: true,
			UseObserverMock:  true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, mock.Anything, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(true)

		txHash := "string"
		chainID := getValidEthChainID()

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     nil,
			BlockHash: "",
			TxIndex:   0,
		})
		require.NoError(t, err)
		_, found := k.GetInTxTracker(ctx, chainID, txHash)
		require.True(t, found)
	})

	t.Run("fail if proof is provided but not verified", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseLightclientMock: true,
			UseObserverMock:    true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, mock.Anything, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("error"))

		txHash := "string"
		chainID := getValidEthChainID()

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
		})
		require.ErrorIs(t, err, types.ErrProofVerificationFail)
	})

	t.Run("fail if proof is provided but can't find chain params to verify body", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseLightclientMock: true,
			UseObserverMock:    true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, mock.Anything, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sample.Bytes(), nil)
		observerMock.On("GetChainParamsByChainID", mock.Anything, mock.Anything).Return(nil, false)

		txHash := "string"
		chainID := getValidEthChainID()

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
		})
		require.ErrorIs(t, err, types.ErrUnsupportedChain)
	})

	t.Run("fail if proof is provided but can't find tss to verify body", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseLightclientMock: true,
			UseObserverMock:    true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, mock.Anything, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sample.Bytes(), nil)
		observerMock.On("GetChainParamsByChainID", mock.Anything, mock.Anything).Return(sample.ChainParams_pell(chains.EthChain().Id), true)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(nil, errors.New("error"))

		txHash := "string"
		chainID := getValidEthChainID()
		setSupportedChain(ctx, zk, chainID)

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
		})
		require.ErrorIs(t, err, relayertypes.ErrTssNotFound)
	})

	t.Run("fail if proof is provided but error while verifying tx body", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseLightclientMock: true,
			UseObserverMock:    true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, mock.Anything, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		observerMock.On("GetChainParamsByChainID", mock.Anything, mock.Anything).Return(sample.ChainParams_pell(chains.EthChain().Id), true)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(&relayertypes.QueryGetTssAddressResponse{
			Eth: sample.EthAddress().Hex(),
		}, nil)

		// verifying the body will fail because the bytes are tried to be unmarshaled but they are not valid
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("invalid"), nil)

		txHash := "string"
		chainID := getValidEthChainID()
		setSupportedChain(ctx, zk, chainID)

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_PELL,
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
		})
		require.ErrorIs(t, err, types.ErrTxBodyVerificationFail)
	})

	t.Run("can add a in tx tracker with a proof", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseAuthorityMock:   true,
			UseLightclientMock: true,
			UseObserverMock:    true,
		})
		msgServer := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()

		chainID := chains.EthChain().Id
		tssAddress := sample.EthAddress()
		ethTx, ethTxBytes := sample.EthTx(t, chainID, tssAddress, 42)
		txHash := ethTx.Hash().Hex()

		authorityMock := keepertest.GetXmsgAuthorityMock(t, k)
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		lightclientMock := keepertest.GetXmsgLightclientMock(t, k)

		keepertest.MockIsAuthorized(&authorityMock.Mock, mock.Anything, authoritytypes.PolicyType_GROUP_EMERGENCY, false)
		observerMock.On("GetSupportedChainFromChainID", mock.Anything, mock.Anything).Return(&chains.Chain{})
		observerMock.On("IsNonTombstonedObserver", mock.Anything, mock.Anything).Return(false)
		chainParams := sample.ChainParams_pell(chains.EthChain().Id)
		chainParams.DelegationManagerContractAddress = tssAddress.Hex()
		observerMock.On("GetChainParamsByChainID", mock.Anything, mock.Anything).Return(chainParams, true)
		observerMock.On("GetTssAddress", mock.Anything, mock.Anything).Return(&relayertypes.QueryGetTssAddressResponse{
			Eth: tssAddress.Hex(),
		}, nil)
		lightclientMock.On("VerifyProof", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ethTxBytes, nil)

		_, err := msgServer.AddToInTxTracker(ctx, &types.MsgAddToInTxTracker{
			Signer:    admin,
			ChainId:   chainID,
			TxHash:    txHash,
			CoinType:  coin.CoinType_GAS, // use coin types gas: the receiver must be the tss address
			Proof:     &proofs.Proof{},
			BlockHash: "",
			TxIndex:   0,
		})
		require.NoError(t, err)
		_, found := k.GetInTxTracker(ctx, chainID, txHash)
		require.True(t, found)
	})
}
