# shellcheck disable=SC2155

wait_for_contracts_deployment() {
  max_tries=50
  intervals=5
  tries=0
  filepath="../../deployments/localhost/MockDVSRegistryRouter-Proxy.json"

  while [ $tries -lt $max_tries ]; do
    ret=$(ssh hardhat "if [ -f $filepath ]; then echo 0; else echo 1; fi" 2>/dev/null || echo 1)
    if [ $ret -eq 0 ]; then
      echo "$filepath found. Deployment is successful."
      return 0
    fi
    echo "Contracts not deployed after $tries tries"
    tries=$((tries + 1))
    sleep $intervals
  done

  echo "Contracts not deployed after $max_tries tries"
  exit 1
}

generate_config() {
  CONFIG_FILE="e2e_config.yml"

  cat <<EOF > $CONFIG_FILE
accounts:
  evm_address: ""
  evm_priv_key: ""
  evm_admin_priv_key: "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

rpcs:
  pevm: "http://pellcore0:8545"
  evm: "http://eth:8545"
  bitcoin:
    user: "smoketest"
    pass: "123"
    host: "bitcoin:18443"
    http_post_mode: true
    disable_tls: true
    params: "regnet"
  pellcore_grpc: "pellcore0:9090"
  pellcore_rpc: "http://pellcore0:26657"

contracts:
  evm:
    pell_eth: ""
    connector_eth: ""
    custody: ""
    erc20: "0xff3135df4F2775f4091b81f4c7B6359CfA07862a"
    st_erc20_addr: "0x610178dA211FEF7D417bC0e6FeD39F05609AD788"
    strategy_addr: "0xA51c1fc2f0D1a1b8494Ed1FE312d7C3a78Ed91C0"
    strategy_manager_addr: "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9"
    delegation_manager_addr: "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9"
    pell_omni_operator_shares_contract_addr: "0x0E801D84Fa97b50751Dbf25036d067dCf18858bF"
    tss_manager_addr: "0x70e0bA845a1A0F2DA3359C97E0285013525FFC49"
    evm_pell_connector: "0x4826533B4897376654Bb4d4AD88B7faFD0C98528"
  pevm:
    system_contract: ""
    eth_zrc20: ""
    erc20_zrc20: ""
    btc_zrc20: ""
    uniswap_factory: ""
    uniswap_router: ""
    connector_pevm: ""
    wpell: ""
    pevm_swap_app: ""
    context_app: ""
    test_dpp: ""
  pevm:
    pell_delegation_addr: ""

pell_chain_id: "ignite_186-1"
EOF

  export EVM_PELL_CONNNECTOR_ADDRESS=$(ssh hardhat cat deployments/localhost/PellConnectorOnService.json | jq -r '.address')
  export DELEGATION_MANAGER_ADDRESS=$(ssh hardhat cat deployments/localhost/DelegationManager-Proxy.json | jq -r '.address')
  export OMNI_OPERATOR_SHARES_MANAGER_ADDRESS=$(ssh hardhat cat deployments/localhost/OmniOperatorSharesManager-Proxy.json | jq -r '.address')
  export STAKING_STRATEGY_ADDRESS=$(ssh hardhat cat deployments/localhost/stBTC-Strategy-Proxy.json | jq -r '.address')
  export ST_ERC20_CONTRACT_ADDRESS=$(ssh hardhat cat deployments/localhost/stBTC-TestnetMintableERC20.json | jq -r '.address')
  export STRTEGY_MANAGER_ADDR_ADDRESS=$(ssh hardhat cat deployments/localhost/StrategyManager-Proxy.json | jq -r '.address')
  export TSS_MANAGER_ADDRESS=$(ssh hardhat cat deployments/localhost/TSSManager.json | jq -r '.address')

  yq -i ".contracts.evm.st_erc20_addr=\"$ST_ERC20_CONTRACT_ADDRESS\"" $CONFIG_FILE
  yq -i ".contracts.evm.strategy_addr=\"$STAKING_STRATEGY_ADDRESS\"" $CONFIG_FILE
  yq -i ".contracts.evm.strategy_manager_addr=\"$STRTEGY_MANAGER_ADDR_ADDRESS\"" $CONFIG_FILE
  yq -i ".contracts.evm.delegation_manager_addr=\"$DELEGATION_MANAGER_ADDRESS\"" $CONFIG_FILE
  yq -i ".contracts.evm.pell_omni_operator_shares_contract_addr=\"$OMNI_OPERATOR_SHARES_MANAGER_ADDRESS\"" $CONFIG_FILE
  yq -i ".contracts.evm.tss_manager_addr=\"$TSS_MANAGER_ADDRESS\"" $CONFIG_FILE
  yq -i ".contracts.evm.evm_pell_connector=\"$EVM_PELL_CONNNECTOR_ADDRESS\"" $CONFIG_FILE
}

wait_for_contracts_deployment
generate_config