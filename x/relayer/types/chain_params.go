package types

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethchains "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

const (
	zeroAddress = "0x0000000000000000000000000000000000000000"
)

var (
	DefaultMinObserverDelegation = math.LegacyMustNewDecFromStr("1000000000000000000000")
	DefaultBallotThreshold       = math.LegacyMustNewDecFromStr("0.66")
)

// Validate checks all chain params correspond to a chain and there is no duplicate chain id
func (cpl ChainParamsList) Validate() error {
	// check all chain params correspond to a chain
	chainMap := make(map[int64]struct{})
	existingChainMap := make(map[int64]struct{})

	externalChainList := chains.ChainsList()
	for _, chain := range externalChainList {
		chainMap[chain.Id] = struct{}{}
	}

	// validate the chain params and check for duplicates
	for _, chainParam := range cpl.ChainParams {
		if err := ValidateChainParams(chainParam); err != nil {
			return err
		}

		if _, ok := chainMap[chainParam.ChainId]; !ok {
			return fmt.Errorf("chain id %d not found in chain list", chainParam.ChainId)
		}
		if _, ok := existingChainMap[chainParam.ChainId]; ok {
			return fmt.Errorf("duplicated chain id %d found", chainParam.ChainId)
		}
		existingChainMap[chainParam.ChainId] = struct{}{}
	}
	return nil
}

// ValidateChainParams performs some basic checks on chain params
func ValidateChainParams(params *ChainParams) error {
	if params == nil {
		return fmt.Errorf("chain params cannot be nil")
	}
	chain, exist := chains.GetChainByChainId(params.ChainId)
	if !exist {
		return fmt.Errorf("ChainId %d not supported", params.ChainId)
	}
	// pell chain skips the rest of the checks for now
	if chain.IsPellChain() {
		return nil
	}

	if params.ConfirmationCount == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ConfirmationCount must be greater than 0")
	}
	if params.GasPriceTicker <= 0 || params.GasPriceTicker > 300 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "GasPriceTicker %d out of range", params.GasPriceTicker)
	}
	if params.InTxTicker <= 0 || params.InTxTicker > 300 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "InTxTicker %d out of range", params.InTxTicker)
	}
	if params.OutTxTicker <= 0 || params.OutTxTicker > 300 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "OutTxTicker %d out of range", params.OutTxTicker)
	}
	if params.OutboundTxScheduleInterval == 0 || params.OutboundTxScheduleInterval > 100 { // 600 secs
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "OutboundTxScheduleInterval %d out of range", params.OutboundTxScheduleInterval)
	}
	if params.OutboundTxScheduleLookahead == 0 || params.OutboundTxScheduleLookahead > 500 { // 500 xmsgs
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "OutboundTxScheduleLookahead %d out of range", params.OutboundTxScheduleLookahead)
	}

	if chains.IsEVMChain(params.ChainId) {
		if params.DelegationManagerContractAddress == "" && params.ConnectorContractAddress == "" {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid both stake layer contracts and service layer contracts")
		}

		if params.DelegationManagerContractAddress != "" {
			if !validChainContractAddress(params.StrategyManagerContractAddress) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid StrategyManagerContractAddress %s", params.StrategyManagerContractAddress)
			}
			if !validChainContractAddress(params.DelegationManagerContractAddress) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid DelegationManagerContractAddress %s", params.DelegationManagerContractAddress)
			}
		}

		if params.ConnectorContractAddress != "" {
			if !validChainContractAddress(params.ConnectorContractAddress) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid ConnectorContractAddress %s", params.ConnectorContractAddress)
			}
			if !validChainContractAddress(params.OmniOperatorSharesManagerContractAddress) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid OmniOperatorSharesManagerContractAddress %s", params.OmniOperatorSharesManagerContractAddress)
			}
		}

	}

	if params.BallotThreshold.IsNil() || params.BallotThreshold.GT(sdkmath.LegacyOneDec()) {
		return ErrParamsThreshold
	}

	if params.MinObserverDelegation.IsNil() {
		return ErrParamsMinObserverDelegation
	}

	return nil
}

func validChainContractAddress(address string) bool {
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	return ethchains.IsHexAddress(address)
}

// GetDefaultChainParams returns a list of default chain params
// TODO: remove this function
func GetDefaultChainParams() ChainParamsList {
	return ChainParamsList{
		ChainParams: []*ChainParams{
			GetDefaultEthMainnetChainParams(),
			GetDefaultBscMainnetChainParams(),
			GetDefaultGoerliTestnetChainParams(),
			GetDefaultBscTestnetChainParams(),
			GetDefaultMumbaiTestnetChainParams(),
			GetDefaultGoerliLocalnetChainParams(),
			GetDefaultPellPrivnetChainParams(),
		},
	}
}

func GetDefaultEthMainnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:                                  chains.EthChain().Id,
		ConfirmationCount:                        14,
		StrategyManagerContractAddress:           zeroAddress,
		DelegationManagerContractAddress:         zeroAddress,
		OmniOperatorSharesManagerContractAddress: zeroAddress,
		ConnectorContractAddress:                 zeroAddress,
		InTxTicker:                               12,
		OutTxTicker:                              15,
		GasPriceTicker:                           30,
		OutboundTxScheduleInterval:               30,
		OutboundTxScheduleLookahead:              60,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
		GasLimit:                                 200000,
	}
}

func GetDefaultBscMainnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:                                  chains.BscMainnetChain().Id,
		ConfirmationCount:                        14,
		StrategyManagerContractAddress:           zeroAddress,
		DelegationManagerContractAddress:         zeroAddress,
		OmniOperatorSharesManagerContractAddress: zeroAddress,
		ConnectorContractAddress:                 zeroAddress,
		InTxTicker:                               5,
		OutTxTicker:                              15,
		GasPriceTicker:                           30,
		OutboundTxScheduleInterval:               30,
		OutboundTxScheduleLookahead:              60,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
		GasLimit:                                 200000,
	}
}

func GetDefaultGoerliTestnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:           chains.GoerliChain().Id,
		ConfirmationCount: 6,
		// This is the actual Pell token Goerli testnet, we need to specify this address for the integration tests to pass
		StrategyManagerContractAddress:           zeroAddress,
		DelegationManagerContractAddress:         zeroAddress,
		OmniOperatorSharesManagerContractAddress: zeroAddress,
		ConnectorContractAddress:                 zeroAddress,
		InTxTicker:                               12,
		OutTxTicker:                              15,
		GasPriceTicker:                           30,
		OutboundTxScheduleInterval:               30,
		OutboundTxScheduleLookahead:              60,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
	}
}

func GetDefaultBscTestnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:                                  chains.BscTestnetChain().Id,
		ConfirmationCount:                        6,
		StrategyManagerContractAddress:           "0x05946993d6260eb0b2131aF58d140649dcA643Bf",
		DelegationManagerContractAddress:         "0x7b502746df19d64Cd824Ca0224287d06bae31DA3",
		OmniOperatorSharesManagerContractAddress: "0x7a51fA37783EA8A2A319859984A4ED4AEBcE5d9A",
		ConnectorContractAddress:                 "0x6B54fCC1Fce34058C6648C9Ed1c8Ac5fe8f1E36A",
		InTxTicker:                               5,
		OutTxTicker:                              15,
		GasPriceTicker:                           30,
		OutboundTxScheduleInterval:               30,
		OutboundTxScheduleLookahead:              60,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
		GasLimit:                                 200000,
	}
}

func GetDefaultMumbaiTestnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:                                  chains.MumbaiChain().Id,
		ConfirmationCount:                        12,
		StrategyManagerContractAddress:           zeroAddress,
		DelegationManagerContractAddress:         zeroAddress,
		OmniOperatorSharesManagerContractAddress: zeroAddress,
		ConnectorContractAddress:                 zeroAddress,
		InTxTicker:                               2,
		OutTxTicker:                              15,
		GasPriceTicker:                           30,
		OutboundTxScheduleInterval:               30,
		OutboundTxScheduleLookahead:              60,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
		GasLimit:                                 200000,
	}
}

func GetDefaultGoerliLocalnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:                                  chains.GoerliLocalnetChain().Id,
		ConfirmationCount:                        1,
		StrategyManagerContractAddress:           "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
		DelegationManagerContractAddress:         "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
		OmniOperatorSharesManagerContractAddress: "0x809d550fca64d94Bd9F66E60752A544199cfAC3D",
		ConnectorContractAddress:                 "0x5eb3Bc0a489C5A8288765d2336659EbCA68FCd00",
		InTxTicker:                               2,
		OutTxTicker:                              1,
		GasPriceTicker:                           5,
		WatchPellTokenTicker:                     20,
		WatchGasTokenTicker:                      20,
		OutboundTxScheduleInterval:               1,
		OutboundTxScheduleLookahead:              50,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
		GasLimit:                                 2000000,
		PellTokenRechargeThreshold:               math.NewInt(1000000000000000000), // 1e18
		GasTokenRechargeThreshold:                math.NewInt(1000000000000000000),
		PellTokenRechargeAmount:                  math.NewInt(1000000000000000000),
		GasTokenRechargeAmount:                   math.NewInt(1000000000000000000),
		GatewayEvmContractAddress:                "0x1291Be112d480055DaFd8a610b7d1e203891C274",
		PellTokenContractAddress:                 "0x4c5859f0F772848b2D91F1D83E2Fe57935348029",
		PellTokenPostInterval:                    10 * 60, // 10 mins
		PellTokenRechargeEnabled:                 false,
		GasTokenPostInterval:                     10 * 60, // 10 mins
		GasTokenRechargeEnabled:                  false,
		GasSwapContractAddress:                   "0x1fA02b2d6A771842690194Cf62D91bdd92BfE28d",
		ChainRegistryInteractorContractAddress:   "0xdbC43Ba45381e02825b14322cDdd15eC4B3164E6",
	}
}

func GetDefaultPellPrivnetChainParams() *ChainParams {
	return &ChainParams{
		ChainId:                                  chains.PellPrivnetChain().Id,
		ConfirmationCount:                        1,
		StrategyManagerContractAddress:           zeroAddress,
		DelegationManagerContractAddress:         zeroAddress,
		OmniOperatorSharesManagerContractAddress: zeroAddress,
		ConnectorContractAddress:                 zeroAddress,
		InTxTicker:                               2,
		OutTxTicker:                              2,
		GasPriceTicker:                           5,
		WatchPellTokenTicker:                     10,
		WatchGasTokenTicker:                      10,
		OutboundTxScheduleInterval:               0,
		OutboundTxScheduleLookahead:              0,
		BallotThreshold:                          DefaultBallotThreshold,
		MinObserverDelegation:                    DefaultMinObserverDelegation,
		IsSupported:                              false,
		GasLimit:                                 200000,
		PellTokenRechargeThreshold:               math.NewInt(1000000000000000000),
		GasTokenRechargeThreshold:                math.NewInt(1000000000000000000),
		PellTokenRechargeAmount:                  math.NewInt(1000000000000000000),
		GasTokenRechargeAmount:                   math.NewInt(1000000000000000000),
	}
}
