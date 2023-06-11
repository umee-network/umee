#!/bin/bash

# USAGE:
# [CLEANUP=1] ./single-gen.sh <path to umeed>

# Starts an umee chain with only a single node. Best used with an umeed bin
# sitting in the same folder as the script, rather than using the one installed.
# Useful for upgrade testing, where two umeed versions can be placed in the
# folder to test.

# Without submitting any governance proposals, it seems umeed 1 and 2 releases
# can just start and continue off the same state back and forth without failing.
# e.g. run this with umeed1, stop umeed1, then run it with umeed2 to continue.

# TODO:
# - It's not supposed to be that easy, so are we missing something?
# - Try switching versions after the gov proposal, see if it actually requires the new one.
# - What port do we need to look at to see something like this API? https://api.resistability.internal-betanet-1.network.umee.cc/cosmos/bank/v1beta1/balances/umee1zlxdlfu9j98nd5hnxpnf5h7ju7k3kfw6xf5caf
# - Hit those apis, make sure our modules are actually working
# - Use the non-validator user to submit borrow / loan transactions, check functionality
# - See if / where genesis is exported when chain stops after gov proposal.
# - Not noob material, but test upgrade with ibc channels open and relayer up. Does it survive?

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-umeetest-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
DENOM="${DENOM:-uumee}"
STAKE_DENOM="${STAKE_DENOM:-$DENOM}"
CLEANUP="${CLEANUP:-0}"
LOG_LEVEL="${LOG_LEVEL:-info}"
SCALE_FACTOR="${SCALE_FACTOR:-000000}"
VOTING_PERIOD="${VOTING_PERIOD:-15s}"

# Default 1 account keys + 1 user key with no special grants
VAL0_KEY="val"
VAL0_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"
VAL0_ADDR="umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due"

USER_KEY="user"
USER_MNEMONIC="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"
USER_ADDR="umee1usr9g5a4s2qrwl63sdjtrs2qd4a7huh6cuuhrc"

USER2_KEY="user2"
USER2_MNEMONIC="voice damage response make wave access object improve outer hazard all digital violin noise scissors pulse phone circle ski market best diary weapon acid"
USER2_ADDR="umee1x2w9cesnz93n8e7jc9pzy9agc35jw7prg7v970"


NEWLINE=$'\n'

hdir="$CHAIN_DIR/$CHAIN_ID"

