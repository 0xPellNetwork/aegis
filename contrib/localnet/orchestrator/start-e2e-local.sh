#!/bin/sh

echo "waiting for geth RPC to start..."
# sleep 20
### Create the accounts and fund them with Ether on local Ethereum network
echo "funding deployer address 0xE5C5367B8224807Ac2207d350E60e1b6F27a7ecC with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0xE5C5367B8224807Ac2207d350E60e1b6F27a7ecC", value: web3.toWei(10000,"ether")})' attach http://eth:8545

# unlock erc20 tester accounts
echo "funding deployer address 0x6F57D5E7c6DBb75e59F1524a3dE38Fc389ec5Fd6 with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0x6F57D5E7c6DBb75e59F1524a3dE38Fc389ec5Fd6", value: web3.toWei(10000,"ether")})' attach http://eth:8545

# unlock pell tester accounts
echo "funding deployer address 0x5cC2fBb200A929B372e3016F1925DcF988E081fd with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0x5cC2fBb200A929B372e3016F1925DcF988E081fd", value: web3.toWei(10000,"ether")})' attach http://eth:8545
# unlock bitcoin tester accounts
echo "funding deployer address 0x283d810090EdF4043E75247eAeBcE848806237fD with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0x283d810090EdF4043E75247eAeBcE848806237fD", value: web3.toWei(10000,"ether")})' attach http://eth:8545

# unlock ethers tester accounts
echo "funding deployer address 0x8D47Db7390AC4D3D449Cc20D799ce4748F97619A with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0x8D47Db7390AC4D3D449Cc20D799ce4748F97619A", value: web3.toWei(10000,"ether")})' attach http://eth:8545

# unlock miscellaneous tests accounts
echo "funding deployer address 0x90126d02E41c9eB2a10cfc43aAb3BD3460523Cdf with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0x90126d02E41c9eB2a10cfc43aAb3BD3460523Cdf", value: web3.toWei(10000,"ether")})' attach http://eth:8545

# unlock admin erc20 tests accounts
echo "funding deployer address 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", value: web3.toWei(10000,"ether")})' attach http://eth:8545

# unlock the TSS account
echo "funding TSS address 0xF421292cb0d3c97b90EEEADfcD660B893592c6A2 with 10000 Ether"
geth --exec 'eth.sendTransaction({from: eth.coinbase, to: "0xF421292cb0d3c97b90EEEADfcD660B893592c6A2", value: web3.toWei(10000,"ether")})' attach http://eth:8545

pelle2e multi --init ./

yq eval '.rpcs.pell_core_rpc = "http://pellcore0:26657"' -i e2e.yaml
yq eval '.rpcs.pell_core_grpc = "pellcore0:9090"' -i e2e.yaml
yq eval '.rpcs.pell_evm = "http://pellcore0:8545"' -i e2e.yaml


wait_for_contracts_deployment() {
  max_tries=50
  intervals=5
  tries=0
  filepath="../../deployments/localhost/MockDVSServiceManager-Implementation.json"

  while [ $tries -lt $max_tries ]; do
    ret=$(ssh hardhat "if [ -f $filepath ]; then echo 0; else echo 1; fi" 2>/dev/null || echo 1)
    if [ $ret -eq 0 ]; then
      echo "$filepath found. Deployment is successful."
      return 0
    fi
    tries=$((tries + 1))
    sleep $intervals
  done

  echo "Contracts not deployed after $max_tries tries"
  exit 1
}

wait_for_contracts_deployment

