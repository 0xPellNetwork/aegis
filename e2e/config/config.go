package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/0xPellNetwork/aegis/e2e/utils"
)

const (
	DEFAULT_EVM_GAS_LIMIT      = 200000
	DEFAULT_TSS_FEE            = "1000000000000000000" // 1e16 0.01
	DEFAULT_ST_TOKEN_AMOUNT    = "1000000000000000000" // 1e16 0.01
	DEFAULT_DEPOSIT_STAKER_FEE = "1000000000000000000" // 1e16 0.01
)

type PellConfig struct {
	Accounts    TestAccounts  `yaml:"accounts"`
	Rpcs        Rpcs          `yaml:"rpcs"`
	PEVM        PEVMContracts `yaml:"pevm"`
	MultiChain  []EVM         `yaml:"multi_chain"`
	PellChainId string        `yaml:"pell_chain_id"`
	Setup       SetupFlag     `yaml:"setup"`
}

type SetupFlag struct {
	PEVMSetup bool `yaml:"pevm_setup"`
	EVMSetup  bool `yaml:"evm_setup"`
}

type EVM struct {
	Rpc string `yaml:"rpc"`
	// pell releated
	Connector                     string        `yaml:"connector"`
	StERC20Addr                   string        `yaml:"st_erc20_addr"`
	StrategyAddr                  string        `yaml:"strategy_addr"`
	StrategyManagerAddr           string        `yaml:"strategy_manager_addr"`
	DelegationManagerAddr         string        `yaml:"delegation_manager_addr"`
	OmniOperatorSharesManagerAddr string        `yaml:"pell_omni_operator_shares_contract_addr"`
	TssManagerAddr                string        `yaml:"tss_manager_addr"`
	E2ETestConfig                 E2ETestConfig `yaml:"e2e_test_config"`
	PellTokenAddr                 string        `yaml:"pell_token_addr"`
	GatewayEVMAddr                string        `yaml:"gateway_evm_addr"`
	GasSwapEVMAddr                string        `yaml:"gas_swap_evm_addr"`
	CentralSchedulerAddr          string        `yaml:"central_scheduler_addr"`
	OperatorStakeManagerAddr      string        `yaml:"operator_stake_manager_addr"`
	EjectionManagerAddr           string        `yaml:"ejection_manager_addr"`
}

// pell releated
type PEVMContracts struct {
	PellDelegationManageInteractor string `yaml:"pell_delegation_manager_interactor"`
	PellDelegationManager          string `yaml:"pell_delegation_manager"`
	PellDvsDirectory               string `yaml:"pell_dvs_directory"`
	PellRegistryRouter             string `yaml:"pell_registry_router"`
	PellSlasher                    string `yaml:"pell_slasher"`
	PellStrategyManager            string `yaml:"pell_strategy_manager"`
	PellTokenAddress               string `yaml:"pell_token_addr"`
	PellGatewayAddr                string `yaml:"pell_gateway_addr"`
	PellGasSwapAddr                string `yaml:"pell_gas_swap_addr"`
	PellRegistryRouterFactory      string `yaml:"pell_registry_router_factory"`
}

type E2ETestConfig struct {
	GasLimit uint64 `yaml:"gas_limit"`
	// Make sure that the tss address has enough tx fees
	TssFee string `yaml:"tss_fee"`
	// test pell deposit st token amount
	StTokenAmount    string `yaml:"st_token_amount"`
	DepositStakerFee string `yaml:"deposit_staker_fee"`
}

type TestAccounts struct {
	EVMAddress            string `yaml:"evm_address"`
	EVMPrivKey            string `yaml:"evm_priv_key"`
	EVMAdminPrivKey       string `yaml:"evm_admin_priv_key"`
	FungibleAdminMnemonic string `yaml:"fungible_admin_mnemonic"`
	DeployerPrivKey       string `yaml:"deployer_priv_key"`
	DeployerAddress       string `yaml:"deployer_address"`
}

type Rpcs struct {
	PellCoreRpc  string `yaml:"pell_core_rpc"`
	PellCoreGrpc string `yaml:"pell_core_grpc"`
	PellEvm      string `yaml:"pell_evm"`
}

