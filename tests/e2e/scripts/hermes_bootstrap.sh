#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /home/hermes/.hermes/
touch /home/hermes/.hermes/config.toml

# setup Hermes relayer configuration
tee /home/hermes/.hermes/config.toml <<EOF
[global]
log_level = 'info'

[mode]
[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = true

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 3031

[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001

[[chains]]
id = '$UMEE_E2E_UMEE_CHAIN_ID'
rpc_addr = 'http://$UMEE_E2E_UMEE_VAL_HOST:26657'
grpc_addr = 'http://$UMEE_E2E_UMEE_VAL_HOST:9090'
websocket_addr = 'ws://$UMEE_E2E_UMEE_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'umee'
key_name = 'val01-umee'
store_prefix = 'ibc'
default_gas = 100000
max_gas = 6000000
gas_price = { price = 0.05, denom = 'uumee' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '1m'
max_block_time = '10s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }


[[chains]]
id = '$UMEE_E2E_GAIA_CHAIN_ID'
rpc_addr = 'http://$UMEE_E2E_GAIA_VAL_HOST:26657'
grpc_addr = 'http://$UMEE_E2E_GAIA_VAL_HOST:9090'
websocket_addr = 'ws://$UMEE_E2E_GAIA_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'val01-gaia'
store_prefix = 'ibc'
default_gas = 100000
max_gas = 6000000
gas_price = { price = 0.001, denom = 'stake' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '1m'
max_block_time = '10s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }
EOF

# import gaia and umee keys
# hermes keys restore ${UMEE_E2E_GAIA_CHAIN_ID} -n "val01-gaia" -m "${UMEE_E2E_GAIA_VAL_MNEMONIC}"
# hermes keys restore ${UMEE_E2E_UMEE_CHAIN_ID} -n "val01-umee" -m "${UMEE_E2E_UMEE_VAL_MNEMONIC}"

### Restore Keys
echo "Restoring keys..."
echo ${UMEE_E2E_GAIA_VAL_MNEMONIC} > /home/hermes/.hermes/mnemonic-file1.txt
echo ${UMEE_E2E_UMEE_VAL_MNEMONIC} > /home/hermes/.hermes/mnemonic-file2.txt
hermes --config /home/hermes/.hermes/config.toml keys add --chain ${UMEE_E2E_GAIA_CHAIN_ID} --mnemonic-file /home/hermes/.hermes/mnemonic-file1.txt
hermes --config /home/hermes/.hermes/config.toml keys add --chain ${UMEE_E2E_UMEE_CHAIN_ID} --mnemonic-file /home/hermes/.hermes/mnemonic-file2.txt

### Configure the clients and connection
# echo "Initiating connection handshake..."
# hermes --config /home/hermes/.hermes/config.toml create connection --a-chain ${UMEE_E2E_UMEE_CHAIN_ID} --b-chain ${UMEE_E2E_GAIA_CHAIN_ID}

# sleep 2 

# echo "Creating the channels..."
# hermes --config /home/hermes/.hermes/config.toml create channel --a-chain ${UMEE_E2E_UMEE_CHAIN_ID} --a-connection connection-0 --a-port transfer --b-port transfer
# start Hermes relayer
echo "Starting the hermes relayer"
hermes --config /home/hermes/.hermes/config.toml start