export EVM_PELL_CONNNECTOR_ADDRESS=$(ssh hardhat cat deployments/localhost/PellConnectorOnService.json | jq -r '.address')
export DELEGATION_MANAGER_ADDRESS=$(ssh hardhat cat deployments/localhost/DelegationManager-Proxy.json | jq -r '.address')
export OMNI_OPERATOR_SHARES_MANAGER_ADDRESS=$(ssh hardhat cat deployments/localhost/OmniOperatorSharesManager-Proxy.json | jq -r '.address')
export STAKING_STRATEGY_ADDRESS=$(ssh hardhat cat deployments/localhost/stBTC-Strategy-Proxy.json | jq -r '.address')
export ST_ERC20_CONTRACT_ADDRESS=$(ssh hardhat cat deployments/localhost/stBTC-TestnetMintableERC20.json | jq -r '.address')
export STRTEGY_MANAGER_ADDR_ADDRESS=$(ssh hardhat cat deployments/localhost/StrategyManager-Proxy.json | jq -r '.address')
export TSS_MANAGER_ADDRESS=$(ssh hardhat cat deployments/localhost/TSSManager.json | jq -r '.address')
export PELL_TOKEN_ADDRESS=$(ssh hardhat cat deployments/localhost/PellToken.json | jq -r '.address')
export GETWAY_EVM_ADDRESS=$(ssh hardhat cat deployments/localhost/GatewayEVM.json | jq -r '.address')
export GAS_SWAP_EVM_ADDRESS=$(ssh hardhat cat deployments/localhost/GasSwapEVM.json | jq -r '.address')
export CENTRAL_SCHEDULER_ADDRESS=$(ssh hardhat cat /app/pell-middleware-contracts/deployments/localhost/CentralScheduler-Proxy.json | jq -r '.address')
export OPERATOR_STAKE_REGISTRY_ADDRESS=$(ssh hardhat cat /app/pell-middleware-contracts/deployments/localhost/OperatorStakeManager-Proxy.json | jq -r '.address')
export EJECTION_MANAGER_ADDRESS=$(ssh hardhat cat /app/pell-middleware-contracts/deployments/localhost/EjectionManager-Proxy.json | jq -r '.address')

yq eval ".multi_chain[0].connector = \"$EVM_PELL_CONNNECTOR_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].st_erc20_addr = \"$ST_ERC20_CONTRACT_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].strategy_addr = \"$STAKING_STRATEGY_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].strategy_manager_addr = \"$STRTEGY_MANAGER_ADDR_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].delegation_manager_addr = \"$DELEGATION_MANAGER_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].pell_omni_operator_shares_contract_addr = \"$OMNI_OPERATOR_SHARES_MANAGER_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].tss_manager_addr = \"$TSS_MANAGER_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].pell_token_addr = \"$PELL_TOKEN_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].gateway_evm_addr = \"$GETWAY_EVM_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].gas_swap_evm_addr = \"$GAS_SWAP_EVM_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].central_scheduler_addr = \"$CENTRAL_SCHEDULER_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].operator_stake_manager_addr = \"$OPERATOR_STAKE_REGISTRY_ADDRESS\"" -i e2e.yaml
yq eval ".multi_chain[0].ejection_manager_addr = \"$EJECTION_MANAGER_ADDRESS\"" -i e2e.yaml
yq eval '.multi_chain[0].rpc = "http://eth:8545"' -i e2e.yaml

yq eval '.setup.pevm_setup = true' -i e2e.yaml
yq eval '.setup.evm_setup = true' -i e2e.yaml

yq eval '.accounts.fungible_admin_mnemonic = "piece struggle ripple use immense kind royal grass okay mobile client head snake follow mass again candy siege announce immense hard cool unable arrange"' -i e2e.yaml

yq e2e.yaml

## 1. Setup the admin
ADMIN_MENEMONIC="piece struggle ripple use immense kind royal grass okay mobile client head snake follow mass again candy siege announce immense hard cool unable arrange"
ADMIN_PELL_ADDRESS=pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd
ADMIN_ACCOUNT="3EE2C833278B2CF9513A1918C4232E144D7AB3F9C92C655D4D968A760B944586"

ADMIN_KEY=$(pellcored keys show admin --keyring-backend=test 2>/dev/null || echo "")
if [ -z "$ADMIN_KEY" ] ; then
  echo $ADMIN_MENEMONIC | pellcored keys add admin --recover --keyring-backend=test
fi


alias pellcored="pellcored --chain-id ignite_186-1  --node http://pellcore0:26657"


waiting_for_block_height() {
  if [ -z "$1" ]; then
    echo "Error: block_height is required"
    return 1
  fi

  local block_height=$1
  while true; do
    PELL_BLOCK_HEIGHT=$(pellcored query status | jq -r '.sync_info.latest_block_height')
    echo "Waiting for pell block height to be greater than $block_height, current height: $PELL_BLOCK_HEIGHT"
    if [ -n "$PELL_BLOCK_HEIGHT" ] && [ "$PELL_BLOCK_HEIGHT" != "null" ] && [ "$PELL_BLOCK_HEIGHT" -gt $block_height ]; then
      break
    fi
    sleep 4
  done
}

waiting_for_block_height 35

pelle2e multi --config=./e2e.yaml
