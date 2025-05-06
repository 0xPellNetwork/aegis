package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	xmsgmocks "github.com/0xPellNetwork/aegis/testutil/keeper/mocks/xmsg"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

type XmsgMockOptions struct {
	UseBankMock        bool
	UseAccountMock     bool
	UseStakingMock     bool
	UseObserverMock    bool
	UsePevmMock        bool
	UseAuthorityMock   bool
	UseLightclientMock bool
}

var (
	XmsgMocksAll = XmsgMockOptions{
		UseBankMock:        true,
		UseAccountMock:     true,
		UseStakingMock:     true,
		UseObserverMock:    true,
		UsePevmMock:        true,
		UseAuthorityMock:   true,
		UseLightclientMock: true,
	}
	XmsgNoMocks = XmsgMockOptions{}
)

// XmsgKeeperWithMocks initializes a xmsg keeper for testing purposes with option to mock specific keepers
func XmsgKeeperWithMocks(
	t testing.TB,
	mockOptions XmsgMockOptions,
) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	SetConfig(false)
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Initialize local store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cdc := NewCodec()
	amino := codec.NewLegacyAmino()
	// Create regular keepers
	sdkKeepers := NewSDKKeepers(cdc, amino, db, stateStore)

	// Create pell keepers
	authorityKeeperTmp := initAuthorityKeeper(cdc, db, stateStore)
	lightclientKeeperTmp := initLightclientKeeper(cdc, db, stateStore, authorityKeeperTmp)
	observerKeeperTmp := initRelayerKeeper(
		cdc,
		db,
		stateStore,
		sdkKeepers.StakingKeeper,
		sdkKeepers.SlashingKeeper,
		sdkKeepers.ParamsKeeper,
		authorityKeeperTmp,
		lightclientKeeperTmp,
	)
	pevmKeeperTmp := initPevmKeeper(
		cdc,
		db,
		stateStore,
		sdkKeepers.AuthKeeper,
		sdkKeepers.BankKeeper,
		sdkKeepers.EvmKeeper,
		observerKeeperTmp,
		authorityKeeperTmp,
	)
	pellKeepers := PellKeepers{
		ObserverKeeper:  observerKeeperTmp,
		PevmKeeper:      pevmKeeperTmp,
		AuthorityKeeper: &authorityKeeperTmp,
	}
	var lightclientKeeper types.LightclientKeeper = lightclientKeeperTmp
	var authorityKeeper types.AuthorityKeeper = authorityKeeperTmp
	var observerKeeper types.RelayerKeeper = observerKeeperTmp
	var pevmKeeper types.PevmKeeper = pevmKeeperTmp

	// Create the fungible keeper
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := NewContext(stateStore)

	// Initialize modules genesis
	sdkKeepers.InitGenesis(ctx)
	pellKeepers.InitGenesis(ctx)

	// Add a proposer to the context
	ctx = sdkKeepers.InitBlockProposer(t, ctx)

	// Initialize mocks for mocked keepers
	var authKeeper types.AccountKeeper = sdkKeepers.AuthKeeper
	var bankKeeper types.BankKeeper = sdkKeepers.BankKeeper
	var stakingKeeper types.StakingKeeper = sdkKeepers.StakingKeeper
	if mockOptions.UseAccountMock {
		authKeeper = xmsgmocks.NewXmsgAccountKeeper(t)
	}
	if mockOptions.UseBankMock {
		bankKeeper = xmsgmocks.NewXmsgBankKeeper(t)
	}
	if mockOptions.UseStakingMock {
		stakingKeeper = xmsgmocks.NewXmsgStakingKeeper(t)
	}

	if mockOptions.UseAuthorityMock {
		authorityKeeper = xmsgmocks.NewXmsgAuthorityKeeper(t)
	}
	if mockOptions.UseObserverMock {
		observerKeeper = xmsgmocks.NewXmsgRelayerKeeper(t)
	}
	if mockOptions.UsePevmMock {
		pevmKeeper = xmsgmocks.NewXmsgPevmKeeper(t)
	}
	if mockOptions.UseLightclientMock {
		lightclientKeeper = xmsgmocks.NewXmsgLightclientKeeper(t)
	}

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		stakingKeeper,
		authKeeper,
		bankKeeper,
		observerKeeper,
		pevmKeeper,
		authorityKeeper,
		lightclientKeeper,
	)

	return k, ctx, sdkKeepers, pellKeepers
}

