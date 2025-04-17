



shopt -s expand_aliases

export PELL_RPC_URL="http://pellcore0:26657"
export PELL_EVM_URL="http://pellcore0:8545"
export CHAIN_ID="ignite_186-1"
alias pellcored="pellcored --chain-id $CHAIN_ID --node $PELL_RPC_URL"
alias hardhat="ssh hardhat ACCOUNT_SECRETKEY=$OPERATOR_ACCOUNT STAKER_SECRETKEY=$STAKER_ACCOUNT npx hardhat"
alias hardhat-mw="ssh hardhat \"cd ../.. && ACCOUNT_SECRETKEY=$OPERATOR_ACCOUNT STAKER_SECRETKEY=$STAKER_ACCOUNT npx hardhat\""
alias cast="ssh hardhat cast"

echo $pellcored

source ./utils.sh
## 0. Confirm that the block has been mined
waiting_for_block_height 0

## 1. Setup the admin
ADMIN_MENEMONIC="piece struggle ripple use immense kind royal grass okay mobile client head snake follow mass again candy siege announce immense hard cool unable arrange"
ADMIN_PELL_ADDRESS=pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd
ADMIN_ACCOUNT="3EE2C833278B2CF9513A1918C4232E144D7AB3F9C92C655D4D968A760B944586"

ADMIN_KEY=$(/go/bin/pellcored keys show admin --keyring-backend=test 2>/dev/null || echo "")
if [ -z "$ADMIN_KEY" ] ; then
  echo $ADMIN_MENEMONIC | /go/bin/pellcored keys add admin --recover --keyring-backend=test
fi

## 2. Deploy the system contracts
pellcored tx fungible deploy-pell-system-contracts --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y
PELL_DELEGATION_ADDRESS=$(pellcored query fungible system-contract  |grep "pell_delegation_manager_proxy"| awk 'NR==1 {print $2}')
PELL_DVS_DIRECTORY=$(pellcored query fungible system-contract  |grep "pell_dvs_directory_proxy"| awk 'NR==1 {print $2}')
PELL_REGISTRY_ROUTER=$(pellcored query fungible system-contract  |grep "pell_registry_router"| awk 'NR==1 {print $2}')
PELL_REGISTRY_ROUTER_FACTORY=$(pellcored query fungible system-contract |grep "pell_registry_router_factory"| awk 'NR==1 {print $2}')
PELL_STRATEGY_MGR=$(pellcored query fungible system-contract  |grep "pell_strategy_manager_proxy"| awk 'NR==1 {print $2}')
PELL_DELEGATION_MANAGER_INTERACTOR=$(pellcored query fungible system-contract  |grep "pell_delegation_manager_interactor_proxy"| awk 'NR==1 {print $2}')
# pellcored query observer show-chain-params  97| yq '.chain_params' | yq eval -o=json > output.json

## 3. Update TSS Address
waiting_for_block_height 25
TSS_ADDRESS=$(pellcored query observer get-tss-address | awk 'NR==2 {print $2}')

cast send $TSS_ADDRESS --value 10000000000000000 --private-key $TSS_MANAGER_ACCOUNT --rpc-url $EXTERNAL_RPC_URL

TSS_BALANCE=$(cast balance $TSS_ADDRESS --rpc-url $EXTERNAL_RPC_URL)
assert_number_gt_zero $TSS_BALANCE