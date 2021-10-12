#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /home/hermes/.hermes/
touch /home/hermes/.hermes/config.toml

# setup Hermes relayer configuration
tee /home/hermes/.hermes/config.toml <<EOF

[global]
strategy = 'packets'
filter = false
log_level = 'info'
clear_packets_interval = 100
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
max_gas = 3000000
gas_price = { price = 0.001, denom = 'uumee' }
gas_adjustment = 0.1
clock_drift = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = '$UMEE_E2E_GAIA_CHAIN_ID'
rpc_addr = 'http://$UMEE_E2E_GAIA_VAL_HOST:26657'
grpc_addr = 'http://$UMEE_E2E_GAIA_VAL_HOST:9090'
websocket_addr = 'ws://$UMEE_E2E_GAIA_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'val01-gaia'
store_prefix = 'ibc'
max_gas = 3000000
gas_price = { price = 0.001, denom = 'stake' }
gas_adjustment = 0.1
clock_drift = '5s'
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
EOF

# import gaia and umee keys
hermes keys restore ${UMEE_E2E_GAIA_CHAIN_ID} -n "val01-gaia" -m "${UMEE_E2E_GAIA_VAL_MNEMONIC}"
hermes keys restore ${UMEE_E2E_UMEE_CHAIN_ID} -n "val01-umee" -m "${UMEE_E2E_UMEE_VAL_MNEMONIC}"

# start Hermes relayer
hermes start
