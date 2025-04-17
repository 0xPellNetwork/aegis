package base_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/chains/base"
	"github.com/pell-chain/pellcore/relayer/logs"
	"github.com/pell-chain/pellcore/relayer/metrics"
	"github.com/pell-chain/pellcore/relayer/testutils/stub"
)

// createSigner creates a new signer for testing
func createSigner(_ *testing.T) *base.Signer {
	// constructor parameters
	chain := chains.EthChain()
	tss := stub.NewTSSMainnet()
	logger := logs.DefaultLogger()

	// create signer
	return base.NewSigner(chain, tss, nil, nil, logger)
}

func TestNewSigner(t *testing.T) {
	signer := createSigner(t)
	require.NotNil(t, signer)
}

func TestSignerGetterAndSetter(t *testing.T) {
	t.Run("should be able to update chain", func(t *testing.T) {
		signer := createSigner(t)

		// update chain
		newChain := chains.BscMainnetChain()
		signer = signer.WithChain(chains.BscMainnetChain())
		require.Equal(t, &newChain, signer.Chain())
	})
	t.Run("should be able to update tss", func(t *testing.T) {
		signer := createSigner(t)

		// update tss
		newTSS := stub.NewTSSIgnite3()
		signer = signer.WithTSS(newTSS)
		require.Equal(t, newTSS, signer.TSS())
	})
	t.Run("should be able to update telemetry server", func(t *testing.T) {
		signer := createSigner(t)

		// update telemetry server
		newTs := metrics.NewTelemetryServer()
		signer = signer.WithTelemetryServer(newTs)
		require.Equal(t, newTs, signer.TelemetryServer())
	})
	t.Run("should be able to get logger", func(t *testing.T) {
		ob := createSigner(t)
		logger := ob.Logger()

		// should be able to print log
		logger.Std.Info().Msg("print standard log")
		logger.Compliance.Info().Msg("print compliance log")
	})
}

func Test_BeingReportedFlag(t *testing.T) {
	signer := createSigner(t)

	// hash to be reported
	hash := "0x1234"
	alreadySet := signer.SetBeingReportedFlag(hash)
	require.False(t, alreadySet)

	// set reported outbound again and check
	alreadySet = signer.SetBeingReportedFlag(hash)
	require.True(t, alreadySet)

	// clear reported outbound and check again
	signer.ClearBeingReportedFlag(hash)
	alreadySet = signer.SetBeingReportedFlag(hash)
	require.False(t, alreadySet)
}
