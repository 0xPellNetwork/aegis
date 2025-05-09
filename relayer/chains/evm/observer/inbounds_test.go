package observer_test

// func TestEVM_CheckAndVoteInboundTokenPell(t *testing.T) {
// 	// load archived PellSent intx, receipt and xmsg
// 	// https://etherscan.io/tx/0xf3935200c80f98502d5edc7e871ffc40ca898e134525c42c2ae3cbc5725f9d76
// 	chain := chains.EthChain()
// 	confirmation := uint64(10)
// 	chainID := chain.ChainId
// 	chainParam := stub.MockChainParams(chain.ChainId, confirmation)
// 	intxHash := "0xf3935200c80f98502d5edc7e871ffc40ca898e134525c42c2ae3cbc5725f9d76"

// 	t.Run("should pass for archived intx, receipt and xmsg", func(t *testing.T) {
// 		tx, receipt, xmsg := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Pell)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenPell(tx, receipt, false)
// 		require.NoError(t, err)
// 		require.Equal(t, xmsg.InboundTxParams.InboundTxBallotIndex, ballot)
// 	})
// 	t.Run("should fail on unconfirmed intx", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Pell)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation - 1

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		_, err := ob.CheckAndVoteInboundTokenPell(tx, receipt, false)
// 		require.ErrorContains(t, err, "not been confirmed")
// 	})
// 	t.Run("should not act if no PellSent event", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Pell)
// 		receipt.Logs = receipt.Logs[:2] // remove PellSent event
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenPell(tx, receipt, true)
// 		require.NoError(t, err)
// 		require.Equal(t, "", ballot)
// 	})
// 	t.Run("should not act if emitter is not PellConnector", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Pell)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		chainID = 56 // use BSC chain connector
// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, stub.MockChainParams(chainID, confirmation))
// 		_, err := ob.CheckAndVoteInboundTokenPell(tx, receipt, true)
// 		require.ErrorContains(t, err, "emitter address mismatch")
// 	})
// }

// func TestEVM_CheckAndVoteInboundTokenERC20(t *testing.T) {
// 	// load archived ERC20 intx, receipt and xmsg
// 	// https://etherscan.io/tx/0x4ea69a0e2ff36f7548ab75791c3b990e076e2a4bffeb616035b239b7d33843da
// 	chain := chains.EthChain()
// 	confirmation := uint64(10)
// 	chainID := chain.ChainId
// 	chainParam := stub.MockChainParams(chain.ChainId, confirmation)
// 	intxHash := "0x4ea69a0e2ff36f7548ab75791c3b990e076e2a4bffeb616035b239b7d33843da"

// 	t.Run("should pass for archived intx, receipt and xmsg", func(t *testing.T) {
// 		tx, receipt, xmsg := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_ERC20)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenERC20(tx, receipt, false)
// 		require.NoError(t, err)
// 		require.Equal(t, xmsg.InboundTxParams.InboundTxBallotIndex, ballot)
// 	})
// 	t.Run("should fail on unconfirmed intx", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_ERC20)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation - 1

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		_, err := ob.CheckAndVoteInboundTokenERC20(tx, receipt, false)
// 		require.ErrorContains(t, err, "not been confirmed")
// 	})
// 	t.Run("should not act if no Deposit event", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_ERC20)
// 		receipt.Logs = receipt.Logs[:1] // remove Deposit event
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenERC20(tx, receipt, true)
// 		require.NoError(t, err)
// 		require.Equal(t, "", ballot)
// 	})
// 	t.Run("should not act if emitter is not ERC20 Custody", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_ERC20)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		chainID = 56 // use BSC chain ERC20 custody
// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, stub.MockChainParams(chainID, confirmation))
// 		_, err := ob.CheckAndVoteInboundTokenERC20(tx, receipt, true)
// 		require.ErrorContains(t, err, "emitter address mismatch")
// 	})
// }

// func TestEVM_CheckAndVoteInboundTokenGas(t *testing.T) {
// 	// load archived Gas intx, receipt and xmsg
// 	// https://etherscan.io/tx/0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532
// 	chain := chains.EthChain()
// 	confirmation := uint64(10)
// 	chainID := chain.ChainId
// 	chainParam := stub.MockChainParams(chain.ChainId, confirmation)
// 	intxHash := "0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532"

