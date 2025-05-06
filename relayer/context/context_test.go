package context_test

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/config"
	corecontext "github.com/0xPellNetwork/aegis/relayer/context"
	clientlogs "github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func assertPanic(t *testing.T, f func(), errorLog string) {
	defer func() {
		r := recover()
		if r != nil {
			require.Contains(t, r, errorLog)
		}
	}()
	f()
}

func getTestCoreContext(
	evmChain chains.Chain,
	evmChainParams *relayertypes.ChainParams,
	ccFlags relayertypes.CrosschainFlags,
	verificationFlags lightclienttypes.VerificationFlags,
) *corecontext.PellCoreContext {
	// create config
	cfg := config.NewConfig()
	cfg.EVMChainConfigs[evmChain.Id] = config.EVMConfig{
		Chain: evmChain,
	}
	// create core context
	coreContext := corecontext.NewPellCoreContext(cfg)
	evmChainParamsMap := make(map[int64]*relayertypes.ChainParams)
	evmChainParamsMap[evmChain.Id] = evmChainParams

	// feed chain params
	coreContext.Update(
		relayertypes.Keygen{},
		[]chains.Chain{evmChain},
		evmChainParamsMap,
		"",
		ccFlags,
		verificationFlags,
		true,
		zerolog.Logger{},
	)
	return coreContext
}

func TestNewPellCoreContext(t *testing.T) {
	t.Run("should create new pell core context with empty config", func(t *testing.T) {
		testCfg := config.NewConfig()

		pellContext := corecontext.NewPellCoreContext(testCfg)
		require.NotNil(t, pellContext)

		// assert keygen
		keyGen := pellContext.GetKeygen()
		require.Equal(t, relayertypes.Keygen{}, keyGen)

		// assert enabled chains
		require.Empty(t, len(pellContext.GetEnabledChains()))

		// assert current tss pubkey
		require.Equal(t, "", pellContext.GetCurrentTssPubkey())

		// assert evm chain params
		allEVMChainParams := pellContext.GetAllEVMChainParams()
		require.Empty(t, allEVMChainParams)
	})

	t.Run("should create new pell core context with config containing evm chain params", func(t *testing.T) {
		testCfg := config.NewConfig()
		testCfg.EVMChainConfigs = map[int64]config.EVMConfig{
			1: {
				Chain: chains.Chain{
					Id: 1,
				},
			},
			2: {
				Chain: chains.Chain{
					Id: 2,
				},
			},
		}
		pellContext := corecontext.NewPellCoreContext(testCfg)
		require.NotNil(t, pellContext)

		// assert evm chain params
		allEVMChainParams := pellContext.GetAllEVMChainParams()
		require.Equal(t, 2, len(allEVMChainParams))
		require.Equal(t, &relayertypes.ChainParams{}, allEVMChainParams[1])
		require.Equal(t, &relayertypes.ChainParams{}, allEVMChainParams[2])

		evmChainParams1, found := pellContext.GetEVMChainParams(1)
		require.True(t, found)
		require.Equal(t, &relayertypes.ChainParams{}, evmChainParams1)

		evmChainParams2, found := pellContext.GetEVMChainParams(2)
		require.True(t, found)
		require.Equal(t, &relayertypes.ChainParams{}, evmChainParams2)
	})

}

