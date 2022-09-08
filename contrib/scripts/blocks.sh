#!/bin/bash -eu

# File with commonly used functions for other scripts

CHAIN_ID="${CHAIN_ID:-888}"
NODE_URL="${NODE_URL:-"tcp://localhost:26657"}"
BLOCK_TIME="${BLOCK_TIME:-6}"
UMEED_BIN="${UMEED_BIN:-umeed}"

cid="--chain-id $CHAIN_ID"
nodeUrlFlag="--node $NODE_URL"

get_block_current_height() {
  QUERY_RESPONSE="$($UMEED_BIN query block $nodeUrlFlag $cid)"
  CURR_BLOCK_HEIGHT=$(echo $QUERY_RESPONSE | jq -r '.block.header.height')
  echo $CURR_BLOCK_HEIGHT
}

wait_until_block() {
  WAIT_BLOCK_HEIGHT=${1:-100}
  CURR_BLOCK_HEIGHT=${2:-1}
  MAX_TRIES=50
  CURRENT_TRY=0

  echo "wait_until_block WAIT_BLOCK_HEIGHT: $WAIT_BLOCK_HEIGHT, CURR_BLOCK_HEIGHT: $CURR_BLOCK_HEIGHT"

  while [ $CURR_BLOCK_HEIGHT -lt $WAIT_BLOCK_HEIGHT ]
  do
    CURR_BLOCK_HEIGHT=$(get_block_current_height)
    echo "Current block height $CURR_BLOCK_HEIGHT, waiting to reach $WAIT_BLOCK_HEIGHT"
    sleep $BLOCK_TIME

    ((CURRENT_TRY=CURRENT_TRY+1))

    if [ $CURRENT_TRY -ge $MAX_TRIES ]; then
      echo "MAX_TRIES: $MAX_TRIES >= than CURRENT_TRY: $CURRENT_TRY exiting"
      exit 1
    fi

  done
}