// 	t.Run("should pass for archived intx, receipt and xmsg", func(t *testing.T) {
// 		tx, receipt, xmsg := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Gas)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenGas(tx, receipt, false)
// 		require.NoError(t, err)
// 		require.Equal(t, xmsg.InboundTxParams.InboundTxBallotIndex, ballot)
// 	})
// 	t.Run("should fail on unconfirmed intx", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Gas)
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation - 1

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		_, err := ob.CheckAndVoteInboundTokenGas(tx, receipt, false)
// 		require.ErrorContains(t, err, "not been confirmed")
// 	})
// 	t.Run("should not act if receiver is not TSS", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Gas)
// 		tx.To = testutils.OtherAddress1 // use other address
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenGas(tx, receipt, false)
// 		require.ErrorContains(t, err, "not TSS address")
// 		require.Equal(t, "", ballot)
// 	})
// 	t.Run("should not act if transaction failed", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Gas)
// 		receipt.Status = ethtypes.ReceiptStatusFailed
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenGas(tx, receipt, false)
// 		require.ErrorContains(t, err, "not a successful tx")
// 		require.Equal(t, "", ballot)
// 	})
// 	t.Run("should not act on nil message", func(t *testing.T) {
// 		tx, receipt, _ := testutils.LoadEVMIntxNReceiptNXmsg(t, chainID, intxHash, coin.CoinType_Gas)
// 		tx.Input = hex.EncodeToString([]byte(constant.DonationMessage)) // donation will result in nil message
// 		require.NoError(t, evm.ValidateEvmTransaction(tx))
// 		lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 		ob := MockEVMClient(t, chain, nil, nil, nil, nil, lastBlock, chainParam)
// 		ballot, err := ob.CheckAndVoteInboundTokenGas(tx, receipt, false)
// 		require.NoError(t, err)
// 		require.Equal(t, "", ballot)
// 	})
// }

// func TestEVM_BuildInboundVoteMsgForPellSentEvent(t *testing.T) {
// 	// load archived PellSent receipt
// 	// https://etherscan.io/tx/0xf3935200c80f98502d5edc7e871ffc40ca898e134525c42c2ae3cbc5725f9d76
// 	chainID := int64(1)
// 	chain := chains.EthChain()
// 	intxHash := "0xf3935200c80f98502d5edc7e871ffc40ca898e134525c42c2ae3cbc5725f9d76"
// 	receipt := testutils.LoadEVMIntxReceipt(t, TestDataDir, chainID, intxHash, coin.CoinType_Pell)
// 	xmsg := testutils.LoadXmsgByIntx(t, chainID, coin.CoinType_Pell, intxHash)

// 	// parse PellSent event
// 	ob := MockEVMClient(t, chain, nil, nil, nil, nil, 1, stub.MockChainParams(1, 1))
// 	connector := stub.MockStrategyManager(chainID)
// 	event := testutils.ParseReceiptPellSent(receipt, connector)

// 	// create test compliance config
// 	cfg := config.Config{
// 		ComplianceConfig: config.ComplianceConfig{},
// 	}

// 	t.Run("should return vote msg for archived PellSent event", func(t *testing.T) {
// 		msg := ob.BuildInboundVoteMsgForPellSentEvent(event)
// 		require.NotNil(t, msg)
// 		require.Equal(t, xmsg.InboundTxParams.InboundTxBallotIndex, msg.Digest())
// 	})
// 	t.Run("should return nil msg if sender is restricted", func(t *testing.T) {
// 		sender := event.PellTxSenderAddress.Hex()
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{sender}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForPellSentEvent(event)
// 		require.Nil(t, msg)
// 	})
// 	t.Run("should return nil msg if receiver is restricted", func(t *testing.T) {
// 		receiver := clienttypes.BytesToEthHex(event.DestinationAddress)
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{receiver}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForPellSentEvent(event)
// 		require.Nil(t, msg)
// 	})
// 	t.Run("should return nil msg if txOrigin is restricted", func(t *testing.T) {
// 		txOrigin := event.SourceTxOriginAddress.Hex()
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{txOrigin}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForPellSentEvent(event)
// 		require.Nil(t, msg)
// 	})
// }

