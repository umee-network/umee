#!/bin/bash
# microtick and bitcanna contributed significantly here.
set -uxe


go mod edit -replace github.com/tendermint/tm-db=github.com/baabeetaa/tm-db@31989c1
go mod tidy

# Set Golang environment variables.
export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Install Umee
# make install

# NOTE: ABOVE YOU CAN USE ALTERNATIVE DATABASES, HERE ARE THE EXACT COMMANDS
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb' -tags boltdb ./...
go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb' -tags pebbledb ./...
# Tendermint team is currently focusing efforts on badgerdb.



# Initialize chain.
#umeed init test
#wget -O ~/.umee/config/genesis.json https://github.com/umee-network/umee/raw/main/networks/umee-1/genesis.json


# Get "trust_hash" and "trust_height".
INTERVAL=1000
LATEST_HEIGHT=$(curl -s https://rpc.apollo.main.network.umee.cc/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "https://rpc.apollo.main.network.umee.cc/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export UMEED_STATESYNC_ENABLE=true
export UMEED_P2P_MAX_NUM_OUTBOUND_PEERS=200
export UMEED_STATESYNC_RPC_SERVERS="https://rpc.blue.main.network.umee.cc:443,https://rpc.blue.main.network.umee.cc:443"
export UMEED_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export UMEED_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
export UMEED_P2P_PERSISTENT_PEERS="08554ecf7c4c33cc809bceefc044c9bd23b933bd@34.146.11.20:26656,6b785fc3a088de3a5e8d222a980936f2187b8c56@34.93.115.217:26656,1d85a200deaefa6ceb20328a0fd83787ce329aa6@34.93.115.217:26656,b3f810438aa53685bba63705f3c29ec122e1e40c@34.127.76.180:26656"

# Start chain.
umeed start --x-crisis-skip-assert-invariants --db_backend pebbledb
