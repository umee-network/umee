#!/bin/bash -eux

# Start a single node chain from an exported genesis file,
# process the genesis file if it wasn't already,
# wait till the node starts to produce blocks,
# upgrade this fork with a software-upgrade proposal.

# USAGE: ./umeemainnet_fork.sh

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
UMEED_BIN_MAINNET_URL_TARBALL=${UMEED_BIN_MAINNET_URL_TARBALL:-"https://github.com/umee-network/umee/releases/download/v1.0.3/umeed-v1.0.3-linux-amd64.tar.gz"}
UMEED_BIN_MAINNET=${UMEED_BIN_MAINNET:-"$CWD/umeed-v103"}

# Checks for the umeed v1 file
if [ ! -f "$UMEED_BIN_MAINNET" ]; then
  echo "$UMEED_BIN_MAINNET doesn't exist"

  if [ -z $UMEED_BIN_MAINNET_URL_TARBALL ]; then
    echo You need to set the UMEED_BIN_MAINNET_URL_TARBALL variable
    exit 1
  fi

  UMEED_RELEASES_PATH=$CWD/umeed-releases
  mkdir -p $UMEED_RELEASES_PATH
  wget -c $UMEED_BIN_MAINNET_URL_TARBALL -O - | tar -xz -C $UMEED_RELEASES_PATH

  UMEED_BIN_MAINNET_BASENAME=$(basename $UMEED_BIN_MAINNET_URL_TARBALL .tar.gz)
  UMEED_BIN_MAINNET=$UMEED_RELEASES_PATH/$UMEED_BIN_MAINNET_BASENAME/umeed
fi

CHAIN_ID="${CHAIN_ID:-umeemain-local-testnet}"
FORK_DIR="${FORK_DIR:-$CWD}"
CHAIN_DIR="${CHAIN_DIR:-$FORK_DIR/node-data}"
LOG_LEVEL="${LOG_LEVEL:-debug}"
BLOCK_TIME="${BLOCK_TIME:-6}"
UPGRADE_TITLE="${UPGRADE_TITLE:-"v1.0-v3.0"}"
UMEED_BIN_CURRENT="${UMEED_BIN_CURRENT:-$FORK_DIR/../../build/umeed}"
UMEEMAINNET_GENESIS_PATH="${UMEEMAINNET_GENESIS_PATH:-$CWD/tinkered_genesis.json}"
NODE_PRIV_KEY="${NODE_PRIV_KEY:-$FORK_DIR/priv_validator_key-coping.json}"

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
if [ ! -f "$UMEEMAINNET_GENESIS_PATH" ]; then
  echo "$UMEEMAINNET_GENESIS_PATH doesn't exist"
  EXPORTED_GENESIS_UNPROCESSED=$CWD/umeemainnet.genesis.json

  if [ ! -f "$EXPORTED_GENESIS_UNPROCESSED" ]; then

    EXPORTED_GENESIS_UNZIPED=$CWD/umeemainnet.genesis.json.gz

    if [ ! -f $EXPORTED_GENESIS_UNZIPED ]; then
      echo "$EXPORTED_GENESIS_UNZIPED doesn't exist, we need to curl it"
      curl https://storage.googleapis.com/umeedropzone/jul-28-umee-1-export.json.gz > $EXPORTED_GENESIS_UNZIPED
    fi

    echo "$EXPORTED_GENESIS_UNPROCESSED doesn't exist, we need to unpack"
    gunzip -k $EXPORTED_GENESIS_UNZIPED
  fi

  echo "$EXPORTED_GENESIS_UNPROCESSED exists and ready to be processed"
  EXPORTED_GENESIS_UNPROCESSED_PATH=$EXPORTED_GENESIS_UNPROCESSED COSMOS_GENESIS_TINKERER_SCRIPT=umeemainnet-fork.py EXPORTED_GENESIS_PROCESSED_PATH=$UMEEMAINNET_GENESIS_PATH $CWD/process_genesis.sh
fi

echo Remove everything from the $CHAIN_DIR
rm -rf $CHAIN_DIR

echo Start the chain basic config
$UMEED_BIN_MAINNET $nodeHome $cid init $ADMIN_KEY

echo Create node admin wallet
yes "$NODE_MNEMONIC$NEWLINE" | $UMEED_BIN_MAINNET $nodeHome keys add $ADMIN_KEY $kbt --recover
NODE_ADDR="$($UMEED_BIN_MAINNET $nodeHome keys show $ADMIN_KEY -a $kbt)"