// func TestEVM_BuildInboundVoteMsgForDepositedEvent(t *testing.T) {
// 	// load archived Deposited receipt
// 	// https://etherscan.io/tx/0x4ea69a0e2ff36f7548ab75791c3b990e076e2a4bffeb616035b239b7d33843da
// 	chain := chains.EthChain()
// 	chainID := chain.ChainId
// 	intxHash := "0x4ea69a0e2ff36f7548ab75791c3b990e076e2a4bffeb616035b239b7d33843da"
// 	tx, receipt := testutils.LoadEVMIntxNReceipt(t, chainID, intxHash, coin.CoinType_ERC20)
// 	xmsg := testutils.LoadXmsgByIntx(t, chainID, coin.CoinType_ERC20, intxHash)

// 	// parse Deposited event
// 	ob := MockEVMClient(t, chain, nil, nil, nil, nil, 1, stub.MockChainParams(1, 1))
// 	custody := stub.MockDelegationManager(chainID)
// 	event := testutils.ParseReceiptERC20Deposited(receipt, custody)
// 	sender := ethcommon.HexToAddress(tx.From)

// 	// create test compliance config
// 	cfg := config.Config{
// 		ComplianceConfig: config.ComplianceConfig{},
// 	}

// 	t.Run("should return vote msg for archived Deposited event", func(t *testing.T) {
// 		msg := ob.BuildInboundVoteMsgForDepositedEvent(event, sender)
// 		require.NotNil(t, msg)
// 		require.Equal(t, xmsg.InboundTxParams.InboundTxBallotIndex, msg.Digest())
// 	})
// 	t.Run("should return nil msg if sender is restricted", func(t *testing.T) {
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{sender.Hex()}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForDepositedEvent(event, sender)
// 		require.Nil(t, msg)
// 	})
// 	t.Run("should return nil msg if receiver is restricted", func(t *testing.T) {
// 		receiver := clienttypes.BytesToEthHex(event.Recipient)
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{receiver}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForDepositedEvent(event, sender)
// 		require.Nil(t, msg)
// 	})
// 	t.Run("should return nil msg on donation transaction", func(t *testing.T) {
// 		event.Message = []byte(constant.DonationMessage)
// 		msg := ob.BuildInboundVoteMsgForDepositedEvent(event, sender)
// 		require.Nil(t, msg)
// 	})
// }

// func TestEVM_BuildInboundVoteMsgForTokenSentToTSS(t *testing.T) {
// 	// load archived gas token transfer to TSS
// 	// https://etherscan.io/tx/0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532
// 	chain := chains.EthChain()
// 	chainID := chain.ChainId
// 	intxHash := "0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532"
// 	tx, receipt := testutils.LoadEVMIntxNReceipt(t, chainID, intxHash, coin.CoinType_Gas)
// 	require.NoError(t, evm.ValidateEvmTransaction(tx))
// 	xmsg := testutils.LoadXmsgByIntx(t, chainID, coin.CoinType_Gas, intxHash)

// 	// load archived gas token donation to TSS
// 	// https://etherscan.io/tx/0x52f214cf7b10be71f4d274193287d47bc9632b976e69b9d2cdeb527c2ba32155
// 	inTxHashDonation := "0x52f214cf7b10be71f4d274193287d47bc9632b976e69b9d2cdeb527c2ba32155"
// 	txDonation, receiptDonation := testutils.LoadEVMIntxNReceiptDonation(t, chainID, inTxHashDonation, coin.CoinType_Gas)
// 	require.NoError(t, evm.ValidateEvmTransaction(txDonation))

// 	// create test compliance config
// 	ob := MockEVMClient(t, chain, nil, nil, nil, nil, 1, stub.MockChainParams(1, 1))
// 	cfg := config.Config{
// 		ComplianceConfig: config.ComplianceConfig{},
// 	}

