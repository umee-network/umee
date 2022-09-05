#!/bin/bash -eux

# It uses an chain already up and start an governance proposal to upgrade to a new binary version
# vote 'yes' for that proposal, wait to reach to reach an upgrade height and kill the process id
# received by the parameter $UMEED_V1_PID_FILE

# USAGE: UMEED_V1_PID_FILE=$umee_pid ./upgrade-test-single-node.sh

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

UMEED_V1_PID_FILE="${UMEED_V1_PID_FILE_FILE:-""}"

if [ $UMEED_V1_PID_FILE -e "" ]; then
  echo "You need to specify the process id of umeed v1 inside of a file by setting UMEED_V1_PID_FILE"
  exit 1
fi

CHAIN_ID="${CHAIN_ID:-888}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
NODE_NAME="${NODE_NAME:-coping}"
LOG_LEVEL="${LOG_LEVEL:-info}"
NODE_URL="${NODE_URL:-"tcp://localhost:26657"}"
BLOCK_TIME="${BLOCK_TIME:-6}"
UPGRADE_TITLE="${UPGRADE_TITLE:-cosmwasm}"

UMEED_BUILD_PATH="${UMEED_BUILD_PATH:-$CWD/umeed-builds}"
UMEED_BIN_V1="${UMEED_BIN_V1:-$UMEED_BUILD_PATH/umeed-fix-testnet-halt}"
UMEED_BIN_V2="${UMEED_BIN_V2:-$UMEED_BUILD_PATH/umeed-cosmwasm}"

VOTING_PERIOD=${VOTING_PERIOD:-8}

hdir="$CHAIN_DIR/$CHAIN_ID"

# Folders for nodes
nodeDir="$hdir/$NODE_NAME"

# Home flag for folder
nodeHomeFlag="--home $nodeDir"
nodeUrlFlag="--node $NODE_URL"

# Common flags
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

CURRENT_HEIGHT=$($UMEED_BIN_V1 q block $nodeUrlFlag | jq ".block.header.height" -r)
echo blockchain CURRENT_HEIGHT is $CURRENT_HEIGHT

UPGRADE_HEIGHT=$(($CURRENT_HEIGHT + 10))
echo blockchain UPGRADE_HEIGHT is $UPGRADE_HEIGHT

$UMEED_BIN_V1 tx gov submit-proposal software-upgrade $UPGRADE_TITLE --deposit 1000000000uumee \
  --upgrade-height $UPGRADE_HEIGHT --upgrade-info '{"binaries":{"linux/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-linux-amd64","linux/arm64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-linux-arm64","darwin/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc1/gaiad-v6.0.0-rc1-darwin-amd64"}}' \
  -b block $nodeHomeFlag --from admin $nodeUrlFlag $kbt --title yeet --description megayeet $cid --yes

PROPOSAL_ID=$($UMEED_BIN_V1 q gov $nodeUrlFlag proposals -o json | jq ".proposals[-1].proposal_id" -r)
echo proposal ID is $PROPOSAL_ID

$UMEED_BIN_V1 tx gov vote -b async --from admin $nodeUrlFlag $kbt $PROPOSAL_ID yes $nodeHomeFlag $cid --yes

echo "..."
echo "Finish voting in the proposal"
echo "It will wait to reach the block height to upgrade"
echo "..."
BLOCK_HEIGHT=0
while [ $BLOCK_HEIGHT -lt $UPGRADE_HEIGHT ]
do
  QUERY_RESPONSE="$($UMEED_BIN_V1 query block $nodeHomeFlag $nodeUrlFlag $cid)"
  BLOCK_HEIGHT=$(echo $QUERY_RESPONSE | jq -r '.block.header.height')
  echo "Current block height $BLOCK_HEIGHT, waiting to reach $UPGRADE_HEIGHT"
  sleep $BLOCK_TIME
done

echo "Reached upgrade block height: $BLOCK_HEIGHT == $UPGRADE_HEIGHT"
pid_value=$(cat $UMEED_V1_PID_FILE)
echo "Kill the process ID '$pid_value'"

kill -s 15 $pid_value

sleep 5

echo "...."
echo "Upgrade finish"
echo "...."
sleep $VOTING_PERIOD

# Starts a different file for logging
nodeLogPath=$hdir.umeed-v2.log

$UMEED_BIN_V2 $nodeHomeFlag start --minimum-gas-prices=0.001uumee --grpc.address="0.0.0.0:9090"  --grpc-web.enable=false --log_level $LOG_LEVEL > $nodeLogPath 2>&1 &

# Gets the node piid
echo $! > $UMEED_V1_PID_FILE

echo
echo "Logs:"
echo "  * tail -f $nodeLogPath"