func TestUpdatePellCoreContext(t *testing.T) {
	t.Run("should update core context after being created from empty config", func(t *testing.T) {
		testCfg := config.NewConfig()

		pellContext := corecontext.NewPellCoreContext(testCfg)
		require.NotNil(t, pellContext)

		keyGenToUpdate := relayertypes.Keygen{
			Status:         relayertypes.KeygenStatus_SUCCESS,
			GranteePubkeys: []string{"testpubkey1"},
		}
		enabledChainsToUpdate := []chains.Chain{
			{
				Id: 1,
			},
			{
				Id: 2,
			},
		}
		evmChainParamsToUpdate := map[int64]*relayertypes.ChainParams{
			1: {
				ChainId: 1,
			},
			2: {
				ChainId: 2,
			},
		}

		tssPubKeyToUpdate := "tsspubkeytest"
		loggers := clientlogs.DefaultLogger()
		crosschainFlags := sample.CrosschainFlags_pell()
		verificationFlags := sample.VerificationFlags()

		require.NotNil(t, crosschainFlags)
		pellContext.Update(
			keyGenToUpdate,
			enabledChainsToUpdate,
			evmChainParamsToUpdate,
			tssPubKeyToUpdate,
			*crosschainFlags,
			verificationFlags,
			false,
			loggers.Std,
		)

		// assert keygen updated
		keyGen := pellContext.GetKeygen()
		require.Equal(t, keyGenToUpdate, keyGen)

		// assert enabled chains updated
		require.Equal(t, enabledChainsToUpdate, pellContext.GetEnabledChains())

		// assert current tss pubkey updated
		require.Equal(t, tssPubKeyToUpdate, pellContext.GetCurrentTssPubkey())

		// assert evm chain params still empty because they were not specified in config
		allEVMChainParams := pellContext.GetAllEVMChainParams()
		require.Empty(t, allEVMChainParams)

		ccFlags := pellContext.GetCrossChainFlags()
		require.Equal(t, *crosschainFlags, ccFlags)

		verFlags := pellContext.GetVerificationFlags()
		require.Equal(t, verificationFlags, verFlags)
	})

	t.Run("should update core context after being created from config with evm and btc chain params", func(t *testing.T) {
		testCfg := config.NewConfig()
		testCfg.EVMChainConfigs = map[int64]config.EVMConfig{
			1: {
				Chain: chains.Chain{
					Id: 1,
				},
			},
			2: {
				Chain: chains.Chain{
					Id: 2,
				},
			},
		}
		testCfg.BitcoinConfig = config.BTCConfig{
			RPCUsername: "test username",
			RPCPassword: "test password",
			RPCHost:     "test host",
			RPCParams:   "test params",
		}

		pellContext := corecontext.NewPellCoreContext(testCfg)
		require.NotNil(t, pellContext)

		keyGenToUpdate := relayertypes.Keygen{
			Status:         relayertypes.KeygenStatus_SUCCESS,
			GranteePubkeys: []string{"testpubkey1"},
		}
		enabledChainsToUpdate := []chains.Chain{
			{
				Id: 1,
			},
			{
				Id: 2,
			},
		}
		evmChainParamsToUpdate := map[int64]*relayertypes.ChainParams{
			1: {
				ChainId: 1,
			},
			2: {
				ChainId: 2,
			},
		}

		tssPubKeyToUpdate := "tsspubkeytest"
		crosschainFlags := sample.CrosschainFlags_pell()
		verificationFlags := sample.VerificationFlags()
		require.NotNil(t, crosschainFlags)
		loggers := clientlogs.DefaultLogger()
		pellContext.Update(
			keyGenToUpdate,
			enabledChainsToUpdate,
			evmChainParamsToUpdate,
			tssPubKeyToUpdate,
			*crosschainFlags,
			verificationFlags,
			false,
			loggers.Std,
		)

		// assert keygen updated
		keyGen := pellContext.GetKeygen()
		require.Equal(t, keyGenToUpdate, keyGen)

		// assert enabled chains updated
		require.Equal(t, enabledChainsToUpdate, pellContext.GetEnabledChains())

		// assert current tss pubkey updated
		require.Equal(t, tssPubKeyToUpdate, pellContext.GetCurrentTssPubkey())

		// assert evm chain params
		allEVMChainParams := pellContext.GetAllEVMChainParams()
		require.Equal(t, evmChainParamsToUpdate, allEVMChainParams)

		evmChainParams1, found := pellContext.GetEVMChainParams(1)
		require.True(t, found)
		require.Equal(t, evmChainParamsToUpdate[1], evmChainParams1)

		evmChainParams2, found := pellContext.GetEVMChainParams(2)
		require.True(t, found)
		require.Equal(t, evmChainParamsToUpdate[2], evmChainParams2)

		ccFlags := pellContext.GetCrossChainFlags()
		require.Equal(t, ccFlags, *crosschainFlags)

		verFlags := pellContext.GetVerificationFlags()
		require.Equal(t, verFlags, verificationFlags)
	})
}

func TestIsOutboundObservationEnabled(t *testing.T) {
	// create test chain params and flags
	evmChain := chains.EthChain()
	ccFlags := *sample.CrosschainFlags_pell()
	verificationFlags := sample.VerificationFlags()
	chainParams := &relayertypes.ChainParams{
		ChainId:     evmChain.Id,
		IsSupported: true,
	}

	t.Run("should return true if chain is supported and outbound flag is enabled", func(t *testing.T) {
		coreCTX := getTestCoreContext(evmChain, chainParams, ccFlags, verificationFlags)
		require.True(t, coreCTX.IsOutboundObservationEnabled(*chainParams))
	})
	t.Run("should return false if chain is not supported yet", func(t *testing.T) {
		paramsUnsupported := &relayertypes.ChainParams{
			ChainId:     evmChain.Id,
			IsSupported: false,
		}
		coreCTXUnsupported := getTestCoreContext(evmChain, paramsUnsupported, ccFlags, verificationFlags)
		require.False(t, coreCTXUnsupported.IsOutboundObservationEnabled(*paramsUnsupported))
	})
	t.Run("should return false if outbound flag is disabled", func(t *testing.T) {
		flagsDisabled := ccFlags
		flagsDisabled.IsOutboundEnabled = false
		coreCTXDisabled := getTestCoreContext(evmChain, chainParams, flagsDisabled, verificationFlags)
		require.False(t, coreCTXDisabled.IsOutboundObservationEnabled(*chainParams))
	})
}

func TestIsInboundObservationEnabled(t *testing.T) {
	// create test chain params and flags
	evmChain := chains.EthChain()
	ccFlags := *sample.CrosschainFlags_pell()
	verificationFlags := sample.VerificationFlags()
	chainParams := &relayertypes.ChainParams{
		ChainId:     evmChain.Id,
		IsSupported: true,
	}

	t.Run("should return true if chain is supported and inbound flag is enabled", func(t *testing.T) {
		coreCTX := getTestCoreContext(evmChain, chainParams, ccFlags, verificationFlags)
		require.True(t, coreCTX.IsInboundObservationEnabled(*chainParams))
	})
	t.Run("should return false if chain is not supported yet", func(t *testing.T) {
		paramsUnsupported := &relayertypes.ChainParams{
			ChainId:     evmChain.Id,
			IsSupported: false,
		}
		coreCTXUnsupported := getTestCoreContext(evmChain, paramsUnsupported, ccFlags, verificationFlags)
		require.False(t, coreCTXUnsupported.IsInboundObservationEnabled(*paramsUnsupported))
	})
	t.Run("should return false if inbound flag is disabled", func(t *testing.T) {
		flagsDisabled := ccFlags
		flagsDisabled.IsInboundEnabled = false
		coreCTXDisabled := getTestCoreContext(evmChain, chainParams, flagsDisabled, verificationFlags)
		require.False(t, coreCTXDisabled.IsInboundObservationEnabled(*chainParams))
	})
}