// 	t.Run("should return vote msg for archived gas token transfer to TSS", func(t *testing.T) {
// 		msg := ob.BuildInboundVoteMsgForTokenSentToTSS(tx, ethcommon.HexToAddress(tx.From), receipt.BlockNumber.Uint64())
// 		require.NotNil(t, msg)
// 		require.Equal(t, xmsg.InboundTxParams.InboundTxBallotIndex, msg.Digest())
// 	})
// 	t.Run("should return nil msg if sender is restricted", func(t *testing.T) {
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{tx.From}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForTokenSentToTSS(tx, ethcommon.HexToAddress(tx.From), receipt.BlockNumber.Uint64())
// 		require.Nil(t, msg)
// 	})
// 	t.Run("should return nil msg if receiver is restricted", func(t *testing.T) {
// 		txCopy := &ethrpc.Transaction{}
// 		*txCopy = *tx
// 		message := hex.EncodeToString(ethcommon.HexToAddress(testutils.OtherAddress1).Bytes())
// 		txCopy.Input = message // use other address as receiver
// 		cfg.ComplianceConfig.RestrictedAddresses = []string{testutils.OtherAddress1}
// 		config.LoadComplianceConfig(cfg)
// 		msg := ob.BuildInboundVoteMsgForTokenSentToTSS(txCopy, ethcommon.HexToAddress(txCopy.From), receipt.BlockNumber.Uint64())
// 		require.Nil(t, msg)
// 	})
// 	t.Run("should return nil msg on donation transaction", func(t *testing.T) {
// 		msg := ob.BuildInboundVoteMsgForTokenSentToTSS(txDonation,
// 			ethcommon.HexToAddress(txDonation.From), receiptDonation.BlockNumber.Uint64())
// 		require.Nil(t, msg)
// 	})
// }

// func TestEVM_ObserveTSSReceiveInBlock(t *testing.T) {
// 	// https://etherscan.io/tx/0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532
// 	chain := chains.EthChain()
// 	chainID := chain.ChainId
// 	confirmation := uint64(1)
// 	chainParam := stub.MockChainParams(chain.ChainId, confirmation)
// 	intxHash := "0xeaec67d5dd5d85f27b21bef83e01cbdf59154fd793ea7a22c297f7c3a722c532"

// 	// load archived tx and receipt
// 	tx, receipt := testutils.LoadEVMIntxNReceipt(t, chainID, intxHash, coin.CoinType_Gas)
// 	require.NoError(t, evm.ValidateEvmTransaction(tx))

// 	// load archived evm block
// 	// https://etherscan.io/block/19363323
// 	blockNumber := receipt.BlockNumber.Uint64()
// 	block := testutils.LoadEVMBlock(t, TestDataDir, chainID, blockNumber, true)

// 	// create mock client
// 	evmClient := stub.NewMockEvmClient()
// 	evmJSONRPC := stub.NewMockJSONRPCClient()
// 	pellBridge := stub.NewMockPellCoreBridge()
// 	tss := stub.NewTSSMainnet()
// 	lastBlock := receipt.BlockNumber.Uint64() + confirmation

// 	t.Run("should observe TSS receive in block", func(t *testing.T) {
// 		ob := MockEVMClient(t, chain, evmClient, evmJSONRPC, pellBridge, tss, lastBlock, chainParam)

// 		// feed archived block and receipt
// 		evmJSONRPC.WithBlock(block)
// 		evmClient.WithReceipt(receipt)
// 		err := ob.ObserveTSSReceiveInBlock(blockNumber)
// 		require.NoError(t, err)
// 	})
// 	t.Run("should not observe on error getting block", func(t *testing.T) {
// 		ob := MockEVMClient(t, chain, evmClient, evmJSONRPC, pellBridge, tss, lastBlock, chainParam)
// 		err := ob.ObserveTSSReceiveInBlock(blockNumber)
// 		// error getting block is expected because the mock JSONRPC contains no block
// 		require.ErrorContains(t, err, "error getting block")
// 	})
// 	t.Run("should not observe on error getting receipt", func(t *testing.T) {
// 		ob := MockEVMClient(t, chain, evmClient, evmJSONRPC, pellBridge, tss, lastBlock, chainParam)
// 		evmJSONRPC.WithBlock(block)
// 		err := ob.ObserveTSSReceiveInBlock(blockNumber)
// 		// error getting block is expected because the mock evmClient contains no receipt
// 		require.ErrorContains(t, err, "error getting receipt")
// 	})
// 	t.Run("should not observe on error posting vote", func(t *testing.T) {
// 		ob := MockEVMClient(t, chain, evmClient, evmJSONRPC, pellBridge, tss, lastBlock, chainParam)

// 		// feed archived block and pause pell bridge
// 		evmJSONRPC.WithBlock(block)
// 		evmClient.WithReceipt(receipt)
// 		pellBridge.Pause()
// 		err := ob.ObserveTSSReceiveInBlock(blockNumber)
// 		// error posting vote is expected because the mock pellClient is paused
// 		require.ErrorContains(t, err, "error checking and voting")
// 	})
// }
