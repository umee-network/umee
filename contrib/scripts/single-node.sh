#!/bin/bash -eu

# USAGE:
# ./single-gen.sh <option of full path to umeed>

# Starts an umee chain with only a single node. Best used with an umeed bin
# sitting in the same folder as the script, rather than using the one installed.
# Useful for upgrade testing, where two umeed versions can be placed in the
# folder to test.

# Without submitting any governance proposals, it seems umeed 1 and 2 releases
# can just start and continue off the same state back and forth without failing.
# e.g. run this with umeed1, stop umeed1, then run it with umeed2 to continue.

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

NODE_BIN="${1:-$CWD/../../build/umeed}"

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-umeetest-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
DENOM="${DENOM:-uumee}"
STAKE_DENOM="${STAKE_DENOM:-$DENOM}"
CLEANUP="${CLEANUP:-1}"
LOG_LEVEL="${LOG_LEVEL:-info}"
SCALE_FACTOR="${SCALE_FACTOR:-000000}"
VOTING_PERIOD="${VOTING_PERIOD:-20s}"

# Default 1 account keys + 1 user key with no special grants
VAL0_KEY="val"
VAL0_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"
VAL0_ADDR="umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due"

USER_KEY="user"
USER_MNEMONIC="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"
USER_ADDR="umee1usr9g5a4s2qrwl63sdjtrs2qd4a7huh6cuuhrc"

NEWLINE=$'\n'

hdir="$CHAIN_DIR/$CHAIN_ID"

if ! command -v jq &> /dev/null
then
  echo "⚠️ jq command could not be found!"
  echo "Install it by checking https://stedolan.github.io/jq/download/"
  exit 1
fi

echo "--- Chain ID = $CHAIN_ID"
echo "--- Chain Dir = $CHAIN_DIR"
echo "--- Coin Denom = $DENOM"

killall "$NODE_BIN" &>/dev/null || true

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

  echo "--- Adding addresses..."
  $NODE_BIN $home0 keys show $VAL0_KEY -a $kbt
  $NODE_BIN $home0 keys show $VAL0_KEY -a --bech val $kbt
  $NODE_BIN $home0 keys show $USER_KEY -a $kbt
  $NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $VAL0_KEY -a $kbt) $coins &>/dev/null
  $NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $USER_KEY -a $kbt) $coins_user &>/dev/null


  echo "--- Patching genesis..."
  if [[ "$STAKE_DENOM" == "$DENOM" ]]; then
    jq '.consensus_params["block"]["time_iota_ms"]="5000"
      | .app_state["crisis"]["constant_fee"]["denom"]="'$DENOM'"
      | .app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'$DENOM'"
      | .app_state["mint"]["params"]["mint_denom"]="'$DENOM'"
      | .app_state["staking"]["params"]["bond_denom"]="'$DENOM'"
      | .app_state["gravity"]["params"]["bridge_ethereum_address"]="0x93b5122922F9dCd5458Af42Ba69Bd7baEc546B3c"
      | .app_state["gravity"]["params"]["bridge_chain_id"]="5"
      | .app_state["gravity"]["params"]["bridge_active"]=false
      | .app_state["gravity"]["delegate_keys"]=[{"validator":"umeevaloper1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymuzzdn","orchestrator":"'$VAL0_ADDR'","eth_address":"0xfac5EC50BdfbB803f5cFc9BF0A0C2f52aDE5b6dd"},{"validator":"umeevaloper1qjehhqdnc4mevtsumk6nkhm39nqrqtcy2f5k6k","orchestrator":"umee1qjehhqdnc4mevtsumk6nkhm39nqrqtcy2dnetu","eth_address":"0x02fa1b44e2EF8436e6f35D5F56607769c658c225"},{"validator":"umeevaloper1s824eseh42ndyawx702gwcwjqn43u89dhmqdw8","orchestrator":"umee1s824eseh42ndyawx702gwcwjqn43u89dhl8zld","eth_address":"0xd8f468c1B719cc2d50eB1E3A55cFcb60e23758CD"}]
      | .app_state["gravity"]["gravity_nonces"]["latest_valset_nonce"]="0"
      | .app_state["gravity"]["gravity_nonces"]["last_observed_nonce"]="0"
      | .app_state["gov"]["voting_params"]["voting_period"]="10s"' \
        $n0cfgDir/genesis.json > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json

    #  '.app_state["gravity"]["valset_confirms"]=[{"nonce":1,"orchestrator":"'$VAL0_ADDR'","eth_address":"0xfac5EC50BdfbB803f5cFc9BF0A0C2f52aDE5b6dd","signature":"0x9d45cbaada227c7681edd24c00bccf32f649209721aa5dd9f85f55e799c6046c78c5a0e9e870b96dfbb42e453e1e305072d0c31ffa03d4c72c8ecb328cd511b601"},{"nonce":1,"orchestrator":"umee1qjehhqdnc4mevtsumk6nkhm39nqrqtcy2dnetu","eth_address":"0x02fa1b44e2EF8436e6f35D5F56607769c658c225","signature":"0x72df46d2c1eac7b70b7337a00d2a72d0b275d96a7badf3f66307a7b5c7e743b66f23047d4f95bc0e6ed9dc15d58173900588d4fe0f13b051062af55625fdc44b00"},{"nonce":1,"orchestrator":"umee1s824eseh42ndyawx702gwcwjqn43u89dhl8zld","eth_address":"0xd8f468c1B719cc2d50eB1E3A55cFcb60e23758CD","signature":"0x04dee9ba5d72b9394a3de3c3a1c6e60fd3d63fa5fafecc705228b23488c8006a4b935fc7549f3e9b3b4a278530c2a1716459a947c8bee4322f4febb1f800731301"}]'
  fi

  jq '.app_state["gov"]["voting_params"]["voting_period"]="'$VOTING_PERIOD'"' $n0cfgDir/genesis.json > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json

  echo "--- Creating gentx..."
  $NODE_BIN $home0 gentx-gravity $VAL0_KEY 1000$SCALE_FACTOR$STAKE_DENOM 0xfac5EC50BdfbB803f5cFc9BF0A0C2f52aDE5b6dd $VAL0_ADDR $kbt $cid

  $NODE_BIN $home0 collect-gentxs > /dev/null

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
  perl -i -pe 's|timeout_commit = ".*?"|timeout_commit = "5s"|g' $n0cfg

  echo "--- Modifying app..."
  perl -i -pe 's|minimum-gas-prices = ""|minimum-gas-prices = "0.05uumee"|g' $n0app

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
echo "export H1='--home $hdir'"
echo
echo "Command Line Access:"
echo "  * $NODE_BIN --home $hdir status"

$NODE_BIN $home0 start --api.enable true --grpc.address="0.0.0.0:9090" --grpc-web.enable=false --log_level trace > $log_path 2>&1 &

# Adds 1 sec to create the log and makes it easier to debug it on CI
sleep 1

cat $log_path