cp $NODE_PRIV_KEY $nodePrivateKeyPath

echo Coping addr is $NODE_ADDR

echo Replace generated genesis with umeemania genesis
rm $nodeDir/$genesisConfigPath

cp $UMEEMAINNET_GENESIS_PATH $nodeDir/$genesisConfigPath

perl -i -pe 's|fast_sync = true|fast_sync = false|g' $nodeCfg
perl -i -pe 's|addr_book_strict = true|addr_book_strict = false|g' $nodeCfg
perl -i -pe 's|external_address = ""|external_address = "tcp://127.0.0.1:26657"|g' $nodeCfg
perl -i -pe 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26657"|g' $nodeCfg
perl -i -pe 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $nodeCfg
perl -i -pe 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $nodeCfg
perl -i -pe 's|timeout_commit = ".*?"|timeout_commit = "5s"|g' $nodeCfg

nodeLogPath=$hdir.umeed-main.log
unset UMEE_ENABLE_BETA

pid_path=$nodeDir.pid
UMEE_ENABLE_BETA=false $UMEED_BIN_MAINNET $nodeHome start --grpc.address="0.0.0.0:9090" --x-crisis-skip-assert-invariants --grpc-web.enable=false --log_level $LOG_LEVEL > $nodeLogPath 2>&1 &

# Gets the node piid
echo $! > $pid_path

echo
echo "Logs:"
echo "  * tail -f $nodeLogPath"

echo Wait for the node to load the genesis state and start to produce blocks ≃ 8min D:
sleep 50

# Any block number to be confirmed
WAIT_UNTIL_HEIGHT=1000
CURRENT_BLOCK_HEIGHT=0
MAX_TRIES=50
CURRENT_TRY=0

while [ $CURRENT_BLOCK_HEIGHT -lt $WAIT_UNTIL_HEIGHT ]
do
  QUERY_RESPONSE="$($UMEED_BIN_MAINNET query block $nodeHome --chain-id $CHAIN_ID)"
  CURRENT_BLOCK_HEIGHT=$(echo $QUERY_RESPONSE | jq -r '.block.header.height')
  echo "Current block height $CURRENT_BLOCK_HEIGHT, waiting to reach $WAIT_UNTIL_HEIGHT"
  ((CURRENT_TRY=CURRENT_TRY+1))

  if [ $CURRENT_TRY -ge $MAX_TRIES ]; then
    exit 1
  fi

  sleep $BLOCK_TIME
done

echo "Current Block: $CURRENT_BLOCK_HEIGHT == $WAIT_UNTIL_HEIGHT"

CURRENT_TRY=0
# we should produce at least 20 blocks with the new version
((WAIT_UNTIL_HEIGHT=CURRENT_BLOCK_HEIGHT+20))

UMEED_V1_PID_FILE=$pid_path CHAIN_DIR=$CHAIN_DIR CHAIN_ID=$CHAIN_ID LOG_LEVEL=$LOG_LEVEL NODE_NAME=node UPGRADE_TITLE=$UPGRADE_TITLE UMEED_BIN_V1=$UMEED_BIN_MAINNET UMEED_BIN_V2=$UMEED_BIN_CURRENT $CWD/upgrade-test-single-node.sh

echo "UPGRADE FINISH, going to wait to produce 20 blocks from: $CURRENT_BLOCK_HEIGHT"
sleep 30

while [ $CURRENT_BLOCK_HEIGHT -lt $WAIT_UNTIL_HEIGHT ]
do
  QUERY_RESPONSE="$($UMEED_BIN_MAINNET query block $nodeHome --chain-id $CHAIN_ID)"
  CURRENT_BLOCK_HEIGHT=$(echo $QUERY_RESPONSE | jq -r '.block.header.height')
  echo "Current block height $CURRENT_BLOCK_HEIGHT, waiting to reach $WAIT_UNTIL_HEIGHT"
  ((CURRENT_TRY=CURRENT_TRY+1))

  if [ $CURRENT_TRY -ge $MAX_TRIES ]; then
    exit 1
  fi

  sleep $BLOCK_TIME
done

echo
echo Upgrade Process Finish
echo

pid_value=$(cat $pid_path)
echo "Kill the process ID '$pid_value'"

kill -s 15 $pid_value

exit 0
