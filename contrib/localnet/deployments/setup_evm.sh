# TODO: Must be an archive node's rpc
export PELL_RPC_URL="http://pellcore0:26657"
export EXTERNAL_RPC_URL="https://bsc-testnet-dataseed.bnbchain.org"
export EXTERNAL_NETWORK="bsc-testnet"
export EXTERNAL_CHAIN_ID="97"

export OPERATOR_ACCOUNT="3EE2C833278B2CF9513A1918C4232E144D7AB3F9C92C655D4D968A760B944586"

# EVM DelegationManager-Proxy contract address
export EVM_DELEGATION_MANAGER_ADDR="0x7b502746df19d64Cd824Ca0224287d06bae31DA3"
export EVM_TSS_MANAGER_ADDR="0xe02939585caA6090067B512Dd6843213aeFF4F9c"

export CHAIN_ID="ignite_186-1"
shopt -s expand_aliases
alias pellcored="pellcored --chain-id $CHAIN_ID --node $PELL_RPC_URL"
alias cast="ssh hardhat cast"

# get tss-address from pell
PELL_TSS_ADDRESS=$(pellcored query observer get-tss-address | awk 'NR==2 {print $2}')

SUGGESTED_GAS_PRICE=$(cast gas-price --rpc-url $EXTERNAL_RPC_URL)
cast send $EVM_TSS_MANAGER_ADDR 'addTSS\(address\)' $PELL_TSS_ADDRESS --rpc-url $EXTERNAL_RPC_URL --private-key $OPERATOR_ACCOUNT --gas-price $SUGGESTED_GAS_PRICE --gas-limit 200000

export TSS_MANAGER_ACCOUNT="0x3EE2C833278B2CF9513A1918C4232E144D7AB3F9C92C655D4D968A760B944586"
cast send $PELL_TSS_ADDRESS --value 10000000000000000 --private-key $TSS_MANAGER_ACCOUNT --rpc-url $EXTERNAL_RPC_URL