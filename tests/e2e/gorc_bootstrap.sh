#!/busybox/sh

set -ex

# import orchestrator key
gorc --config=/root/gorc/config.toml keys cosmos recover orchestrator "$UMEE_E2E_ORCH_MNEMONIC"

# import Ethereum key
gorc --config=/root/gorc/config.toml keys eth import signer $UMEE_E2E_ETH_PRIV_KEY

# start gorc orchestrator
gorc --config=/root/gorc/config.toml orchestrator start --cosmos-key=orchestrator --ethereum-key=signer
