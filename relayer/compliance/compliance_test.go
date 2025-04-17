package compliance

import (
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/config"
	"github.com/pell-chain/pellcore/relayer/testutils"
)

func TestXmsgRestricted(t *testing.T) {
	// load archived xmsg
	chain := chains.EthChain()
	xmsg := testutils.LoadXmsgByNonce(t, chain.Id, 6270)

	// create config
	cfg := config.Config{
		ComplianceConfig: config.ComplianceConfig{},
	}

	t.Run("should return true if sender is restricted", func(t *testing.T) {
		cfg.ComplianceConfig.RestrictedAddresses = []string{xmsg.InboundTxParams.Sender}
		config.LoadComplianceConfig(cfg)
		require.True(t, IsXmsgRestricted(xmsg))
	})
	t.Run("should return true if receiver is restricted", func(t *testing.T) {
		cfg.ComplianceConfig.RestrictedAddresses = []string{xmsg.GetCurrentOutTxParam().Receiver}
		config.LoadComplianceConfig(cfg)
		require.True(t, IsXmsgRestricted(xmsg))
	})
	t.Run("should return false if sender and receiver are not restricted", func(t *testing.T) {
		// restrict other address
		cfg.ComplianceConfig.RestrictedAddresses = []string{"0x27104b8dB4aEdDb054fCed87c346C0758Ff5dFB1"}
		config.LoadComplianceConfig(cfg)
		require.False(t, IsXmsgRestricted(xmsg))
	})
	t.Run("should be able to restrict coinbase address", func(t *testing.T) {
		cfg.ComplianceConfig.RestrictedAddresses = []string{ethcommon.Address{}.String()}
		config.LoadComplianceConfig(cfg)
		xmsg.InboundTxParams.Sender = ethcommon.Address{}.String()
		require.True(t, IsXmsgRestricted(xmsg))
	})
	t.Run("should ignore empty address", func(t *testing.T) {
		cfg.ComplianceConfig.RestrictedAddresses = []string{""}
		config.LoadComplianceConfig(cfg)
		xmsg.InboundTxParams.Sender = ""
		require.False(t, IsXmsgRestricted(xmsg))
	})
}
