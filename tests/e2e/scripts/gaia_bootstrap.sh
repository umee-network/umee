#!/bin/bash

# TEST BRANCH - DO NOT MERGE

set -ex

gaiad init val01 --chain-id=$UMEE_E2E_GAIA_CHAIN_ID
echo $UMEE_E2E_GAIA_VAL_MNEMONIC | gaiad keys add val01 --recover --keyring-backend=test
gaiad add-genesis-account $(gaiad keys show val01 -a --keyring-backend=test) 1000000000000stake
gaiad gentx val01 500000000000stake --chain-id=$UMEE_E2E_GAIA_CHAIN_ID --keyring-backend=test
gaiad collect-gentxs
sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' /root/.gaia/config/config.toml
gaiad start --x-crisis-skip-assert-invariants
