package querytests

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcfg "github.com/evmos/ethermint/cmd/config"
	"github.com/stretchr/testify/suite"

	"github.com/0xPellNetwork/aegis/app"
	cmdcfg "github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/testutil/network"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

type CliTestSuite struct {
	suite.Suite

	cfg           network.Config
	network       *network.Network
	xmsgState     *types.GenesisState
	observerState *relayertypes.GenesisState
}

func NewCLITestSuite(cfg network.Config) *CliTestSuite {
	return &CliTestSuite{cfg: cfg}
}

func (s *CliTestSuite) Setconfig() {
	config := sdk.GetConfig()
	cmdcfg.SetBech32Prefixes(config)
	ethcfg.SetBip44CoinType(config)
	// Make sure address is compatible with ethereum
	config.SetAddressVerifier(app.VerifyAddressFormat)
	config.Seal()
}

func (s *CliTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	s.Setconfig()
	minOBsDel, ok := sdkmath.NewIntFromString("100000000000000000000")
	s.Require().True(ok)
	s.cfg.StakingTokens = minOBsDel.Mul(sdkmath.NewInt(int64(10)))
	s.cfg.BondedTokens = minOBsDel
	observerList := []string{"pell1xmfl2kmunufzcp5kuxwtm7j2wgm3z04hpapxrl",
		"pell1v6dpnhuj6ghm02vf8jv67jauapk9kca7slyul3",
	}
	network.SetupPellGenesisState(s.T(), s.cfg.GenesisState, s.cfg.Codec, observerList, false)
	s.xmsgState = network.AddXmsgData(s.T(), 2, s.cfg.GenesisState, s.cfg.Codec)
	s.observerState = network.AddObserverData(s.T(), 2, s.cfg.GenesisState, s.cfg.Codec, nil)
	net, err := network.New(s.T(), app.NodeDir, s.cfg)
	s.Assert().NoError(err)
	s.network = net
	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

}

func (s *CliTestSuite) TearDownSuite() {
	s.T().Log("tearing down genesis test suite")
	s.network.Cleanup()
}
