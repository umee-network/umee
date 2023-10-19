#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /root/.hermes/
touch /root/.hermes/config.toml

# setup Hermes relayer configuration
tee /root/.hermes/config.toml <<EOF
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
clear_on_start = false
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 3000

[telemetry]
enabled = false
host = '0.0.0.0'
port = 3001

[[chains]]
id = '$UMEE_E2E_UMEE_CHAIN_ID'
rpc_addr = 'http://$UMEE_E2E_UMEE_VAL_HOST:26657'
grpc_addr = 'http://$UMEE_E2E_UMEE_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$UMEE_E2E_UMEE_VAL_HOST:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'umee'
key_name = 'val01-umee'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
max_gas = 3000000
gas_price = { price = 0.05, denom = 'uumee' }
gas_multiplier = 1.1
max_msg_num = 30
trusted_node = false
max_tx_size = 2097152
max_block_time = '5s'
trusting_period = '14days'
clock_drift = '5s' # to accommodate docker containers
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = '$UMEE_E2E_GAIA_CHAIN_ID'
rpc_addr = 'http://$UMEE_E2E_GAIA_VAL_HOST:26657'
grpc_addr = 'http://$UMEE_E2E_GAIA_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$UMEE_E2E_GAIA_VAL_HOST:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'val01-gaia'
address_type = { derivation = 'cosmos' }
store_prefix = 'ibc'
default_gas = 100000
trusted_node = false
max_gas = 3000000
gas_price = { price = 0.001, denom = 'stake' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
EOF

# import gaia and umee keys
echo ${UMEE_E2E_GAIA_VAL_MNEMONIC} > /root/.hermes/val01-gaia
echo ${UMEE_E2E_UMEE_VAL_MNEMONIC} > /root/.hermes/val01-umee

cat /root/.hermes/val01-umee
cat /root/.hermes/val01-gaia 

hermes keys add --chain ${UMEE_E2E_GAIA_CHAIN_ID} --key-name "val01-gaia" --mnemonic-file /root/.hermes/val01-gaia
hermes keys add --chain ${UMEE_E2E_UMEE_CHAIN_ID} --key-name "val01-umee" --mnemonic-file /root/.hermes/val01-umee


### Configure the clients and connection
echo "Initiating connection handshake..."
hermes create connection --a-chain $UMEE_E2E_UMEE_CHAIN_ID --b-chain $UMEE_E2E_GAIA_CHAIN_ID
sleep 2 
echo "Creating the channels..."
hermes create channel --order unordered --a-chain $UMEE_E2E_UMEE_CHAIN_ID --a-connection connection-0 --a-port transfer --b-port transfer

# start Hermes relayer
hermes start 