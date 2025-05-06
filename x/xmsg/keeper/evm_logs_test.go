package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgkeeper "github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// SetupStateForProcessLogsPellSent sets up additional state required for processing logs for PellSent events
// This sets up the gas coin, zrc20 contract, gas price, zrc20 pool.
// This should be used in conjunction with SetupStateForProcessLogs for processing PellSent events
func SetupStateForProcessLogsPellSent(
	t *testing.T,
	ctx sdk.Context,
	k *xmsgkeeper.Keeper,
	zk keepertest.PellKeepers,
	sdkk keepertest.SDKKeepers,
	chain chains.Chain,
	admin string,
) {
	k.SetGasPrice(ctx, xmsgtypes.GasPrice{
		ChainId:     chain.Id,
		MedianIndex: 0,
		Prices:      []uint64{gasPrice},
	})
}

// SetupStateForProcessLogs sets up observer state for required for processing logs
// It deploys system contracts, sets up TSS, gas price, chain nonce's, pending nonce's.These are all required to create a xmsg from a log
func SetupStateForProcessLogs(
	t *testing.T,
	ctx sdk.Context,
	k *xmsgkeeper.Keeper,
	zk keepertest.PellKeepers,
	sdkk keepertest.SDKKeepers,
	chain chains.Chain,
) {

	deploySystemContracts(t, ctx, zk.PevmKeeper, sdkk.EvmKeeper)
	tss := sample.Tss_pell()
	zk.ObserverKeeper.SetTSS(ctx, tss)
	k.SetGasPrice(ctx, xmsgtypes.GasPrice{
		ChainId: chain.Id,
		Prices:  []uint64{100},
	})

	zk.ObserverKeeper.SetChainNonces(ctx, relayertypes.ChainNonces{
		Index:   chain.ChainName(),
		ChainId: chain.Id,
		Nonce:   0,
	})
	zk.ObserverKeeper.SetPendingNonces(ctx, relayertypes.PendingNonces{
		NonceLow:  0,
		NonceHigh: 0,
		ChainId:   chain.Id,
		Tss:       tss.TssPubkey,
	})
}

func TestKeeper_ParsePellSentEvent(t *testing.T) {
	// t.Run("successfully parse a valid event", func(t *testing.T) {
	// 	logs := sample.GetValidPellSent_pell(t).Logs
	// 	for i, log := range logs {
	// 		connector := log.Address
	// 		event, err := xmsgkeeper.ParsePRC20PellSentEvent(*log, connector)
	// 		if i < 4 {
	// 			require.ErrorContains(t, err, "event signature mismatch")
	// 			require.Nil(t, event)
	// 			continue
	// 		}
	// 		require.Equal(t, chains.EthChain().ChainId, event.DestinationChainId.Int64())
	// 		require.Equal(t, "0x60983881bdf302dcfa96603A58274D15D5966209", event.SourceTxOriginAddress.String())
	// 		require.Equal(t, "0xF0a3F93Ed1B126142E61423F9546bf1323Ff82DF", event.PellTxSenderAddress.String())
	// 	}
	// })

	// t.Run("unable to parse if topics field is empty", func(t *testing.T) {
	// 	logs := sample.GetValidPellSent_pell(t).Logs
	// 	for _, log := range logs {
	// 		connector := log.Address
	// 		log.Topics = nil
	// 		event, err := xmsgkeeper.ParsePRC20PellSentEvent(*log, connector)
	// 		require.ErrorContains(t, err, "ParsePRC20PellSentEvent: invalid log - no topics")
	// 		require.Nil(t, event)
	// 	}
	// })

	// t.Run("unable to parse if connector address does not match", func(t *testing.T) {
	// 	logs := sample.GetValidPellSent_pell(t).Logs
	// 	for i, log := range logs {
	// 		event, err := xmsgkeeper.ParsePRC20PellSentEvent(*log, sample.EthAddress())
	// 		if i < 4 {
	// 			require.ErrorContains(t, err, "event signature mismatch")
	// 			require.Nil(t, event)
	// 			continue
	// 		}
	// 		require.ErrorContains(t, err, "does not match pellConnector")
	// 		require.Nil(t, event)
	// 	}
	// })
}

