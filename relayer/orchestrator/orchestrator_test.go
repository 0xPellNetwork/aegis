package orchestrator

import (
	"testing"

	cosmossdk_io_math "cosmossdk.io/math"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/config"
	corecontext "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/testutils"
	"github.com/0xPellNetwork/aegis/relayer/testutils/stub"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func MockCoreObserver(
	t *testing.T,
	evmChain chains.Chain,
	evmChainParams *observertypes.ChainParams,
) *Orchestrator {
	// create mock signers and clients
	evmSigner := stub.NewEVMSigner(
		evmChain,
		ethcommon.HexToAddress(evmChainParams.DelegationManagerContractAddress),
	)
	evmClient := stub.NewEVMClient(evmChainParams)

	// create core observer
	observer := &Orchestrator{
		signerMap: map[int64]interfaces.ChainSigner{
			evmChain.Id: evmSigner,
		},
		observerMap: map[int64]interfaces.ChainClient{
			evmChain.Id: evmClient,
		},
	}
	return observer
}

func CreateCoreContext(
	evmChain chains.Chain,
	evmChainParams *observertypes.ChainParams,
) *corecontext.PellCoreContext {
	// new config
	cfg := config.NewConfig()
	cfg.EVMChainConfigs[evmChain.Id] = config.EVMConfig{
		Chain: evmChain,
	}
	cfg.BitcoinConfig = config.BTCConfig{
		RPCHost: "localhost",
	}
	// new core context
	coreContext := corecontext.NewPellCoreContext(cfg)
	evmChainParamsMap := make(map[int64]*observertypes.ChainParams)
	evmChainParamsMap[evmChain.Id] = evmChainParams
	ccFlags := sample.CrosschainFlags_pell()
	verificationFlags := sample.VerificationFlags()

	// feed chain params
	coreContext.Update(
		observertypes.Keygen{},
		[]chains.Chain{evmChain},
		evmChainParamsMap,
		"",
		*ccFlags,
		verificationFlags,
		true,
		zerolog.Logger{},
	)
	return coreContext
}

func Test_GetUpdatedSigner(t *testing.T) {
	// initial parameters for core observer creation
	evmChain := chains.EthChain()
	evmChainParams := &observertypes.ChainParams{
		ChainId:                          evmChain.Id,
		ConnectorContractAddress:         testutils.StrategyManagerAddresses[evmChain.Id].Hex(),
		DelegationManagerContractAddress: testutils.DelegationManagerAddresses[evmChain.Id].Hex(),
	}

	// new chain params in core context
	evmChainParamsNew := &observertypes.ChainParams{
		ChainId:                          evmChain.Id,
		ConnectorContractAddress:         testutils.OtherAddress1,
		DelegationManagerContractAddress: testutils.OtherAddress2,
	}

	t.Run("signer should not be found", func(t *testing.T) {
		observer := MockCoreObserver(t, evmChain, evmChainParams)
		coreContext := CreateCoreContext(evmChain, evmChainParamsNew)
		// BSC signer should not be found
		_, err := observer.resolveSigner(coreContext, chains.BscMainnetChain().Id)
		require.ErrorContains(t, err, "signer not found")
	})
}

func Test_GetUpdatedChainClient(t *testing.T) {
	// initial parameters for core observer creation
	evmChain := chains.EthChain()
	evmChainParams := &observertypes.ChainParams{
		ChainId:                          evmChain.Id,
		ConnectorContractAddress:         testutils.StrategyManagerAddresses[evmChain.Id].Hex(),
		DelegationManagerContractAddress: testutils.DelegationManagerAddresses[evmChain.Id].Hex(),
		BallotThreshold:                  cosmossdk_io_math.LegacyOneDec(),
		MinObserverDelegation:            cosmossdk_io_math.LegacyOneDec(),
		PellTokenRechargeThreshold:       cosmossdk_io_math.NewInt(1000000),
		GasTokenRechargeThreshold:        cosmossdk_io_math.NewInt(1000000),
		PellTokenRechargeAmount:          cosmossdk_io_math.NewInt(1000000),
		GasTokenRechargeAmount:           cosmossdk_io_math.NewInt(1000000),
	}

	// new chain params in core context
	evmChainParamsNew := &observertypes.ChainParams{
		ChainId:                          evmChain.Id,
		ConfirmationCount:                10,
		GasPriceTicker:                   11,
		InTxTicker:                       12,
		OutTxTicker:                      13,
		StrategyManagerContractAddress:   testutils.OtherAddress1,
		ConnectorContractAddress:         testutils.OtherAddress2,
		DelegationManagerContractAddress: testutils.OtherAddress3,
		OutboundTxScheduleInterval:       15,
		OutboundTxScheduleLookahead:      16,
		BallotThreshold:                  cosmossdk_io_math.LegacyOneDec(),
		MinObserverDelegation:            cosmossdk_io_math.LegacyOneDec(),
		IsSupported:                      true,
		PellTokenRechargeThreshold:       cosmossdk_io_math.NewInt(1000000),
		GasTokenRechargeThreshold:        cosmossdk_io_math.NewInt(1000000),
		PellTokenRechargeAmount:          cosmossdk_io_math.NewInt(1000000),
		GasTokenRechargeAmount:           cosmossdk_io_math.NewInt(1000000),
	}

	t.Run("evm chain client should not be found", func(t *testing.T) {
		observer := MockCoreObserver(t, evmChain, evmChainParams)
		coreContext := CreateCoreContext(evmChain, evmChainParamsNew)
		// BSC chain client should not be found
		_, err := observer.resolveObserver(coreContext, chains.BscMainnetChain().Id)
		require.ErrorContains(t, err, "observer not found for chainID")
	})

	t.Run("chain params in evm chain client should be updated successfully", func(t *testing.T) {
		observer := MockCoreObserver(t, evmChain, evmChainParams)
		coreContext := CreateCoreContext(evmChain, evmChainParamsNew)
		// update evm chain client with new chain params
		chainOb, err := observer.resolveObserver(coreContext, evmChain.Id)
		require.NoError(t, err)
		require.NotNil(t, chainOb)
		require.Equal(t, evmChainParamsNew.StrategyManagerContractAddress, chainOb.GetChainParams().StrategyManagerContractAddress)
	})
}
