# Canon-2

Canon-2 is a long-running testnet mirroring mainnet functionality. We use it also as the last test before mainnet when doing chain upgrades.

`canon-1` was created by stepping through the same upgrade process as
mainnet (1.0 -> 1.1 -> 3.0) however the v3.0 update introduced consensus bug and we had to start the testnet from scratch.

`canon-2` starts from v3.0.3 and follows mainnet updates (3.0.3 -> 3.1 -> 3.2 ...).

## Syncing

If you start from scratch, start with umeed v3.0.2 and follow the mainnet upgrades outlined in the [Compatibility Matrix](https://github.com/umee-network/umee#release-compatibility-matrix).

### Our validator + archive nodes

```shell
api.ruby.canon-2.network.umee.cc
rpc.ruby.canon-2.network.umee.cc
grpc.ruby.canon-2.network.umee.cc

api.emerald.canon-2.network.umee.cc
rpc.emerald.canon-2.network.umee.cc
grpc.emerald.canon-2.network.umee.cc

api.sapphire.canon-2.network.umee.cc
rpc.sapphire.canon-2.network.umee.cc
grpc.sapphire.canon-2.network.umee.cc
```

### To state sync

- init with umeed 3.0.2
- modify config.toml
- modify app.toml
- overwrite genesis.json
- start node

Note: Additional steps required to become a validator

### config.toml changes

(Trusted height and hash should be replaced with a more recent block when syncing - usually the last multiple of `10000`)

```toml
[mempool]
version = "v1"

[statesync]
enable = true
rpc_servers = "https://rpc.sapphire.canon-2.network.umee.cc:443,https://rpc.emerald.canon-2.network.umee.cc:443"
trust_height = 30000
trust_hash = "67D5E02CCA0508FF464D991C9B2F688A804A6AF821A91461A353C53E90FFD0D3"

[p2p]
seeds = "ab9e8d7227a3199c2832018eec42ade5bf47e71d@35.215.72.45:26656,e89407a37d2ebe0dfa2291c5240abe3a5410995f@35.212.203.22:26656"
```

### app.toml changes

```toml
minimum-gas-prices = "0uumee"

[state-sync]
snapshot-interval=1000
```

## Peggo

Gravity Bridge Smart contract: 0x4d6D7b1dF43C9dE926BeC5F733980Ad7da6D9486
Network: [Goerli](https://goerli.etherscan.io/tx/0x57f296d59d9be9604133fa951f15a1bcc03a2a332972b5761629a9f76d17e36d)
