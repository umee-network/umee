#!/bin/sh

set -ex

# initialize price-feeder configuration
mkdir -p /root/.price-feeder/
touch /root/.price-feeder/config.toml

# setup price-feeder configuration
tee /root/.price-feeder/config.toml <<EOF
gas_adjustment = 1.5

[server]
listen_addr = "0.0.0.0:7171"
read_timeout = "20s"
verbose_cors = true
write_timeout = "20s"

[[currency_pairs]]
base = "UMEE"
providers = [
  "mock",
]
quote = "USDT"

[[currency_pairs]]
base = "ATOM"
providers = [
  "mock",
]
quote = "USDC"

[account]
address = '$UMEE_E2E_PRICE_FEEDER_ADDRESS'
chain_id = "umee-local-testnet"
validator = '$UMEE_E2E_PRICE_FEEDER_VALIDATOR'

[keyring]
backend = "test"
dir = '$UMEE_E2E_UMEE_VAL_KEY_DIR'

[rpc]
grpc_endpoint = 'tcp://$UMEE_E2E_UMEE_VAL_HOST:9090'
rpc_timeout = "100ms"
tmrpc_endpoint = 'http://$UMEE_E2E_UMEE_VAL_HOST:26657'

[telemetry]
service_name = "price-feeder"
enabled = true
enable_hostname = true
enable_hostname_label = true
enable_service_label = true
type = "prometheus"
global_labels = [["chain-id", "umee-local-testnet"]]
EOF

# start price-feeder
price-feeder /root/.price-feeder/config.toml