func TestKeeper_ProcessPellSentEvent(t *testing.T) {
	// t.Run("successfully process PellSentEvent", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)

	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	chainID := chain.ChainId
	// 	setSupportedChain(ctx, zk, chainID)

	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)
	// 	admin := keepertest.SetAdminPolices(ctx, zk.AuthorityKeeper)
	// 	SetupStateForProcessLogsPellSent(t, ctx, k, zk, sdkk, chain, admin)

	// 	amount, ok := sdkmath.NewIntFromString("20000000000000000000000")
	// 	require.True(t, ok)
	// 	err := sdkk.BankKeeper.MintCoins(ctx, pevmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	// 	require.NoError(t, err)

	// 	event, err := xmsgkeeper.ParsePRC20PellSentEvent(*sample.GetValidPellSent_pell(t).Logs[4], sample.GetValidPellSent_pell(t).Logs[4].Address)
	// 	require.NoError(t, err)
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()
	// 	tss := sample.Tss()

	// 	err = k.ProcessPellSentEvent(ctx, event, emittingContract, txOrigin.Hex(), tss)
	// 	require.NoError(t, err)
	// 	xmsgList := k.GetAllXmsg(ctx)
	// 	require.Len(t, xmsgList, 1)
	// 	require.Equal(t, strings.Compare("0x60983881bdf302dcfa96603a58274d15d5966209", xmsgList[0].GetCurrentOutTxParam().Receiver), 0)
	// 	require.Equal(t, chains.EthChain().ChainId, xmsgList[0].GetCurrentOutTxParam().ReceiverChainId)
	// 	require.Equal(t, emittingContract.Hex(), xmsgList[0].InboundTxParams.Sender)
	// 	require.Equal(t, txOrigin.Hex(), xmsgList[0].InboundTxParams.TxOrigin)
	// })

	// t.Run("unable to process PellSentEvent if pevm module does not have enough balance", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	chainID := chain.ChainId
	// 	setSupportedChain(ctx, zk, chainID)
	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)
	// 	admin := keepertest.SetAdminPolices(ctx, zk.AuthorityKeeper)
	// 	SetupStateForProcessLogsPellSent(t, ctx, k, zk, sdkk, chain, admin)

	// 	event, err := xmsgkeeper.ParsePRC20PellSentEvent(*sample.GetValidPellSent_pell(t).Logs[4], sample.GetValidPellSent_pell(t).Logs[4].Address)
	// 	require.NoError(t, err)
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()
	// 	tss := sample.Tss()

	// 	err = k.ProcessPellSentEvent(ctx, event, emittingContract, txOrigin.Hex(), tss)
	// 	require.ErrorContains(t, err, "ProcessPellSentEvent: failed to burn coins from pevm")
	// })

	// t.Run("unable to process PellSentEvent if receiver chain is not supported", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)
	// 	admin := keepertest.SetAdminPolices(ctx, zk.AuthorityKeeper)
	// 	SetupStateForProcessLogsPellSent(t, ctx, k, zk, sdkk, chain, admin)

	// 	amount, ok := sdkmath.NewIntFromString("20000000000000000000000")
	// 	require.True(t, ok)
	// 	err := sdkk.BankKeeper.MintCoins(ctx, pevmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	// 	require.NoError(t, err)

	// 	event, err := xmsgkeeper.ParsePRC20PellSentEvent(*sample.GetValidPellSent_pell(t).Logs[4], sample.GetValidPellSent_pell(t).Logs[4].Address)
	// 	require.NoError(t, err)
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()
	// 	tss := sample.Tss()
	// 	err = k.ProcessPellSentEvent(ctx, event, emittingContract, txOrigin.Hex(), tss)
	// 	require.ErrorContains(t, err, "chain not supported")
	// })

	// t.Run("unable to process PellSentEvent if pellchain chain id not correctly set in context", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	chainID := chain.ChainId
	// 	setSupportedChain(ctx, zk, chainID)
	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)
	// 	admin := keepertest.SetAdminPolices(ctx, zk.AuthorityKeeper)
	// 	SetupStateForProcessLogsPellSent(t, ctx, k, zk, sdkk, chain, admin)

	// 	amount, ok := sdkmath.NewIntFromString("20000000000000000000000")
	// 	require.True(t, ok)
	// 	err := sdkk.BankKeeper.MintCoins(ctx, pevmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	// 	require.NoError(t, err)

	// 	event, err := xmsgkeeper.ParsePRC20PellSentEvent(*sample.GetValidPellSent_pell(t).Logs[4], sample.GetValidPellSent_pell(t).Logs[4].Address)
	// 	require.NoError(t, err)
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()
	// 	tss := sample.Tss()
	// 	ctx = ctx.WithChainID("test-21-1")
	// 	err = k.ProcessPellSentEvent(ctx, event, emittingContract, txOrigin.Hex(), tss)
	// 	require.ErrorContains(t, err, "ProcessPellSentEvent: failed to convert chainID")
	// })

	// t.Run("unable to process PellSentEvent if gas pay fails", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	chainID := chain.ChainId
	// 	setSupportedChain(ctx, zk, chainID)
	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)

	// 	amount, ok := sdkmath.NewIntFromString("20000000000000000000000")
	// 	require.True(t, ok)
	// 	err := sdkk.BankKeeper.MintCoins(ctx, pevmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	// 	require.NoError(t, err)
	// 	event, err := xmsgkeeper.ParsePRC20PellSentEvent(*sample.GetValidPellSent_pell(t).Logs[4], sample.GetValidPellSent_pell(t).Logs[4].Address)
	// 	require.NoError(t, err)
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()
	// 	tss := sample.Tss()

	// 	err = k.ProcessPellSentEvent(ctx, event, emittingContract, txOrigin.Hex(), tss)
	// 	require.ErrorContains(t, err, "ProcessWithdrawalEvent: pay gas failed")
	// })

	// t.Run("unable to process PellSentEvent if process xmsg fails", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	chainID := chain.ChainId
	// 	setSupportedChain(ctx, zk, chainID)

	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)
	// 	admin := keepertest.SetAdminPolices(ctx, zk.AuthorityKeeper)
	// 	SetupStateForProcessLogsPellSent(t, ctx, k, zk, sdkk, chain, admin)

	// 	zk.ObserverKeeper.SetChainNonces(ctx, relayertypes.ChainNonces{
	// 		Index:   chain.ChainName(),
	// 		ChainId: chain.ChainId,
	// 		Nonce:   1,
	// 	})
	// 	amount, ok := sdkmath.NewIntFromString("20000000000000000000000")
	// 	require.True(t, ok)
	// 	err := sdkk.BankKeeper.MintCoins(ctx, pevmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	// 	require.NoError(t, err)

	// 	event, err := xmsgkeeper.ParsePRC20PellSentEvent(*sample.GetValidPellSent_pell(t).Logs[4], sample.GetValidPellSent_pell(t).Logs[4].Address)
	// 	require.NoError(t, err)
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()
	// 	tss := sample.Tss()
	// 	err = k.ProcessPellSentEvent(ctx, event, emittingContract, txOrigin.Hex(), tss)
	// 	require.ErrorContains(t, err, "ProcessWithdrawalEvent: update nonce failed")
	// })
}

