
set -x
set -e
set -o pipefail

source ./setup-pell.sh

# Variables should be defined in the script
# STAKER_ADDRESS=
# STAKER_ACCOUNT=

# Deployer is also the operator
# OPERATOR_PELL_ADDRESS=
# OPERATOR_ADDRESS=
# OPERATOR_ACCOUNT=
# TSS_MANAGER_ACCOUNT=

# TSS_MANAGER_CONTRACT_ADDRESS=
# STAKING_STRATEGY_ADDRESS=
# STAKING_STRATEGY_MGR_ADDRESS=
# STAKING_DELEGATION_MGR_ADDRESS=
# REGISTRY_COORDINATOR_ADDRESS=
# SERVICE_OMNI_OPERATOR_SHARES_MGR_ADDRESS=

# EXTERNAL_RPC_URL
# EXTERNAL_NETWORK
# EXTERNAL_CHAIN_ID

## 1. Mock Deposit
hardhat --network $EXTERNAL_NETWORK mock-deposit
STAKER_DEPOSIT=$(cast call $STAKING_STRATEGY_MGR_ADDRESS 'stakerStrategyShares\(address,address\)\(uint256\)' $STAKER_ADDRESS $STAKING_STRATEGY_ADDRESS --rpc-url $EXTERNAL_RPC_URL | awk 'NR==1 {print $1}')
assert_number_gt_zero $STAKER_DEPOSIT

STAKER_DEPOSIT=$(cast call $PELL_STRATEGY_MGR 'stakerStrategyShares\(uint256,address,address\)\(uint256\)' 97 $STAKER_ADDRESS $STAKING_STRATEGY_ADDRESS --rpc-url $PELL_EVM_URL | awk 'NR==1 {print $1}')

## 2. Mock register as operator
pellcored tx bank send $ADMIN_PELL_ADDRESS $OPERATOR_PELL_ADDRESS 100pell --fees=0.1pell --gas=200000 --keyring-backend=test -y
sleep 5
hardhat --network pell-goerli mock-pell-register-as-operator --address $PELL_DELEGATION_ADDRESS
sleep 20

# check operator registered
IS_PELL_OPERATOR=$(cast call $PELL_DELEGATION_ADDRESS 'isOperator\(address\)\(bool\)' $OPERATOR_ADDRESS --rpc-url $PELL_EVM_URL)
assert_true $IS_PELL_OPERATOR

IS_STAKING_OPERATOR=$(cast call $STAKING_DELEGATION_MGR_ADDRESS 'isOperator\(address\)\(bool\)' $OPERATOR_ADDRESS --rpc-url $EXTERNAL_RPC_URL)
assert_true $IS_STAKING_OPERATOR

##3. Mock delegate
hardhat --network $EXTERNAL_NETWORK mock-delegate
sleep 10

OPERATOR_SHARES=$(cast call $STAKING_DELEGATION_MGR_ADDRESS 'getOperatorShares\(address,address[]\)\(uint256[]\)' $OPERATOR_ADDRESS "[$STAKING_STRATEGY_ADDRESS]" --rpc-url $EXTERNAL_RPC_URL | tr -d '[]' | awk 'NR==1 {print $1}')
assert_number_gt_zero $OPERATOR_SHARES

OMNI_OPERATOR_SHARES=$(cast call $SERVICE_OMNI_OPERATOR_SHARES_MGR_ADDRESS 'getOperatorShares\(address,\(uint256,address\)[]\)\(uint256[]\)' $OPERATOR_ADDRESS "[\(1337,$STAKING_STRATEGY_ADDRESS\)]" --rpc-url $EXTERNAL_RPC_URL | tr -d '[]' | awk 'NR==1 {print $1}')
assert_number_gt_zero $OMNI_OPERATOR_SHARES

assert_equal $OPERATOR_SHARES $OMNI_OPERATOR_SHARES

##4. Mock create DVS
hardhat-mw --network pell-goerli mock-create-quorum --address $PELL_REGISTRY_ROUTER_FACTORY
REGISTRY_ROUTER_ADDRESS=$(ssh hardhat "cd ../.. && cat deployments/pell-goerli/MockDVSRegistryRouter-Proxy.json 2> /dev/null" | jq -r '.address')
assert_not_null $REGISTRY_ROUTER_ADDRESS

##5. Mock register DVS
hardhat-mw --network pell-goerli mock-operator-register-dvs --address $REGISTRY_ROUTER_ADDRESS
OPERATOR_STATUS=$(cast call $REGISTRY_ROUTER_ADDRESS 'getOperatorStatus\(address\)\(uint256\)' $OPERATOR_ADDRESS --rpc-url $PELL_EVM_URL)
assert_equal $OPERATOR_STATUS 1