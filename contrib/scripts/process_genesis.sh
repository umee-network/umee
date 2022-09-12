#!/bin/bash -eu

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

EXPORTED_GENESIS_UNPROCESSED_PATH="${EXPORTED_GENESIS_UNPROCESSED_PATH:-$CWD/umeemania/umeemania.genesis.json}"
EXPORTED_GENESIS_PROCESSED_PATH="${EXPORTED_GENESIS_PROCESSED_PATH:-$CWD/umeemania/tinkered_genesis.json}"
COSMOS_GENESIS_TINKERER_PATH="${COSMOS_GENESIS_TINKERER_PATH:-$CWD/cosmos-genesis-tinkerer}"
COSMOS_GENESIS_TINKERER_SCRIPT="${COSMOS_GENESIS_TINKERER_SCRIPT:-umeemania-fork.py}"
PYTHON_CLI="${PYTHON_CLI:-python3}"

if [ -d "$COSMOS_GENESIS_TINKERER_PATH" ]; then
  echo "$COSMOS_GENESIS_TINKERER_PATH exists."
else
  git clone --depth 1 --sparse https://github.com/umee-network/cosmos-genesis-tinkerer.git $CWD/cosmos-genesis-tinkerer
  pip install -r $COSMOS_GENESIS_TINKERER_PATH/requirements.txt
fi

$PYTHON_CLI $COSMOS_GENESIS_TINKERER_PATH/$COSMOS_GENESIS_TINKERER_SCRIPT $EXPORTED_GENESIS_UNPROCESSED_PATH $EXPORTED_GENESIS_PROCESSED_PATH
