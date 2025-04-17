package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestValidateAddressForChain(t *testing.T) {
	// test for eth chain
	require.Error(t, types.ValidateAddressForChain("0x123", chains.GoerliChain().Id))
	require.Error(t, types.ValidateAddressForChain("", chains.GoerliChain().Id))
	require.Error(t, types.ValidateAddressForChain("%%%%", chains.GoerliChain().Id))
	require.NoError(t, types.ValidateAddressForChain("0x792c127Fa3AC1D52F904056Baf1D9257391e7D78", chains.GoerliChain().Id))

	// test for pell chain
	require.NoError(t, types.ValidateAddressForChain("bcrt1qs758ursh4q9z627kt3pp5yysm78ddny6txaqgw", chains.PellChainMainnet().Id))
	require.NoError(t, types.ValidateAddressForChain("0x792c127Fa3AC1D52F904056Baf1D9257391e7D78", chains.PellChainMainnet().Id))
}

func TestValidatePellIndex(t *testing.T) {
	require.NoError(t, types.ValidatePellIndex("0x84bd5c9922b63c52d8a9ca686e0a57ff978150b71be0583514d01c27aa341910"))
	require.NoError(t, types.ValidatePellIndex(sample.PellIndex(t)))
	require.Error(t, types.ValidatePellIndex("0"))
	require.Error(t, types.ValidatePellIndex("0x70e967acFcC17c3941E87562161406d41676FD83"))
}

func TestValidateHashForChain(t *testing.T) {
	require.NoError(t, types.ValidateHashForChain("0x84bd5c9922b63c52d8a9ca686e0a57ff978150b71be0583514d01c27aa341910", chains.GoerliChain().Id))
	require.Error(t, types.ValidateHashForChain("", chains.GoerliChain().Id))
	require.Error(t, types.ValidateHashForChain("a0fa5a82f106fb192e4c503bfa8d54b2de20a821e09338094ab825cc9b275059", chains.GoerliChain().Id))
}
