#!/bin/sh

set -ex

# initialize price-feeder configuration
mkdir -p /root/.price-feeder/
touch /root/.price-feeder/config.toml

# setup price-feeder configuration
tee /root/.price-feeder/config.toml <<EOF
gas_adjustment = 1
provider_timeout = "5000ms"

[server]
listen_addr = "0.0.0.0:7171"
read_timeout = "20s"
verbose_cors = true
write_timeout = "20s"

[[deviation_thresholds]]
base = "USDT"
threshold = "1.5"

[[deviation_thresholds]]
base = "UMEE"
threshold = "1.5"

[[deviation_thresholds]]
base = "ATOM"
threshold = "1.5"

[[deviation_thresholds]]
base = "USDC"
threshold = "1.5"

[[deviation_thresholds]]
base = "CRO"
threshold = "1.5"

[[deviation_thresholds]]
base = "DAI"
threshold = "2"

[[deviation_thresholds]]
base = "ETH"
threshold = "2"

[[deviation_thresholds]]
base = "WBTC"
threshold = "1.5"

[[currency_pairs]]
base = "UMEE"
providers = [
  "okx",
  "gate",
  "mexc",
]
quote = "USDT"

[[currency_pairs]]
base = "USDT"
providers = [
  "kraken",
  "coinbase",
  "binanceus",
]
quote = "USD"

[[currency_pairs]]
base = "ATOM"
providers = [
  "okx",
  "bitget",
]
quote = "USDT"

[[currency_pairs]]
base = "ATOM"
providers = [
  "kraken",
]
quote = "USD"

[[currency_pairs]]
base = "USDC"
providers = [
  "okx",
  "bitget",
  "kraken",
]
quote = "USDT"

[[currency_pairs]]
base = "DAI"
providers = [
  "okx",
  "bitget",
  "huobi",
]
quote = "USDT"

[[currency_pairs]]
base = "DAI"
providers = [
  "kraken",
]
quote = "USD"

[[currency_pairs]]
base = "ETH"
providers = [
  "okx",
  "bitget",
]
quote = "USDT"

[[currency_pairs]]
base = "ETH"
providers = [
  "kraken",
]
quote = "USD"

[[currency_pairs]]
base = "WBTC"
providers = [
  "okx",
  "bitget",
  "crypto",
]
quote = "USDT"

[account]
address = '$UMEE_E2E_PRICE_FEEDER_ADDRESS'
chain_id = '$UMEE_E2E_CHAIN_ID'
validator = '$UMEE_E2E_PRICE_FEEDER_VALIDATOR'

[keyring]
backend = "test"
dir = '$UMEE_E2E_UMEE_VAL_KEY_DIR'

[rpc]
grpc_endpoint = 'tcp://$UMEE_E2E_UMEE_VAL_HOST:9090'
rpc_timeout = "100ms"
tmrpc_endpoint = 'http://$UMEE_E2E_UMEE_VAL_HOST:26657'

[telemetry]
service-name = "price-feeder"
enabled = true
enable-hostname = true
enable-hostname-label = true
enable-service-label = true
type = "prometheus"
global-labels = [["chain-id", "umee-local-testnet"]]
EOF

# start price-feeder
price-feeder /root/.price-feeder/config.toml --log-level debug
