#!/bin/sh


## 1. Setup the admin
ADMIN_MENEMONIC="piece struggle ripple use immense kind royal grass okay mobile client head snake follow mass again candy siege announce immense hard cool unable arrange"
ADMIN_PELL_ADDRESS=pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd
ADMIN_ACCOUNT="3EE2C833278B2CF9513A1918C4232E144D7AB3F9C92C655D4D968A760B944586"

ADMIN_KEY=$(pellcored keys show admin --keyring-backend=test 2>/dev/null || echo "")
if [ -z "$ADMIN_KEY" ] ; then
  echo $ADMIN_MENEMONIC | pellcored keys add admin --recover --keyring-backend=test
fi

alias pellcored="pellcored --chain-id ignite_186-1 --node http://pellcore0:26657"


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


waiting_for_block_height 20

pellcored tx relayer update-chain-params ./mantle_testnet.json  --from=$ADMIN_PELL_ADDRESS --fees=0.6pell --gas=60000000 --keyring-backend=test -y


pelle2e multi --config=./e2e_template.yaml