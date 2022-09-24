#!/bin/bash -eu

# Start a single node chain from an exported genesis file,
# process the genesis file if it wasn't already,
# wait till the node starts to produce blocks,
# upgrade this fork with a software-upgrade proposal.

# USAGE: ./umeemainnet_fork.sh
set -e

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

CHAIN_ID="${CHAIN_ID:-umeemain-local-testnet}"
FORK_DIR="${FORK_DIR:-$CWD}"
CHAIN_DIR="${CHAIN_DIR:-$FORK_DIR/node-data}"
LOG_LEVEL="${LOG_LEVEL:-debug}"
BLOCK_TIME="${BLOCK_TIME:-1}"
UPGRADE_TITLE="${UPGRADE_TITLE:-"v1.1-v3.0"}"
UMEEMAINNET_GENESIS_PATH="${UMEEMAINNET_GENESIS_PATH:-$CWD/mainnet_tinkered_genesis.json}"
NODE_PRIV_KEY="${NODE_PRIV_KEY:-$FORK_DIR/priv_validator_key.json}"
SEC_AWAIT_NODE_START="${SEC_AWAIT_NODE_START:-80}"
MAINNET_VERSION="${MAINNET_VERSION:-"v1.1.2"}"

UMEED_BIN_CURRENT="${UMEED_BIN_CURRENT:-$FORK_DIR/../../build/umeed}"
# UMEED_BIN_MAINNET="${UMEED_BIN_MAINNET:-$FORK_DIR/umeed-releases/umeed-v1.1.2-linux-amd64/umeed}"

# Loads another sources ,
# It will download the mainnet binaries 
. $CWD/download-mainnet-umeed.sh

# It will download the mainnet genesis 
UMEEMAINNET_GENESIS_PATH=$UMEEMAINNET_GENESIS_PATH . $CWD/tinker-mainnet-genesis.sh

. $CWD/blocks.sh

nodeHome="$CHAIN_DIR/$CHAIN_ID"
home="--home $nodeHome"

hdir="$CHAIN_DIR/$CHAIN_ID"

# Folders for nodes
nodeDir="$hdir/node"

# Home flag for folder
nodeHome="--home $nodeDir"

# Config directories for nodes
nodeCfgDir="$nodeDir/config"

# Config files for nodes
nodeCfg="$nodeCfgDir/config.toml"

# App config files for nodes
nodeApp="$nodeCfgDir/app.toml"

# Node private key to sign blocks
nodePrivateKeyPath=$nodeDir/config/priv_validator_key.json

# Common flags
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

ADMIN_KEY="admin"
NODE_MNEMONIC="festival dumb luxury boss forum clip scatter moral ribbon language rib unable burden agree burden misery suspect find crucial seat canvas endless bring lobster"
NEWLINE=$'\n'

genesisConfigPath="config/genesis.json"

# Checks for private_key file
if test -f "$NODE_PRIV_KEY"; then
  echo "$NODE_PRIV_KEY exists."
else
  echo "$NODE_PRIV_KEY does not exists, exiting..."
  exit 1
fi

# Checks for the tikered genesis file


echo Remove everything from the $CHAIN_DIR
rm -rf $CHAIN_DIR

echo Start the chain basic config
$UMEED_BIN_MAINNET $nodeHome $cid init $ADMIN_KEY

echo Create node admin wallet
yes "$NODE_MNEMONIC$NEWLINE" | $UMEED_BIN_MAINNET $nodeHome keys add $ADMIN_KEY $kbt --recover
NODE_ADDR="$($UMEED_BIN_MAINNET $nodeHome keys show $ADMIN_KEY -a $kbt)"

cp $NODE_PRIV_KEY $nodePrivateKeyPath

echo Node addr is $NODE_ADDR

echo Replace generated genesis with tinkered genesis
rm $nodeDir/$genesisConfigPath

cp $UMEEMAINNET_GENESIS_PATH $nodeDir/$genesisConfigPath

## Updating the gov proposal voting perioid to 20seconds 
jq '.app_state.gov.voting_params.voting_period = "20s"' $nodeDir/$genesisConfigPath >  $nodeDir/new-genesis.json
## Copy the new updated genesis
cp $nodeDir/new-genesis.json $nodeDir/$genesisConfigPath


perl -i -pe 's|fast_sync = true|fast_sync = false|g' $nodeCfg
perl -i -pe 's|addr_book_strict = true|addr_book_strict = false|g' $nodeCfg
perl -i -pe 's|external_address = ""|external_address = "tcp://127.0.0.1:26657"|g' $nodeCfg
perl -i -pe 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26657"|g' $nodeCfg
perl -i -pe 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $nodeCfg
perl -i -pe 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $nodeCfg
perl -i -pe 's|timeout_commit = ".*?"|timeout_commit = "1s"|g' $nodeCfg
perl -i -pe 's|minimum-gas-prices = ""|minimum-gas-prices = "0.05uumee"|g' $nodeApp

nodeLogPath=$hdir.umeed-main.log
unset UMEE_ENABLE_BETA

pid_path=$nodeDir.pid
UMEE_ENABLE_BETA=false $UMEED_BIN_MAINNET $nodeHome start --grpc.address="0.0.0.0:9090" --x-crisis-skip-assert-invariants --grpc-web.enable=false --log_level $LOG_LEVEL > $nodeLogPath 2>&1 &

# Gets the node pid
echo $! > $pid_path

echo
echo "Logs:"
echo "  * tail -f $nodeLogPath"

echo Wait for the node to load the genesis state and start to produce blocks D:
sleep $SEC_AWAIT_NODE_START

# Any block number to be confirmed
WAIT_UNTIL_HEIGHT=1000

CHAIN_ID=$CHAIN_ID UMEED_BIN=$UMEED_BIN_CURRENT wait_until_block $WAIT_UNTIL_HEIGHT
echo "Finish wait_until_block"

CURRENT_BLOCK_HEIGHT=$(CHAIN_ID=$CHAIN_ID UMEED_BIN=$UMEED_BIN_CURRENT get_block_current_height)

echo "Current Block: $CURRENT_BLOCK_HEIGHT >= $WAIT_UNTIL_HEIGHT"

# we should produce at least 20 blocks with the new version
((WAIT_UNTIL_HEIGHT=CURRENT_BLOCK_HEIGHT+40))

UMEED_V1_PID_FILE=$pid_path CHAIN_DIR=$CHAIN_DIR CHAIN_ID=$CHAIN_ID LOG_LEVEL=$LOG_LEVEL NODE_NAME=node UPGRADE_TITLE=$UPGRADE_TITLE UMEED_BIN_V1=$UMEED_BIN_MAINNET UMEED_BIN_V2=$UMEED_BIN_CURRENT $CWD/upgrade-test-single-node.sh

echo "UPGRADE FINISH, going to wait to produce blocks from upgrade height to $WAIT_UNTIL_HEIGHT"
echo "Sleep for 50s, wait for upgrade binary to produce blocks for sometime"
sleep 50

CHAIN_ID=$CHAIN_ID UMEED_BIN=$UMEED_BIN_CURRENT wait_until_block $WAIT_UNTIL_HEIGHT

echo
echo "👍 Upgrade Process Finish to $UMEED_BIN_CURRENT"
echo

pid_value=$(cat $pid_path)
echo "Kill the process ID '$pid_value'"

kill -s 15 $pid_value

exit 0
