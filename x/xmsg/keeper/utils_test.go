// This file contains helper functions for testing the xmsg module
package keeper_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/dvsdirectory.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	testkeeper "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	pevmkeeper "github.com/0xPellNetwork/aegis/x/pevm/keeper"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var (
	// gasLimit = big.NewInt(21_000) - value used in SetupChainGasCoinAndPool for gas limit initialization
	gasPrice uint64 = 2
)

func createNLastBlockHeight(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.LastBlockHeight {
	items := make([]types.LastBlockHeight, n)
	for i := range items {
		items[i].Signer = "any"
		items[i].Index = fmt.Sprint(i)
		keeper.SetLastBlockHeight(ctx, items[i])
	}
	return items
}

// createXmsgWithNonceRange create in the store:
// mined xmsg from nonce 0 to low
// pending xmsg from low to high
// set pending nonces from low to higg
// return pending xmsgs
func createXmsgWithNonceRange(
	t *testing.T,
	ctx sdk.Context,
	k keeper.Keeper,
	lowPending int,
	highPending int,
	chainID int64,
	tss relayertypes.TSS,
	zk testkeeper.PellKeepers,
) (xmsgs []*types.Xmsg) {
	for i := 0; i < lowPending; i++ {
		xmsg := sample.Xmsg_pell(t, fmt.Sprintf("%d-%d", chainID, i))
		xmsg.XmsgStatus.Status = types.XmsgStatus_OUTBOUND_MINED
		xmsg.InboundTxParams.SenderChainId = chainID
		k.SetXmsg(ctx, *xmsg)
		zk.ObserverKeeper.SetNonceToXmsg(ctx, relayertypes.NonceToXmsg{
			ChainId:   chainID,
			Nonce:     int64(i),
			XmsgIndex: xmsg.Index,
			Tss:       tss.TssPubkey,
		})
	}
	for i := lowPending; i < highPending; i++ {
		xmsg := sample.Xmsg_pell(t, fmt.Sprintf("%d-%d", chainID, i))
		xmsg.XmsgStatus.Status = types.XmsgStatus_PENDING_OUTBOUND
		xmsg.InboundTxParams.SenderChainId = chainID
		k.SetXmsg(ctx, *xmsg)
		zk.ObserverKeeper.SetNonceToXmsg(ctx, relayertypes.NonceToXmsg{
			ChainId:   chainID,
			Nonce:     int64(i),
			XmsgIndex: xmsg.Index,
			Tss:       tss.TssPubkey,
		})
		xmsgs = append(xmsgs, xmsg)
	}
	zk.ObserverKeeper.SetPendingNonces(ctx, relayertypes.PendingNonces{
		ChainId:   chainID,
		NonceLow:  int64(lowPending),
		NonceHigh: int64(highPending),
		Tss:       tss.TssPubkey,
	})

	return
}

// getValidPellChainID get a valid pell chain id
func getValidPellChainID() int64 {
	return getValidPellChain().Id
}

func getValidPellChain() *chains.Chain {
	pell := chains.PellPrivnetChain()
	return &pell
}

// getValidEthChain get a valid eth chain
func getValidEthChain() *chains.Chain {
	goerli := chains.GoerliLocalnetChain()
	return &goerli
}

// getValidBscChainID get a valid eth chain id
func getValidBscChainID() int64 {
	return getValidBscChain().Id
}

// getValidBscChain get a valid eth chain
func getValidBscChain() *chains.Chain {
	bscTestNet := chains.BscTestnetChain()
	return &bscTestNet
}

// getValidEthChainID get a valid eth chain id
func getValidEthChainID() int64 {
	return getValidEthChain().Id
}

// getValidEthChainIDWithIndex get a valid eth chain id with index
func getValidEthChainIDWithIndex(t *testing.T, index int) int64 {
	switch index {
	case 0:
		return chains.GoerliLocalnetChain().Id
	case 1:
		return chains.GoerliChain().Id
	default:
		require.Fail(t, "invalid index")
	}
	return 0
}

// require that a contract has been deployed by checking stored code is non-empty.
func assertContractDeployment(t *testing.T, k *evmkeeper.Keeper, ctx sdk.Context, contractAddress common.Address) {
	acc := k.GetAccount(ctx, contractAddress)
	require.NotNil(t, acc)

	code := k.GetCode(ctx, common.BytesToHash(acc.CodeHash))
	require.NotEmpty(t, code)
}

// deploySystemContracts deploys the system contracts and returns their addresses.
func deploySystemContracts(
	t *testing.T,
	ctx sdk.Context,
	k *pevmkeeper.Keeper,
	evmk *evmkeeper.Keeper,
) (systemContract,
	connector,
	proxyAdmin,
	strategyManagerProxy,
	delegationManagerInteractorProxy,
	delegationManagerProxy,
	slasherProxy,
	dvsDirectoryProxy,
	registryRouter common.Address,
) {
	var err error

	systemContract, err = k.DeployPellSystemContract(ctx, types.ModuleAddressEVM)
	require.NoError(t, err)
	require.NotEmpty(t, systemContract)
	assertContractDeployment(t, evmk, ctx, systemContract)

	emptyContract, err := k.DeployPellEmptyContract(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, emptyContract)
	assertContractDeployment(t, evmk, ctx, emptyContract)

	proxyAdmin, err = k.DeployPellProxyAdmin(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, proxyAdmin)
	assertContractDeployment(t, evmk, ctx, proxyAdmin)

	connector, err = k.DeployPellConnector(ctx, systemContract, types.ModuleAddressEVM)
	require.NoError(t, err)
	require.NotEmpty(t, connector)
	assertContractDeployment(t, evmk, ctx, connector)

	strategyManagerProxy, err = k.DeployPellStrategyManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, strategyManagerProxy)
	assertContractDeployment(t, evmk, ctx, strategyManagerProxy)

	delegationManagerProxy, err = k.DeployPellDelegationManagerProxy(ctx, emptyContract, proxyAdmin, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, delegationManagerProxy)
	assertContractDeployment(t, evmk, ctx, delegationManagerProxy)

	slasherProxy, err = k.DeployPellSlasherProxy(ctx, emptyContract, proxyAdmin, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, slasherProxy)
	assertContractDeployment(t, evmk, ctx, slasherProxy)

	dvsDirectoryImpl, err := k.DeployPellDvsDirectory(ctx, delegationManagerProxy)
	require.NoError(t, err)
	require.NotEmpty(t, dvsDirectoryImpl)
	assertContractDeployment(t, evmk, ctx, dvsDirectoryImpl)

	dvsAbi, err := dvsdirectory.DVSDirectoryMetaData.GetAbi()
	require.NoError(t, err)
	require.NotEmpty(t, dvsAbi)

	data, err := dvsAbi.Pack("initialize", types.ModuleAddressEVM, []common.Address{types.ModuleAddressEVM}, types.ModuleAddressEVM, big.NewInt(0))
	require.NoError(t, err)
	require.NotEmpty(t, data)

	dvsDirectoryProxy, err = k.DeployPellDvsDirectoryProxy(ctx, dvsDirectoryImpl, proxyAdmin, data)
	require.NoError(t, err)
	require.NotEmpty(t, dvsDirectoryProxy)
	assertContractDeployment(t, evmk, ctx, dvsDirectoryProxy)

	registryRouter, err = k.DeployPellRegistryRouter(ctx, dvsDirectoryProxy, types.ModuleAddressEVM)
	require.NoError(t, err)
	require.NotEmpty(t, registryRouter)
	assertContractDeployment(t, evmk, ctx, registryRouter)

	return
}

