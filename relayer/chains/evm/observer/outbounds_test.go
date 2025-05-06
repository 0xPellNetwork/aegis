package observer_test

import (
	"testing"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/coin"
	"github.com/0xPellNetwork/aegis/relayer/config"
	"github.com/0xPellNetwork/aegis/relayer/testutils"
	"github.com/0xPellNetwork/aegis/relayer/testutils/stub"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

// getContractsByChainID is a helper func to get contracts and addresses by chainID
func getContractsByChainID(chainID int64) (*pellconnector.PellConnector, ethcommon.Address) {
	connector := stub.MockPellConnector(chainID)
	connectorAddress := testutils.StrategyManagerAddresses[chainID]
	return connector, connectorAddress
}

func Test_IsOutboundProcessed(t *testing.T) {
	// load archived outtx receipt that contains stakerdeposit event
	// https://etherscan.io/tx/0x81342051b8a85072d3e3771c1a57c7bdb5318e8caf37f5a687b7a91e50a7257f
	chain := chains.EthChain()
	chainID := chains.EthChain().Id
	nonce := uint64(9718)
	chainParam := stub.MockChainParams(chain.Id, 1)
	// outtxHash := "0x81342051b8a85072d3e3771c1a57c7bdb5318e8caf37f5a687b7a91e50a7257f"
	// xmsg := testutils.LoadXmsgByNonce(t, chainID, nonce)
	// receipt := testutils.LoadEVMOuttxReceipt(t, TestDataDir, chainID, outtxHash, coin.CoinType_Pell, testutils.EventPellSent)
	xmsg, outtx, receipt := testutils.LoadEVMXmsgNOuttxNReceipt(t, TestDataDir, chainID, nonce, testutils.EventPellSent)

	ctx := context.Background()

	t.Run("should post vote and return true if outtx is processed", func(t *testing.T) {
		// create evm client and set outtx and receipt
		client, _ := MockEVMClient(t, chain, config.EVMConfig{}, nil, nil, nil, nil, 1, chainParam)
		client.SetTxNReceipt(nonce, receipt, outtx)
		// post outbound vote
		isIncluded, isConfirmed, err := client.IsOutboundProcessed(ctx, xmsg, zerolog.Logger{})
		require.NoError(t, err)
		require.True(t, isIncluded)
		require.True(t, isConfirmed)
	})
	t.Run("should post vote and return true on restricted address", func(t *testing.T) {
		// load xmsg and modify sender address to arbitrary address
		// Note: other tests cases will fail if we use the original sender address because the
		// compliance config is globally set and will impact other tests when running in parallel
		xmsg := testutils.LoadXmsgByNonce(t, chainID, nonce)
		xmsg.InboundTxParams.Sender = sample.EthAddress().Hex()

		// create evm client and set outtx and receipt
		client, _ := MockEVMClient(t, chain, config.EVMConfig{}, nil, nil, nil, nil, 1, chainParam)
		client.SetTxNReceipt(nonce, receipt, outtx)

		// modify compliance config to restrict sender address
		cfg := config.Config{
			ComplianceConfig: config.ComplianceConfig{},
		}
		cfg.ComplianceConfig.RestrictedAddresses = []string{xmsg.InboundTxParams.Sender}
		config.LoadComplianceConfig(cfg)

		// post outbound vote
		isIncluded, isConfirmed, err := client.IsOutboundProcessed(ctx, xmsg, zerolog.Logger{})
		require.NoError(t, err)
		require.True(t, isIncluded)
		require.True(t, isConfirmed)
	})
	t.Run("should return false if outtx is not confirmed", func(t *testing.T) {
		// create evm client and DO NOT set outtx as confirmed
		client, _ := MockEVMClient(t, chain, config.EVMConfig{}, nil, nil, nil, nil, 1, chainParam)
		isIncluded, isConfirmed, err := client.IsOutboundProcessed(ctx, xmsg, zerolog.Logger{})
		require.NoError(t, err)
		require.False(t, isIncluded)
		require.False(t, isConfirmed)
	})
}

func Test_IsOutboundProcessed_ContractError(t *testing.T) {
	// Note: this test is skipped because it will cause CI failure.
	// The only way to replicate a contract error is to use an invalid ABI.
	// See the code: https://github.com/ethereum/go-ethereum/blob/v1.10.26/accounts/abi/bind/base.go#L97
	// The ABI is hardcoded in the protocol-contracts package and initialized the 1st time it binds the contract.
	// Any subsequent modification to the ABI will not work and therefor fail the unit test.
	t.Skip("uncomment this line to run this test separately, otherwise it will fail CI")

	// load archived outtx receipt that contains pellsent event
	// https://etherscan.io/tx/0x81342051b8a85072d3e3771c1a57c7bdb5318e8caf37f5a687b7a91e50a7257f
	chain := chains.EthChain()
	chainID := chains.EthChain().Id
	nonce := uint64(9718)
	chainParam := stub.MockChainParams(chain.Id, 1)
	outtxHash := "0x81342051b8a85072d3e3771c1a57c7bdb5318e8caf37f5a687b7a91e50a7257f"
	xmsg := testutils.LoadXmsgByNonce(t, chainID, nonce)
	receipt := testutils.LoadEVMOuttxReceipt(t, TestDataDir, chainID, outtxHash, coin.CoinType_PELL, testutils.EventPellSent)
	xmsg, outtx, receipt := testutils.LoadEVMXmsgNOuttxNReceipt(t, TestDataDir, chainID, nonce, testutils.EventPellSent)

	ctx := context.Background()

	t.Run("should fail if unable to get connector contract", func(t *testing.T) {
		// create evm client and set outtx and receipt
		client, _ := MockEVMClient(t, chain, config.EVMConfig{}, nil, nil, nil, nil, 1, chainParam)
		client.SetTxNReceipt(nonce, receipt, outtx)
		abiConnector := pellconnector.PellConnectorMetaData.ABI

		// set invalid connector ABI
		pellconnector.PellConnectorMetaData.ABI = "invalid abi"
		isIncluded, isConfirmed, err := client.IsOutboundProcessed(ctx, xmsg, zerolog.Logger{})
		pellconnector.PellConnectorMetaData.ABI = abiConnector // reset connector ABI
		require.ErrorContains(t, err, "error getting pell connector")
		require.False(t, isIncluded)
		require.False(t, isConfirmed)
	})
}

func Test_PostVoteOutbound(t *testing.T) {
	// Note: outtx of Gas/ERC20 token can also be used for this test
	// load archived xmsg, outtx and receipt for a pellsent event
	// https://etherscan.io/tx/0x81342051b8a85072d3e3771c1a57c7bdb5318e8caf37f5a687b7a91e50a7257f
	chain := chains.EthChain()
	nonce := uint64(9718)
	xmsg, outtx, receipt := testutils.LoadEVMXmsgNOuttxNReceipt(t, TestDataDir, chain.Id, nonce, testutils.EventPellSent)

	ctx := context.Background()

	t.Run("post vote outbound successfully", func(t *testing.T) {
		// the amount and status to be used for vote
		receiveStatus := chains.ReceiveStatus_SUCCESS

		// create evm client using mock pellBridge and post outbound vote
		pellBridge := stub.NewMockPellCoreBridge()
		client, _ := MockEVMClient(t, chain, config.EVMConfig{}, nil, nil, pellBridge, nil, 1, relayertypes.ChainParams{ChainId: chain.Id})
		client.PostVoteOutbound(ctx, xmsg.Index, receipt, outtx, receiveStatus, nonce, zerolog.Logger{})

		// pause the mock pellBridge to simulate error posting vote
		pellBridge.Stop()
		client.PostVoteOutbound(ctx, xmsg.Index, receipt, outtx, receiveStatus, nonce, zerolog.Logger{})
	})
}
