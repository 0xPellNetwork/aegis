

waiting_for_block_height() {
  if [ -z "$1" ]; then
    echo "Error: block_height is required"
    return 1
  fi

  local block_height=$1
  while true; do
    PELL_BLOCK_HEIGHT=$(pellcored query block 2>/dev/null | jq -r '.block.header.height')
    echo "Waiting for pell block height to be greater than $block_height, current height: $PELL_BLOCK_HEIGHT"
    if [ -n "$PELL_BLOCK_HEIGHT" ] && [ "$PELL_BLOCK_HEIGHT" != "null" ] && [ "$PELL_BLOCK_HEIGHT" -gt $block_height ]; then
      break
    fi
    sleep 1
  done
}

assert_not_null() {
  if [ -n "$1" ] && [ "$1" == "null" ]; then
    echo "[FAIL]: $1 is null"
    exit 1
  fi
  echo "[PASS]: $1 is not null"
}

assert_equal() {
  if [ "$1" != "$2" ]; then
    echo "[FAIL]: $1 is not equal to $2"
    exit 1
  fi
  echo "[PASS]: $1 is equal to $2"
}

assert_number_gt_zero() {
  if [ $1 -le 0 ]; then
    echo "[FAIL]: $1 is not greater than zero"
    exit 1
  fi
  echo "[PASS]: $1 is greater than zero"
}

assert_true() {
  if [ "$1" != "true" ]; then
    echo "[FAIL]: FALSE"
    exit 1
  fi
  echo "[PASS]: TRUE"
}
