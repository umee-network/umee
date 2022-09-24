#!/bin/bash -eu

# Using an already running chain, starts a governance proposal to upgrade to a new binary version,
# votes 'yes' on that proposal, waits to reach to reach an upgrade height and kills the process id
# received by the parameter $UMEED_V1_PID_FILE

# USAGE: UMEED_V1_PID_FILE=$umee_pid_file ./upgrade-test-single-node.sh

CWD="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

UMEED_V1_PID_FILE="${UMEED_V1_PID_FILE:-""}"

if [ ! -f $UMEED_V1_PID_FILE ]; then
  echo "You need to specify the file with process id of umeed v1 inside of a file by setting UMEED_V1_PID_FILE"
  exit 1
fi

CHAIN_ID="${CHAIN_ID:-888}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/node-data}"
NODE_NAME="${NODE_NAME:-n0}"
LOG_LEVEL="${LOG_LEVEL:-info}"
NODE_URL="${NODE_URL:-"tcp://localhost:26657"}"
BLOCK_TIME="${BLOCK_TIME:-6}"
UPGRADE_TITLE="${UPGRADE_TITLE:-cosmwasm}"

UMEED_BUILD_PATH="${UMEED_BUILD_PATH:-$CWD/umeed-builds}"
UMEED_BIN_V1="${UMEED_BIN_V1:-$UMEED_BUILD_PATH/umeed-fix-testnet-halt}"
UMEED_BIN_V2="${UMEED_BIN_V2:-$UMEED_BUILD_PATH/umeed-cosmwasm}"

VOTING_PERIOD=${VOTING_PERIOD:-8}

hdir="$CHAIN_DIR/$CHAIN_ID"

# Loads another sources
. $CWD/blocks.sh

# Folders for nodes
nodeDir="$hdir/$NODE_NAME"

# Home flag for folder
nodeHomeFlag="--home $nodeDir"
nodeUrlFlag="--node $NODE_URL"

# Common flags
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

CURRENT_HEIGHT=$(CHAIN_ID=$CHAIN_ID UMEED_BIN=$UMEED_BIN_V1 get_block_current_height)
echo blockchain CURRENT_HEIGHT is $CURRENT_HEIGHT

UPGRADE_HEIGHT=$(($CURRENT_HEIGHT + 30))
echo blockchain UPGRADE_HEIGHT is $UPGRADE_HEIGHT

echo "Submitting the software-upgrade proposal..."
$UMEED_BIN_V1 tx gov submit-proposal software-upgrade $UPGRADE_TITLE --deposit 1000000000uumee \
  --upgrade-height $UPGRADE_HEIGHT \
  -b block $nodeHomeFlag --from admin $nodeUrlFlag $kbt --title yeet --description megayeet $cid --yes --fees 100000uumee

##
PROPOSAL_ID=$($UMEED_BIN_V1 q gov $nodeUrlFlag proposals -o json | jq ".proposals[-1].proposal_id" -r)
echo proposal ID is $PROPOSAL_ID

echo "Voting on proposaal : $PROPOSAL_ID"
$UMEED_BIN_V1 tx gov vote $PROPOSAL_ID yes -b block --from admin $nodeHomeFlag $cid $nodeUrlFlag $kbt  --yes --fees 100000uumee

echo "..."
echo "Finished voting on the proposal"
echo "Waiting to reach the upgrade height"
echo "..."
CHAIN_ID=$CHAIN_ID UMEED_BIN=$UMEED_BIN_V1 wait_until_block $UPGRADE_HEIGHT

CURRENT_HEIGHT=$(CHAIN_ID=$CHAIN_ID UMEED_BIN=$UMEED_BIN_V1 get_block_current_height)

echo "Reached upgrade block height: $CURRENT_HEIGHT == $UPGRADE_HEIGHT"

node_pid_value=$(cat $UMEED_V1_PID_FILE)

echo "Kill the process ID '$node_pid_value'"

kill -s 15 $node_pid_value

sleep 5

echo "...."
echo "Upgrade finished"
echo "...."
sleep $VOTING_PERIOD

# Starts a different file for logging
nodeLogPath=$hdir.umeed-v2.log

$UMEED_BIN_V2 $nodeHomeFlag start --grpc.address="0.0.0.0:9090" --grpc-web.enable=false --log_level $LOG_LEVEL > $nodeLogPath 2>&1 &

# Gets the node pid
echo $! > $UMEED_V1_PID_FILE

echo
echo "Logs:"
echo "  * tail -f $nodeLogPath"
