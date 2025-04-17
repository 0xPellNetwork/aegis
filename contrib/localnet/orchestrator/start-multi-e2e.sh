#!/bin/sh
alias pellcored="pellcored --chain-id 1337 --node http://pellcore0:26657"

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

pelle2e multi --config=./e2e_template.yaml