func Default() *PellConfig {
	return &PellConfig{
		PellChainId: "ignite_186-1",
		Rpcs: Rpcs{
			PellCoreRpc:  "http://pellcore0:26657",
			PellCoreGrpc: "pellcore0:9090",
			PellEvm:      "http://pellcore0:8545",
		},
		Accounts: TestAccounts{
			EVMAddress:            "0x6F57D5E7c6DBb75e59F1524a3dE38Fc389ec5Fd6",
			EVMAdminPrivKey:       "fda3be1b1517bdf48615bdadacc1e6463d2865868dc8077d2cdcfa4709a16894",
			EVMPrivKey:            "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
			FungibleAdminMnemonic: "snow grace federal cupboard arrive fancy gym lady uniform rotate exercise either leave alien grass",
			DeployerPrivKey:       "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
			DeployerAddress:       "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		MultiChain: []EVM{
			{
				Rpc:                           "http://eth:8545",
				Connector:                     "0x6B54fCC1Fce34058C6648C9Ed1c8Ac5fe8f1E36A",
				StERC20Addr:                   "0x0C519B951759C2f98BB1281324b1663C666bE128",
				StrategyAddr:                  "0xa8694C5b6EE0C53e57F77FB9e2E6A019D2787C6F",
				StrategyManagerAddr:           "0x05946993d6260eb0b2131aF58d140649dcA643Bf",
				DelegationManagerAddr:         "0x7b502746df19d64Cd824Ca0224287d06bae31DA3",
				OmniOperatorSharesManagerAddr: "0x173D5e14DB039745b69A03A9953bD5156975f358",
				TssManagerAddr:                "0xe02939585caA6090067B512Dd6843213aeFF4F9c",
				E2ETestConfig: E2ETestConfig{
					GasLimit:         DEFAULT_EVM_GAS_LIMIT,
					TssFee:           DEFAULT_TSS_FEE,         // 1e16 = 0.01
					StTokenAmount:    DEFAULT_ST_TOKEN_AMOUNT, // 1e16 0.01
					DepositStakerFee: DEFAULT_DEPOSIT_STAKER_FEE,
				},
			},
		},

		PEVM: PEVMContracts{
			PellDelegationManager:     "0x8A46B8916f768FD93f621350Cf290f5bb3310B24",
			PellDvsDirectory:          "0xdb0F054CCA62451424Dc9f3871e32E2A2e2C67B2",
			PellRegistryRouter:        "0xa7a437cdD31c7128C19FBF0414c205CF5ff4c6B8",
			PellSlasher:               "0xCCB019e26DAc3f5AD6b67526d36b27d82C6102A2",
			PellStrategyManager:       "0xF93dA256989C38D6a8d54Fc1252534245e869A9a",
			PellTokenAddress:          "0xF3A1F6CDf0f939D86B644D78AeAA620f67bc0EfC",
			PellGatewayAddr:           "0x8807F84c85096780438CfA93D2BbE765Ecc9320a",
			PellGasSwapAddr:           "0x7690c36D2fEd7786beAaeFb84bd1e1d2b335cf0d",
			PellRegistryRouterFactory: "0xe57e4f1565876F2BcEc80f105BF234BB099E4928",
		},
		Setup: SetupFlag{PEVMSetup: false, EVMSetup: false},
	}
}

func (conf *PellConfig) Export(path string) error {
	if path == "" {
		return errors.New("file name cannot be empty")
	}

	filePath := filepath.Join(path, "e2e.yaml")
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := yaml.Marshal(*conf)
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// GetConfig returns config from file from the command line flag
func LoadConfig(configFile string) (*PellConfig, error) {
	// use default config if no config file is specified
	if configFile == "" {
		return Default(), nil
	}

	configFile, err := filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := new(PellConfig)
	if err := yaml.Unmarshal(b, config); err != nil {
		return nil, err
	}

	config.Validate()

	return config, nil
}

func (conf *PellConfig) Validate() {
	utils.Assert(conf.PellChainId != "", "conf.PellChainId invalid")
	// rpc node validate
	utils.Assert(conf.Rpcs.PellCoreRpc != "", "conf.PellCoreRpc invalid")
	utils.Assert(conf.Rpcs.PellCoreGrpc != "", "conf.PellCoreGrpc invalid")
	utils.Assert(conf.Rpcs.PellEvm != "", "conf.PellEvm invalid")

	// test account validate
	utils.Assert(utils.IsEthAddress(conf.Accounts.EVMAddress), "conf.Accounts.EVMAddress invalid")
	utils.Assert(conf.Accounts.EVMAdminPrivKey != "", "conf.Accounts.EVMAdminPrivKey invalid")
	utils.Assert(conf.Accounts.EVMPrivKey != "", "conf.Accounts.EVMPrivKey invalid")
	utils.Assert(conf.Accounts.FungibleAdminMnemonic != "", "conf.Accounts.FungibleAdminMnemonic invalid")
	utils.Assert(conf.Accounts.DeployerPrivKey != "", "conf.Accounts.DeployerPrivKey invalid")
	utils.Assert(utils.IsEthAddress(conf.Accounts.DeployerAddress), "conf.Accounts.DeployerAddress invalid")

	for _, chain := range conf.MultiChain {
		utils.Assert(utils.IsEthAddress(chain.Connector), "conf.Accounts.Connector address invalid")
		utils.Assert(utils.IsEthAddress(chain.StERC20Addr), "conf.Accounts.StERC20Addr address invalid")
		utils.Assert(utils.IsEthAddress(chain.StrategyAddr), "conf.Accounts.StrategyAddr address invalid")
		utils.Assert(utils.IsEthAddress(chain.StrategyManagerAddr), "conf.Accounts.StrategyManagerAddr address invalid")
		utils.Assert(utils.IsEthAddress(chain.DelegationManagerAddr), "conf.Accounts.DelegationManagerAddr address invalid")
		utils.Assert(utils.IsEthAddress(chain.OmniOperatorSharesManagerAddr), "conf.Accounts.OmniOperatorSharesManagerAddr address invalid")
		utils.Assert(utils.IsEthAddress(chain.TssManagerAddr), "conf.Accounts.TssManagerAddr address invalid")

		utils.Assert(chain.E2ETestConfig.DepositStakerFee != "", "chain.E2ETestConfig.DepositStakerFee invalid")
		utils.Assert(chain.E2ETestConfig.StTokenAmount != "", "chain.E2ETestConfig.StTokenAmount invalid")
		utils.Assert(chain.E2ETestConfig.TssFee != "", "chain.E2ETestConfig.TssFee invalid")
		utils.Assert(chain.E2ETestConfig.GasLimit != 0, "chain.E2ETestConfig.GasLimit invalid")
	}
}
