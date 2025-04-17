
set -x

shopt -s expand_aliases
export PELL_RPC_URL="http://pellcore0:26657"
export PELL_EVM_RPC_URL="http://pellcore0:8545"
export CHAIN_ID="ignite_186-1"
alias pellcored="pellcored --chain-id $CHAIN_ID --node $PELL_RPC_URL"
alias cast="ssh hardhat cast"

# admin account setting
ADMIN_MENEMONIC="piece struggle ripple use immense kind royal grass okay mobile client head snake follow mass again candy siege announce immense hard cool unable arrange"
ADMIN_PELL_ADDRESS=pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd

# set up admin keyring
ADMIN_KEY=$(/go/bin/pellcored keys show admin --keyring-backend=test 2>/dev/null || echo "")
if [ -z "$ADMIN_KEY" ] ; then
  echo $ADMIN_MENEMONIC | /go/bin/pellcored keys add admin --recover --keyring-backend=test
fi

pellcored tx fungible deploy-pell-system-contracts --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y

sleep 3
# update chain params from json
pellcored tx observer update-and-sync-chain-params ./chains/bsc.json 97 --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y

sleep 3
PELL_DELEGATION_MANAGER_INTERACTOR=$(pellcored query fungible system-contract  |grep "pell_delegation_manager_interactor_proxy"| awk 'NR==1 {print $2}')
cast call $PELL_DELEGATION_MANAGER_INTERACTOR 'targetDelegations\(uint256\)\(uint256,address\)' 0 --rpc-url $PELL_EVM_RPC_URL
