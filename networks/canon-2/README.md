# Canon-2

Canon-2 will be a long-running testnet.

It was created by stepping through the same upgrade process as
mainnet (1.0 -> 1.1 -> 3.0) and will soon test the (3.0 -> 3.1) upgrade.

## Validators

Before the (3.0 -> 3.1) upgrade, run:

- Umeed v3.0.2
- Peggo v1.2.1 with --eth-merge-pause=true
- Price feeder v1.0.0

After the (3.0 -> 3.1) upgrade, run:

- Umeed v3.1.0-rc2
- Peggo v1.2.1 with --eth-merge-pause=false
  - or do not set flag (default is false)
- Price feeder v1.0.0

## Our validator + archive nodes

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

## To state sync

- init with umeed 3.0.2
- modify config.toml
- modify app.toml
- overwrite genesis.json
- start node

Note: Additional steps required to become a validator

## config.toml changes

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

## app.toml changes

```toml
minimum-gas-prices = "0uumee"

[state-sync]
snapshot-interval=1000
```