// XmsgKeeperAllMocks initializes a xmsg keeper for testing purposes with all mocks
func XmsgKeeperAllMocks(t testing.TB) (*keeper.Keeper, sdk.Context) {
	k, ctx, _, _ := XmsgKeeperWithMocks(t, XmsgMocksAll)
	return k, ctx
}

// XmsgKeeper initializes a xmsg keeper for testing purposes
func XmsgKeeper(t testing.TB) (*keeper.Keeper, sdk.Context, SDKKeepers, PellKeepers) {
	return XmsgKeeperWithMocks(t, XmsgNoMocks)
}

// GetXmsgLightclientMock returns a new xmsg lightclient keeper mock
func GetXmsgLightclientMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgLightclientKeeper {
	lk, ok := keeper.GetLightclientKeeper().(*xmsgmocks.XmsgLightclientKeeper)
	require.True(t, ok)
	return lk
}

// GetXmsgAuthorityMock returns a new xmsg authority keeper mock
func GetXmsgAuthorityMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgAuthorityKeeper {
	cok, ok := keeper.GetAuthorityKeeper().(*xmsgmocks.XmsgAuthorityKeeper)
	require.True(t, ok)
	return cok
}

func GetXmsgAccountMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgAccountKeeper {
	cak, ok := keeper.GetAuthKeeper().(*xmsgmocks.XmsgAccountKeeper)
	require.True(t, ok)
	return cak
}

func GetXmsgBankMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgBankKeeper {
	cbk, ok := keeper.GetBankKeeper().(*xmsgmocks.XmsgBankKeeper)
	require.True(t, ok)
	return cbk
}

func GetXmsgStakingMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgStakingKeeper {
	csk, ok := keeper.GetStakingKeeper().(*xmsgmocks.XmsgStakingKeeper)
	require.True(t, ok)
	return csk
}

func GetXmsgObserverMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgRelayerKeeper {
	cok, ok := keeper.GetRelayerKeeper().(*xmsgmocks.XmsgRelayerKeeper)
	require.True(t, ok)
	return cok
}

func GetXmsgPevmMock(t testing.TB, keeper *keeper.Keeper) *xmsgmocks.XmsgPevmKeeper {
	cfk, ok := keeper.GetPevmKeeper().(*xmsgmocks.XmsgPevmKeeper)
	require.True(t, ok)
	return cfk
}

func MockGetSupportedChainFromChainID_pell(m *xmsgmocks.XmsgRelayerKeeper, senderChain *chains.Chain) {
	m.On("GetSupportedChainFromChainID", mock.Anything, senderChain.Id).
		Return(senderChain).Once()
}

func MockProcessFailedOutboundForPEVMTx_pell(m *xmsgmocks.XmsgPevmKeeper, ctx sdk.Context, xmsg *types.Xmsg) {
	indexBytes, _ := xmsg.GetXmsgIndicesBytes()
	m.On("PELLRevertAndCallContract",
		mock.Anything,
		ethcommon.HexToAddress(xmsg.GetInboundTxParams().Sender),
		ethcommon.HexToAddress(xmsg.GetCurrentOutTxParam().Receiver),
		xmsg.GetInboundTxParams().SenderChainId,
		xmsg.GetCurrentOutTxParam().ReceiverChainId,
		indexBytes,
	).Return(nil, nil)
}

func MockPayGasAndUpdateXmsg_pell(m *xmsgmocks.XmsgPevmKeeper, m2 *xmsgmocks.XmsgRelayerKeeper, ctx sdk.Context, k keeper.Keeper, senderChain chains.Chain) {
	m2.On("GetSupportedChainFromChainID", mock.Anything, senderChain.Id).
		Return(&senderChain).Twice()
	// m.On("QueryGasLimit", mock.Anything, mock.Anything).
	// 	Return(big.NewInt(100), nil).Once()
	k.SetGasPrice(ctx, types.GasPrice{
		ChainId:     senderChain.Id,
		MedianIndex: 0,
		Prices:      []uint64{1},
	})
}

func MockUpdateNonce_pell(m *xmsgmocks.XmsgRelayerKeeper, senderChain chains.Chain) (nonce uint64) {
	nonce = uint64(1)
	tss := sample.Tss_pell()
	m.On("GetSupportedChainFromChainID", mock.Anything, senderChain.Id).
		Return(senderChain)
	m.On("GetChainNonces", mock.Anything, senderChain.ChainName()).
		Return(relayertypes.ChainNonces{Nonce: nonce}, true)
	m.On("GetTSS", mock.Anything).
		Return(tss, true)
	m.On("GetPendingNonces", mock.Anything, tss.TssPubkey, mock.Anything).
		Return(relayertypes.PendingNonces{NonceHigh: int64(nonce)}, true)
	m.On("SetChainNonces", mock.Anything, mock.Anything)
	m.On("SetPendingNonces", mock.Anything, mock.Anything)
	return
}

