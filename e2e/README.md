# `e2e`

`e2e` is a comprehensive suite of E2E tests designed to validate the integration and functionality of the PellChain network, particularly its interactions with Bitcoin and EVM (Ethereum Virtual Machine) networks. This tool is essential for ensuring the robustness and reliability of PellChain's cross-chain functionalities.

## Packages

The E2E testing project is organized into several packages, each with a specific role:

- `config`: Provides general configuration for E2E tests, including RPC addresses for connected networks, addresses of deployed smart contracts, and account details for test transactions.
- `contracts`: Includes sample Solidity smart contracts used in testing scenarios.
- `runner`: Responsible for executing E2E tests, handling interactions with various network clients.
- `e2etests`: Houses a collection of E2E tests that can be run against the PellChain network.
- `txserver`: A minimalistic client for interacting with the PellChain RPC interface.
- `utils`: Offers utility functions to facilitate interactions with the different blockchain networks involved in testing.

## Config

The E2E testing suite utilizes a flexible and comprehensive configuration system defined in the config package, which is central to setting up and customizing your test environments. The configuration is structured as follows:

A config YAML file can be provided to the E2E test tool via the `--config` flag. If no config file is provided, the tool will use default values for all configuration parameters.

### Config Structure

- `RPCs`: Defines the RPC endpoints for various networks involved in the testing.
- `Contracts`: Specifies the addresses of pre-deployed smart contracts relevant to the tests.
- `PellChainID`: The specific chain ID of the PellChain network being tested.

### RPCs Configuration

- `Pevm`: RPC endpoint for the PellChain EVM.
- `EVM`: RPC endpoint for the Ethereum network.
- `Bitcoin`: RPC endpoint for the Bitcoin network.
- `PellCoreGRPC`: GRPC endpoint for PellCore.
- `PellCoreRPC`: RPC endpoint for PellCore.

### Contracts Configuration

**EVM Contracts**

- `PellEthAddress`: Address of Pell token contract on EVM chain.
- `ConnectorEthAddr`: Address of a connector contract on EVM chain.
- `ERC20`: Address of the ERC20 token contract on EVM chain.

### Config Example

```yaml
accounts:
  evm_address: 0x6F57D5E7c6DBb75e59F1524a3dE38Fc389ec5Fd6
  evm_priv_key: ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
  evm_admin_priv_key: fda3be1b1517bdf48615bdadacc1e6463d2865868dc8077d2cdcfa4709a16894
  fungible_admin_mnemonic: piece struggle ripple use immense kind royal grass okay mobile client head snake follow mass again candy siege announce immense hard cool unable arrange
  deployer_priv_key: 3ee2c833278b2cf9513a1918c4232e144d7ab3f9c92c655d4d968a760b944586
  deployer_address: 0x38B50390f1e5a54Bf2a712b012464AE76771adB5
rpcs:
  pell_core_rpc: http://pellcored0:26657
  pell_core_grpc: pellcored0:9090
  pell_evm: http://pellcored0:8545
  evm: https://eth:8545
pevm:
  pell_delegation_manager_interactor: 0x13A0c5930C028511Dc02665E7285134B6d11A5f4
  pell_delegation_manager: 0xd97B1de3619ed2c6BEb3860147E30cA8A7dC9891
  pell_dvs_directory: 0xcC683A782f4B30c138787CB5576a86AF66fdc31d
  pell_registry_router: 0x777915D031d1e8144c90D025C594b3b8Bf07a08d
  pell_slasher: 0x48f80608B672DC30DC7e3dbBd0343c5F02C738Eb
  pell_strategy_manager: 0x91d18e54DAf4F677cB28167158d6dd21F6aB3921
  pell_token_addr: 0x0C519B951759C2f98BB1281324b1663C666bE128
  gateway_addr: 0xF3A1F6CDf0f939D86B644D78AeAA620f67bc0EfC
evm:
  connector: 0x6B54fCC1Fce34058C6648C9Ed1c8Ac5fe8f1E36A
  st_erc20_addr: 0x0C519B951759C2f98BB1281324b1663C666bE128
  strategy_addr: 0xa8694C5b6EE0C53e57F77FB9e2E6A019D2787C6F
  strategy_manager_addr: 0x05946993d6260eb0b2131aF58d140649dcA643Bf
  delegation_manager_addr: 0x7b502746df19d64Cd824Ca0224287d06bae31DA3
  pell_omni_operator_shares_contract_addr: 0x173D5e14DB039745b69A03A9953bD5156975f358
  tss_manager_addr: 0xe02939585caA6090067B512Dd6843213aeFF4F9c
  evm_pell_token_addr: 0x0C519B951759C2f98BB1281324b1663C666bE128
  gateway_evm_addr: 0x809d550fca64d94Bd9F66E60752A544199cfAC3D
pell_chain_id: ignite_186-1

```

NOTE: config is in progress, contracts on the pEVM must be added
