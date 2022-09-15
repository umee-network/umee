#!/bin/bash -eu

# tinker the mainnet genesis by swapping polychain -> priv_validator_key

# USAGE: ./tinker-mainnet-genesis.sh

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

UMEEMAINNET_GENESIS_PATH="${UMEEMAINNET_GENESIS_PATH:-$CWD/mainnet_tinkered_genesis.json}"
MAINNET_EXPORTED_GENESIS_URL="${MAINNET_EXPORTED_GENESIS_URL:-"https://storage.googleapis.com/umeedropzone/artemis-mainnet-sep-15-exported-genesis.json.gz"}"
EXPORTED_GENESIS_UNPROCESSED="${EXPORTED_GENESIS_UNPROCESSED:-$CWD/umeemainnet.genesis.json}"
EXPORTED_GENESIS_UNZIPED="${EXPORTED_GENESIS_UNZIPED:-$CWD/umeemainnet.genesis.json.gz}"

# Checks for the tinkered genesis file
if [ ! -f "$UMEEMAINNET_GENESIS_PATH" ]; then
  echo "$UMEEMAINNET_GENESIS_PATH doesn't exist"

  if [ ! -f "$EXPORTED_GENESIS_UNPROCESSED" ]; then

    if [ ! -f $EXPORTED_GENESIS_UNZIPED ]; then
      echo "$EXPORTED_GENESIS_UNZIPED doesn't exist, we need to curl it"
      curl $MAINNET_EXPORTED_GENESIS_URL > $EXPORTED_GENESIS_UNZIPED
    fi

    echo "$EXPORTED_GENESIS_UNPROCESSED doesn't exist, we need to unpack"
    gunzip -k $EXPORTED_GENESIS_UNZIPED
  fi

  echo "$EXPORTED_GENESIS_UNPROCESSED exists and ready to be processed"
  EXPORTED_GENESIS_UNPROCESSED_PATH=$EXPORTED_GENESIS_UNPROCESSED COSMOS_GENESIS_TINKERER_SCRIPT=umeemainnet-fork.py EXPORTED_GENESIS_PROCESSED_PATH=$UMEEMAINNET_GENESIS_PATH $CWD/process_genesis.sh
fi
