#!/busybox/sh

set -ex

# import orchestrator Umee key
gorc --config=/root/gorc/config.toml keys cosmos recover orch-umee-key "$UMEE_E2E_ORCH_MNEMONIC"

# import orchestrator Ethereum key
gorc --config=/root/gorc/config.toml keys eth import orch-eth-key $UMEE_E2E_ETH_PRIV_KEY

# start gorc orchestrator
gorc --config=/root/gorc/config.toml orchestrator start --cosmos-key=orch-umee-key --ethereum-key=orch-eth-key