func MockRevertForHandleEVMEvents_pell(m *xmsgmocks.XmsgPevmKeeper, senderChainID int64, err error) {
	m.On("GetPellDelegationManagerProxyContractAddress", mock.Anything).Return(sample.EthAddress(), nil)
	m.On("CallSyncDelegatedStateOnPellDelegationManager",
		mock.Anything,                         // types.Context
		mock.AnythingOfType("[]uint8"),        // Chain ID as []uint8
		senderChainID,                         // Height as int64
		mock.AnythingOfType("common.Address"), // Staker
		mock.AnythingOfType("common.Address"), // Operator
	).Return(&evmtypes.MsgEthereumTxResponse{VmError: "execution reverted"}, false, err)
}

func MockVoteOnOutboundSuccessBallot_pell(m *xmsgmocks.XmsgRelayerKeeper, ctx sdk.Context, xmsg *types.Xmsg, senderChain chains.Chain, observer string) {
	m.On("VoteOnOutboundBallot", ctx, mock.Anything, xmsg.GetCurrentOutTxParam().ReceiverChainId, chains.ReceiveStatus_SUCCESS, observer).
		Return(true, true, relayertypes.Ballot{BallotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION}, senderChain.ChainName(), nil).Once()
}

func MockVoteOnOutboundFailedBallot_pell(m *xmsgmocks.XmsgRelayerKeeper, ctx sdk.Context, xmsg *types.Xmsg, senderChain chains.Chain, observer string) {
	m.On("VoteOnOutboundBallot", ctx, mock.Anything, xmsg.GetCurrentOutTxParam().ReceiverChainId, chains.ReceiveStatus_FAILED, observer).
		Return(true, true, relayertypes.Ballot{BallotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION}, senderChain.ChainName(), nil).Once()
}

func MockGetOutBound_pell(m *xmsgmocks.XmsgRelayerKeeper, ctx sdk.Context) {
	m.On("GetTSS", ctx).Return(relayertypes.TSS{}, true).Once()
}

func MockSaveOutBound_pell(m *xmsgmocks.XmsgRelayerKeeper, ctx sdk.Context, xmsg *types.Xmsg, tss relayertypes.TSS) {
	m.On("RemoveFromPendingNonces",
		ctx, tss.TssPubkey, xmsg.GetCurrentOutTxParam().ReceiverChainId, mock.Anything).
		Return().Once()
	m.On("GetTSS", ctx).Return(relayertypes.TSS{}, true)
}

func MockSaveOutBoundNewRevertCreated_pell(m *xmsgmocks.XmsgRelayerKeeper, ctx sdk.Context, xmsg *types.Xmsg, tss relayertypes.TSS) {
	m.On("RemoveFromPendingNonces",
		ctx, tss.TssPubkey, xmsg.GetCurrentOutTxParam().ReceiverChainId, mock.Anything).
		Return().Once()
	m.On("GetTSS", ctx).Return(relayertypes.TSS{}, true)
	m.On("SetNonceToXmsg", mock.Anything, mock.Anything).Return().Once()
}

// MockXmsgByNonce is a utility function using observer mock to returns a xmsg of the given status from xmsg keeper
// mocks the methods called by XmsgByNonce to directly return the given xmsg or error
func MockXmsgByNonce_pell(
	t *testing.T,
	ctx sdk.Context,
	k keeper.Keeper,
	observerKeeper *xmsgmocks.XmsgRelayerKeeper,
	xmsgStatus types.XmsgStatus,
	isErr bool,
) {
	if isErr {
		// return error on GetTSS to make XmsgByNonce return error
		observerKeeper.On("GetTSS", mock.Anything).Return(relayertypes.TSS{}, false).Once()
		return
	}

	xmsg := sample.Xmsg_pell(t, sample.StringRandom(sample.Rand(), 10))
	xmsg.XmsgStatus = &types.Status{
		Status: xmsgStatus,
	}
	k.SetXmsg(ctx, *xmsg)

	observerKeeper.On("GetTSS", mock.Anything).Return(relayertypes.TSS{}, true).Once()
	observerKeeper.On("GetNonceToXmsg", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(relayertypes.NonceToXmsg{
		XmsgIndex: xmsg.Index,
	}, true).Once()
}
