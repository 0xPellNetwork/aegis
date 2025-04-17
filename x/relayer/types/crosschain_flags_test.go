package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestDefaultDefaultCrosschainFlags(t *testing.T) {
	defaultCrosschainFlags := types.DefaultCrosschainFlags()

	require.Equal(t, &types.CrosschainFlags{
		IsInboundEnabled:             true,
		IsOutboundEnabled:            true,
		GasPriceIncreaseFlags:        &types.DefaultGasPriceIncreaseFlags,
		BlockHeaderVerificationFlags: &types.DefaultBlockHeaderVerificationFlags,
	}, defaultCrosschainFlags)
}
