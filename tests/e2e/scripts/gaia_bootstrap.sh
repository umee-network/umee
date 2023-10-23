#!/bin/bash

set -ex

chmod +x /usr/local/bin/gaiad
gaiad version
gaiad init val01 --chain-id=$UMEE_E2E_GAIA_CHAIN_ID
echo $UMEE_E2E_GAIA_VAL_MNEMONIC | gaiad keys add val01 --recover --keyring-backend=test
gaiad add-genesis-account $(gaiad keys show val01 -a --keyring-backend=test) 1000000000000stake,1000000000000uatom
gaiad gentx val01 500000000000stake --chain-id=$UMEE_E2E_GAIA_CHAIN_ID --keyring-backend=test
gaiad collect-gentxs
sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' /root/.gaia/config/config.toml
sed -i -e 's/enable = false/enable = true/g' /root/.gaia/config/app.toml
# sed -i -e 's/pruning = "default"/pruning = "nothing"/g' /root/.gaia/config/app.toml
sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/g' /root/.gaia/config/config.toml
# 362880 (3*7*24*60*60 / 2 = 907200) represents a 3 week unbonding period (assuming 2 seconds per block).
gaiad start --x-crisis-skip-assert-invariants --pruning=custom --pruning-keep-recent=907200 --pruning-keep-every=0 --pruning-interval=100