func TestKeeper_ProcessLogs(t *testing.T) {

	// t.Run("successfully parse and process PellSentEvent", func(t *testing.T) {
	// 	k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	chain := chains.EthChain()
	// 	chainID := chain.ChainId
	// 	setSupportedChain(ctx, zk, chainID)
	// 	SetupStateForProcessLogs(t, ctx, k, zk, sdkk, chain)
	// 	admin := keepertest.SetAdminPolices(ctx, zk.AuthorityKeeper)
	// 	SetupStateForProcessLogsPellSent(t, ctx, k, zk, sdkk, chain, admin)

	// 	amount, ok := sdkmath.NewIntFromString("20000000000000000000000")
	// 	require.True(t, ok)
	// 	err := sdkk.BankKeeper.MintCoins(ctx, pevmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, amount)))
	// 	require.NoError(t, err)
	// 	block := sample.GetValidPellSent_pell(t)
	// 	system, found := zk.PevmKeeper.GetSystemContract(ctx)
	// 	require.True(t, found)
	// 	for _, log := range block.Logs {
	// 		log.Address = ethcommon.HexToAddress(system.Connector)
	// 	}
	// 	emittingContract := sample.EthAddress()
	// 	txOrigin := sample.EthAddress()

	// 	err = k.ProcessLogs(ctx, block.Logs, emittingContract, txOrigin.Hex())
	// 	require.NoError(t, err)
	// 	xmsgList := k.GetAllXmsg(ctx)
	// 	require.Len(t, xmsgList, 1)
	// 	require.Equal(t, strings.Compare("0x60983881bdf302dcfa96603a58274d15d5966209", xmsgList[0].GetCurrentOutTxParam().Receiver), 0)
	// 	require.Equal(t, chains.EthChain().ChainId, xmsgList[0].GetCurrentOutTxParam().ReceiverChainId)
	// 	require.Equal(t, emittingContract.Hex(), xmsgList[0].InboundTxParams.Sender)
	// 	require.Equal(t, txOrigin.Hex(), xmsgList[0].InboundTxParams.TxOrigin)
	// })

	// t.Run("unable to process logs if system contract not found", func(t *testing.T) {
	// 	k, ctx, _, _ := keepertest.XmsgKeeper(t)
	// 	k.GetAuthKeeper().GetModuleAccount(ctx, pevmtypes.ModuleName)

	// 	err := k.ProcessLogs(ctx, sample.GetValidPellSent_pell(t).Logs, sample.EthAddress(), "")
	// 	require.ErrorContains(t, err, "cannot find system contract")
	// 	xmsgList := k.GetAllXmsg(ctx)
	// 	require.Len(t, xmsgList, 0)
	// })
}
