package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	. "gopkg.in/check.v1"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestChainParamsList_Validate(t *testing.T) {
	t.Run("should return no error for default list", func(t *testing.T) {
		list := types.GetDefaultChainParams()
		err := list.Validate()
		require.NoError(t, err)
	})

	t.Run("should return error for invalid chain id", func(t *testing.T) {
		list := types.GetDefaultChainParams()
		list.ChainParams[0].ChainId = 999
		err := list.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for duplicated chain ID", func(t *testing.T) {
		list := types.GetDefaultChainParams()
		list.ChainParams = append(list.ChainParams, list.ChainParams[0])
		err := list.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicated chain id")
	})
}

type UpdateChainParamsSuite struct {
	suite.Suite
	evmParams *types.ChainParams
	btcParams *types.ChainParams
}

var _ = Suite(&UpdateChainParamsSuite{})

func TestUpdateChainParamsSuiteSuite(t *testing.T) {
	suite.Run(t, new(UpdateChainParamsSuite))
}

func TestChainParamsEqual(t *testing.T) {
	params := types.GetDefaultChainParams()
	require.True(t, cmp.Equal(*params.ChainParams[0], *params.ChainParams[0]))
	require.False(t, cmp.Equal(*params.ChainParams[0], *params.ChainParams[1]))
}

func (s *UpdateChainParamsSuite) SetupTest() {
	s.evmParams = &types.ChainParams{
		ConfirmationCount:                        1,
		GasPriceTicker:                           1,
		InTxTicker:                               1,
		OutTxTicker:                              1,
		StrategyManagerContractAddress:           "0xA8D5060feb6B456e886F023709A2795373691E63",
		ConnectorContractAddress:                 "0x733aB8b06DDDEf27Eaa72294B0d7c9cEF7f12db9",
		DelegationManagerContractAddress:         "0xD28D6A0b8189305551a0A8bd247a6ECa9CE781Ca",
		OmniOperatorSharesManagerContractAddress: "0x733aB8b06DDDEf27Eaa72294B0d7c9cEF7f12db9",
		ChainId:                                  5,
		OutboundTxScheduleInterval:               1,
		OutboundTxScheduleLookahead:              1,
		BallotThreshold:                          types.DefaultBallotThreshold,
		MinObserverDelegation:                    types.DefaultMinObserverDelegation,
		IsSupported:                              false,
	}
	s.btcParams = &types.ChainParams{
		ConfirmationCount:                        1,
		GasPriceTicker:                           1,
		InTxTicker:                               1,
		OutTxTicker:                              1,
		StrategyManagerContractAddress:           "",
		ConnectorContractAddress:                 "",
		DelegationManagerContractAddress:         "",
		OmniOperatorSharesManagerContractAddress: "",
		ChainId:                                  18332,
		OutboundTxScheduleInterval:               1,
		OutboundTxScheduleLookahead:              1,
		BallotThreshold:                          types.DefaultBallotThreshold,
		MinObserverDelegation:                    types.DefaultMinObserverDelegation,
		IsSupported:                              false,
	}
}

func (s *UpdateChainParamsSuite) TestValidParams() {
	err := types.ValidateChainParams(s.evmParams)
	require.Nil(s.T(), err)
	err = types.ValidateChainParams(s.btcParams)
	require.Nil(s.T(), err)
}

func (s *UpdateChainParamsSuite) TestCommonParams() {
	s.Validate(s.evmParams)
	s.Validate(s.btcParams)
}

func (s *UpdateChainParamsSuite) TestCoreContractAddresses() {
	copy := *s.evmParams
	copy.StrategyManagerContractAddress = "0x123"
	err := types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.StrategyManagerContractAddress = "733aB8b06DDDEf27Eaa72294B0d7c9cEF7f12db9"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.ConnectorContractAddress = "0x123"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.ConnectorContractAddress = "733aB8b06DDDEf27Eaa72294B0d7c9cEF7f12db9"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.DelegationManagerContractAddress = "0x123"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.DelegationManagerContractAddress = "733aB8b06DDDEf27Eaa72294B0d7c9cEF7f12db9"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.OmniOperatorSharesManagerContractAddress = "0x123"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *s.evmParams
	copy.OmniOperatorSharesManagerContractAddress = "733aB8b06DDDEf27Eaa72294B0d7c9cEF7f12db9"
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
}

func (s *UpdateChainParamsSuite) Validate(params *types.ChainParams) {
	copy := *params
	copy.ConfirmationCount = 0
	err := types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *params
	copy.GasPriceTicker = 0
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.GasPriceTicker = 300
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)
	copy.GasPriceTicker = 301
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *params
	copy.InTxTicker = 0
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.InTxTicker = 300
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)
	copy.InTxTicker = 301
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *params
	copy.OutTxTicker = 0
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.OutTxTicker = 300
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)
	copy.OutTxTicker = 301
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *params
	copy.OutboundTxScheduleInterval = 0
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.OutboundTxScheduleInterval = 100
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)
	copy.OutboundTxScheduleInterval = 101
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *params
	copy.OutboundTxScheduleLookahead = 0
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.OutboundTxScheduleLookahead = 500
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)
	copy.OutboundTxScheduleLookahead = 501
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)

	copy = *params
	copy.BallotThreshold = sdkmath.LegacyDec{}
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.BallotThreshold = sdkmath.LegacyMustNewDecFromStr("1.2")
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.BallotThreshold = sdkmath.LegacyMustNewDecFromStr("0.9")
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)

	copy = *params
	copy.MinObserverDelegation = sdkmath.LegacyDec{}
	err = types.ValidateChainParams(&copy)
	require.NotNil(s.T(), err)
	copy.MinObserverDelegation = sdkmath.LegacyMustNewDecFromStr("0.9")
	err = types.ValidateChainParams(&copy)
	require.Nil(s.T(), err)
}
