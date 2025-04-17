#!/bin/env bash

set -e

set -x

source /root/.env.localnet

echo $PELL_CHAIN_ID
echo $PELL_ADMIN_MENEMONIC

alias pellcored="pellcored --chain-id $PELL_CHAIN_ID --node $PELL_RPC_URL"
shopt -s expand_aliases



# Common function to check transaction result, exits script directly if failed
check_tx_result() {
    # Read from stdin if no argument is provided
    local result
    if [ -z "$1" ]; then
        result=$(cat)
    else
        result="$1"
    fi

    # Extract code value from result
    local code=$(echo "$result" | grep "code:" | awk '{print $2}')

    # Check if code is not 0
    if [ "$code" != "0" ]; then
        # Only extract txhash when transaction fails
        local txhash=$(echo "$result" | grep "txhash:" | awk '{print $2}')
        echo "Transaction failed, code: $code"
        echo "Transaction hash: $txhash"
        echo "Full output:"
        echo "$result"
        exit 1  # Exit script with status code 1
    fi

    # Check if transaction was executed successfully
    local raw_log=$(echo "$result" | grep "raw_log:" | cut -d'"' -f2)
    if [[ "$raw_log" == *"failed to execute message"* ]]; then
        local txhash=$(echo "$result" | grep "txhash:" | awk '{print $2}')
        echo "Transaction execution failed"
        echo "Transaction hash: $txhash"
        echo "Error message: $raw_log"
        echo "Full output:"
        echo "$result"
        exit 1
    fi

    echo "Transaction successful"
}


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

waiting_for_block_height 21

EVM_CHAIN_ID=$(cast chain-id --rpc-url $EVM_RPC_URL)

# set up pell chain by admin account
# ----------------------------------------------------------------------------------------------------

ADMIN_KEY=$(/usr/local/bin/pellcored keys show admin --keyring-backend=test 2>/dev/null || echo "")
if [ -z "$ADMIN_KEY" ] ; then
  echo $PELL_ADMIN_MENEMONIC | /usr/local/bin/pellcored keys add admin --recover --keyring-backend=test
fi

ADMIN_PELL_ADDRESS=$(/usr/local/bin/pellcored keys show admin --keyring-backend=test --output=json | jq -r '.address')

# send pell to the emissions module
EMISSIONS_MODULE_ADDRESS=$(pellcored query auth module-account emissions | awk '/address:/ {print $2}')

pellcored tx bank send $ADMIN_PELL_ADDRESS $EMISSIONS_MODULE_ADDRESS 1000000pell --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

sleep 6

# deploy the system contracts
PELL_DELEGATION_ADDRESS=$(pellcored query pevm system-contract  |grep "delegation_manager_proxy"| awk 'NR==1 {print $2}')

if [ -z "$PELL_DELEGATION_ADDRESS" ]; then
    echo "Delegation address is empty, deploying system contracts..."
    pellcored tx pevm deploy-system-contracts --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result
    sleep 6

    PELL_DELEGATION_ADDRESS=$(pellcored query pevm system-contract  |grep "delegation_manager_proxy"| awk 'NR==1 {print $2}')
else
    echo "Delegation address found: $PELL_DELEGATION_ADDRESS"
fi

sleep 6

# enable gateway send xmsg
PELL_GATEWAY_CONTRACT_ADDRESS=$(pellcored query pevm system-contract  |grep "gateway"| awk 'NR==1 {print $2}')
pellcored tx xmsg add-allowed-xmsg-sender $PELL_GATEWAY_CONTRACT_ADDRESS --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

sleep 6

# enable gas swap send xmsg
GAS_SWAP_ADDRESS=$(pellcored query pevm system-contract  |grep "gas_swap"| awk 'NR==1 {print $2}')
pellcored tx xmsg add-allowed-xmsg-sender $GAS_SWAP_ADDRESS --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

sleep 6

# upsert the chain params. to support the inbound tx from bsc
pellcored tx relayer upsert-chain-params /root/${EVM_CHAIN_NAME}_chain_params.json --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

sleep 6

pellcored tx restaking update-blocks-per-epoch 5 --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

sleep 6

pellcored tx restaking upsert-outbound-state $EVM_CHAIN_ID 1 0 --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

sleep 6

pellcored tx xmsg upsert-crosschain-fee-params /root/local-eth-fee_params.json --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y | check_tx_result

waiting_for_block_height 25

TSS_ADDRESS=$(pellcored query relayer get-tss-address | awk 'NR==1 {print $2}')

# ----------------------------------------------------------------------------------------------------
# evm setup

SUGGESTED_GAS_PRICE=$(cast gas-price --rpc-url $EVM_RPC_URL)

# add the tss address to the bsc tss manager
cast send $EVM_TSS_MANAGER_ADDRESS "addTSS(address)" $TSS_ADDRESS \
    --rpc-url $EVM_RPC_URL  \
    --private-key $EVM_ADMIN_PRIVATE_KEY \
    --gas-price $SUGGESTED_GAS_PRICE \
    --gas-limit 400000

# update the connector address in the bsc delegation manager
cast send $EVM_DELEGATION_MANAGER_ADDRESS "updateConnector(address)" $EVM_CONNECTOR_ADDRESS \
    --rpc-url $EVM_RPC_URL \
    --private-key $EVM_ADMIN_PRIVATE_KEY \
    --gas-price $SUGGESTED_GAS_PRICE \
    --gas-limit 400000

cast send \
    --rpc-url $EVM_RPC_URL \
    --private-key $EVM_ADMIN_PRIVATE_KEY \
    --value 1000000000000000000 \
    --gas-price $SUGGESTED_GAS_PRICE \
    --gas-limit 400000 \
    $TSS_ADDRESS \
    "0x"

cast send $EVM_GATEWAY_CONTRACT_ADDRESS "updateSourceAddress(uint256,bytes)" $PELL_CHAIN_ID_INT $PELL_GATEWAY_CONTRACT_ADDRESS \
    --rpc-url $EVM_RPC_URL \
    --private-key $EVM_ADMIN_PRIVATE_KEY \
    --gas-price $SUGGESTED_GAS_PRICE \
    --gas-limit 400000

cast send $EVM_GATEWAY_CONTRACT_ADDRESS "updateDestinationAddress(uint256,bytes)" $PELL_CHAIN_ID_INT $PELL_GATEWAY_CONTRACT_ADDRESS \
    --rpc-url $EVM_RPC_URL \
    --private-key $EVM_ADMIN_PRIVATE_KEY \
    --gas-price $SUGGESTED_GAS_PRICE \
    --gas-limit 400000

# print current time
date




