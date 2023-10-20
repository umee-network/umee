rly config init 

mkdir -p $HOME/rly-configs
mkdir -p $HOME/rly-paths

tee $HOME/rly-configs/$UMEE_E2E_UMEE_CHAIN_ID.json <<EOF
{
  "type": "cosmos",
  "value": {
    "key": "ibc",
    "chain-id": "$UMEE_E2E_UMEE_CHAIN_ID",
    "rpc-addr": "http://$UMEE_E2E_UMEE_VAL_HOST:26657",
    "grpc-addr": "http://$UMEE_E2E_UMEE_VAL_HOST:9090",
    "account-prefix": "umee",
    "keyring-backend": "test",
    "gas-adjustment": 1.5,
    "gas-prices": "0.01uumee",
    "min-gas-amount": 1000000,
    "max-gas-amount": 3000000,
    "debug": true,
    "timeout": "10s",
    "output-format": "json",
    "sign-mode": "direct"
  }
}
EOF

tee $HOME/rly-configs/$UMEE_E2E_GAIA_CHAIN_ID.json <<EOF
{
  "type": "cosmos",
  "value": {
    "key": "ibc",
    "chain-id": "$UMEE_E2E_GAIA_CHAIN_ID",
    "rpc-addr": "http://$UMEE_E2E_GAIA_VAL_HOST:26657",
    "grpc-addr": "http://$UMEE_E2E_GAIA_VAL_HOST:9090",
    "account-prefix": "cosmos",
    "keyring-backend": "test",
    "gas-adjustment": 1.5,
    "gas-prices": "0.01stake",
    "min-gas-amount": 1000000,
    "max-gas-amount": 3000000,
    "debug": true,
    "timeout": "10s",
    "output-format": "json",
    "sign-mode": "direct"
  }
}
EOF

tee $HOME/rly-paths/rly_path.json <<EOF
{
  "src": {
    "chain-id": "$UMEE_E2E_UMEE_CHAIN_ID",
    "port-id": "transfer",
    "order": "unordered",
    "version": "ics20-1"
  },
  "dst": {
    "chain-id": "$UMEE_E2E_GAIA_CHAIN_ID",
    "port-id": "transfer",
    "order": "unordered",
    "version": "ics20-1"
  },
  "strategy": { "type": "naive" }
}
EOF

rly chains add-dir $HOME/rly-configs
rly paths add-dir $HOME/rly-paths
rly keys restore $UMEE_E2E_UMEE_CHAIN_ID ibc "$UMEE_E2E_UMEE_VAL_MNEMONIC"
rly keys restore $UMEE_E2E_GAIA_CHAIN_ID ibc "$UMEE_E2E_GAIA_VAL_MNEMONIC"
rly q balance $UMEE_E2E_UMEE_CHAIN_ID
rly q balance $UMEE_E2E_GAIA_CHAIN_ID 
rly tx link rly_path
rly start