if [[ $# -eq 0 ]]; then
  echo "Usage: $0 UMEED_BIN_RELATIVE_PATH"
  exit 1
fi

if ! command -v jq &> /dev/null
then
    echo "⚠️ jq command could not be found!"
    echo "Install it by checking https://stedolan.github.io/jq/download/"
    exit 1
fi

NODE_BIN="$(pwd)/$1"

echo "--- Chain ID = $CHAIN_ID"
echo "--- Chain Dir = $CHAIN_DIR"
echo "--- Coin Denom = $DENOM"

killall "$1" &>/dev/null || true

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
  rm -rf "$CHAIN_DIR"
  echo "Removed $CHAIN_DIR"
fi

# Folder for node
n0dir="$hdir/n0"

# Home flag for folder
home0="--home $n0dir"

# Config directories for node
n0cfgDir="$n0dir/config"

# Config files for nodes
n0cfg="$n0cfgDir/config.toml"

# App config file for node
n0app="$n0cfgDir/app.toml"

# Common flags
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

# Check if the node-data dir has been initialized already
if [[ ! -d "$hdir" ]]; then
  echo "====================================="
  echo "STARTING NEW CHAIN WITH GENESIS STATE"
  echo "====================================="

  echo "--- Creating $NODE_BIN validator with chain-id=$CHAIN_ID"

  # Build genesis file and create accounts
  if [[ "$STAKE_DENOM" != "$DENOM" ]]; then
    coins="1000000$SCALE_FACTOR$STAKE_DENOM,1000000$SCALE_FACTOR$DENOM"
  else
    coins="1000000$SCALE_FACTOR$DENOM"
  fi
  coins_user="1000000$SCALE_FACTOR$DENOM"

  echo "--- Initializing home..."

  # Initialize the home directory of node
  $NODE_BIN $home0 $cid init n0 &>/dev/null

  echo "--- Enabling node API and Swagger"
  sed -i -s '108s/enable = false/enable = true/' $n0app
  sed -i -s '111s/swagger = false/enable = true/' $n0app

  # Generate new random key
  # $NODE_BIN $home0 keys add val $kbt &>/dev/null

  echo "--- Importing keys..."
  echo "$VAL0_MNEMONIC$NEWLINE"
  yes "$VAL0_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $VAL0_KEY $kbt --recover
  yes "$USER_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $USER_KEY $kbt --recover
  yes "$USER2_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $USER2_KEY $kbt --recover

  echo "--- imported keys addresses..."
  echo ">>>>>>>>>>>>"
  echo $NODE_BIN $home0 keys show $VAL0_KEY -a $kbt
  $NODE_BIN $home0 keys show $VAL0_KEY -a $kbt
  $NODE_BIN $home0 keys show $VAL0_KEY -a --bech val $kbt
  $NODE_BIN $home0 keys show $USER_KEY -a $kbt
  $NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $VAL0_KEY -a $kbt) $coins &>/dev/null

  $NODE_BIN $home0 add-genesis-account $USER_ADDR \
$coins_user,"20001${SCALE_FACTOR}SCC","30001${SCALE_FACTOR}MilkyWay","10001${SCALE_FACTOR}FucSEC" \
&>/dev/null


  echo "--- Patching genesis..."
  if [[ "$STAKE_DENOM" == "$DENOM" ]]; then
    jq '.consensus_params["block"]["time_iota_ms"]="5000"
      | .app_state["crisis"]["constant_fee"]["denom"]="'$DENOM'"
      | .app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'$DENOM'"
      | .app_state["mint"]["params"]["mint_denom"]="'$DENOM'"
      | .app_state["staking"]["params"]["bond_denom"]="'$DENOM'"
      | .app_state["gov"]["voting_params"]["voting_period"]="'$VOTING_PERIOD'"' \
        $n0cfgDir/genesis.json > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
  fi

  jq '.app_state["gov"]["voting_params"]["voting_period"]="'$VOTING_PERIOD'"' $n0cfgDir/genesis.json > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json

  python load_hackathon_data.py


  echo "--- Creating gentx..."
  $NODE_BIN $home0 gentx $VAL0_KEY 1000$SCALE_FACTOR$STAKE_DENOM  $kbt $cid


  $NODE_BIN $home0 collect-gentxs 2> /dev/null

  echo "--- Validating genesis..."
  $NODE_BIN $home0 validate-genesis

  # Use perl for cross-platform compatibility
  # Example usage: perl -i -pe 's/^param = ".*?"/param = "100"/' config.toml

  echo "--- Modifying config..."
  perl -i -pe 's|addr_book_strict = true|addr_book_strict = false|g' $n0cfg
  perl -i -pe 's|external_address = ""|external_address = "tcp://127.0.0.1:26657"|g' $n0cfg
  perl -i -pe 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26657"|g' $n0cfg
  perl -i -pe 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $n0cfg
  perl -i -pe 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $n0cfg
  perl -i -pe 's|timeout_commit = ".*?"|timeout_commit = "2s"|g' $n0cfg
  #perl -i -pe 's|db_backend = ".*"|db_backend = "badgerdb"|g' $n0cfg

  echo "--- Modifying app..."
  perl -i -pe 's|minimum-gas-prices = ""|minimum-gas-prices = "0.1uumee"|g' $n0app
  #perl -i -pe 's|app-db-backend = ".*"|app-db-backend = "badgerdb"|g' $n0app

  # Don't need to set peers if just one node, right?
else
  echo "===================================="
  echo "CONTINUING CHAIN FROM PREVIOUS STATE"
  echo "===================================="
fi # data dir check

log_path=$hdir.n0.log

# Start the instance
echo "--- Starting node..."
echo
echo "Logs:"
echo "  * tail -f $log_path"
echo
echo "Env for easy access:"
echo "H1='--home $hdir'"
echo "$cid"
echo
echo "Command Line Access:"
echo "  * $NODE_BIN --home $hdir status"
echo "  * $NODE_BIN --home $hdir $kbt $cid tx "
# ./umeed --home ./node-data/umeetest-1/n0 --keyring-backend test --chain-id umeetest-1 tx


$NODE_BIN $home0 start --api.enable true --grpc.address="0.0.0.0:9090" --grpc-web.enable=false --log_level info > $log_path 2>&1