// setSupportedChain sets the supported chains for the observer module
func setSupportedChain(ctx sdk.Context, zk testkeeper.PellKeepers, chainIDs ...int64) {
	chainParamsList := make([]*relayertypes.ChainParams, len(chainIDs))
	for i, chainID := range chainIDs {
		chainParams := sample.ChainParams_pell(chainID)
		chainParams.IsSupported = true
		chainParamsList[i] = chainParams
	}
	zk.ObserverKeeper.SetChainParamsList(ctx, relayertypes.ChainParamsList{
		ChainParams: chainParamsList,
	})
}

// buildXmsg returns a sample Xmsg with Pell params. This is used for testing Inbound and Outbound voting transactions
func buildXmsg(t *testing.T, receiver common.Address, senderChain chains.Chain) *types.Xmsg {
	r := sample.Rand()
	xmsg := &types.Xmsg{
		Signer:           sample.AccAddress(),
		Index:            sample.PellIndex(t),
		XmsgStatus:       &types.Status{Status: types.XmsgStatus_PENDING_INBOUND},
		InboundTxParams:  sample.InboundTxParams_pell(r),
		OutboundTxParams: []*types.OutboundTxParams{sample.OutboundTxParams_pell(r)},
	}

	xmsg.GetInboundTxParams().SenderChainId = senderChain.Id
	xmsg.GetInboundTxParams().InboundTxHash = sample.Hash().String()
	xmsg.GetInboundTxParams().InboundTxBallotIndex = sample.PellIndex(t)

	xmsg.GetCurrentOutTxParam().ReceiverChainId = senderChain.Id
	xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
	xmsg.GetCurrentOutTxParam().OutboundTxHash = sample.Hash().String()
	xmsg.GetCurrentOutTxParam().OutboundTxBallotIndex = sample.PellIndex(t)

	xmsg.GetInboundTxParams().Sender = sample.EthAddress().String()
	xmsg.GetCurrentOutTxParam().OutboundTxTssNonce = 42
	xmsg.GetCurrentOutTxParam().OutboundTxGasUsed = 100
	xmsg.GetCurrentOutTxParam().OutboundTxEffectiveGasLimit = 100
	return xmsg
}

func createNInTxHashToXmsg(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.InTxHashToXmsg {
	items := make([]types.InTxHashToXmsg, n)
	for i := range items {
		items[i].InTxHash = strconv.Itoa(i)

		keeper.SetInTxHashToXmsg(ctx, items[i])
	}
	return items
}

func createNGasPrice(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.GasPrice {
	items := make([]types.GasPrice, n)
	for i := range items {
		items[i].Signer = "any"
		items[i].ChainId = int64(i)
		items[i].Index = strconv.FormatInt(int64(i), 10)
		keeper.SetGasPrice(ctx, items[i])
	}
	return items
}
