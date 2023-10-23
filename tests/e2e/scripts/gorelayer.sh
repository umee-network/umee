rly config init 
rly chains add-dir $HOME/relayer/configs
rly paths add-dir $HOME/relayer/paths
rly keys restore $UMEE_E2E_UMEE_CHAIN_ID ibc "$UMEE_E2E_UMEE_VAL_MNEMONIC"
rly keys restore $UMEE_E2E_GAIA_CHAIN_ID ibc "$UMEE_E2E_GAIA_VAL_MNEMONIC"
rly q balance $UMEE_E2E_UMEE_CHAIN_ID
rly q balance $UMEE_E2E_GAIA_CHAIN_ID 
rly tx link rly-path