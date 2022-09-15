#!/bin/bash -eu

# USAGE:
# ./single-gen.sh <iterations> <sleep> <num-blocks> <node-address>

CNT=0
ITER="${1:-50}"
SLEEP="${2:-5}"
NUMBLOCKS="${3:-50}"
RPC="${4:-localhost:26657}"

while [ ${CNT} -lt $ITER ]; do
  curr_block=$(curl -s $RPC/status | jq -r '.result.sync_info.latest_block_height')

  if [ -z "$curr_block" ]; then
    echo "Current block is empty, is the node active?"
  fi

  echo "Current block: ${curr_block} iteration: ${CNT}"

  if [ ! -z ${curr_block} ] && [ ${curr_block} -gt ${NUMBLOCKS} ]; then
    echo "Success: number of blocks reached"
    exit 0
  fi

  ((CNT=CNT+1))
  sleep $SLEEP
done

echo "Failed: timeout reached"
exit 1
