# Validator Notes

This document describes a process of joining a testnet or a mainnet as a validator.

## Umeeversity

Full documentation is hosted at [learning.ux.xyz](https://learning.ux.xyz). However it may not be up to date.

## Getting a Binary

To run a validator you need 2 binaries: `umeed` and `price-feeder`.

### Umeed

You can get a binary by:

1. [Build](./README.md#build) yourself and follow the latest [Release Notes](./RELEASE_NOTES.md).
   If you build the binary on a different OS than your validator OS, then you need to copy `libwasmvm`:

   ```sh
   scp $GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@<version>/internal/api/libwasmvm.$(uname -m).so running_os:/<lib/path>
   ```

   NOTE: use the correct `wasmvm` version, according to the latest [Release Notes](./RELEASE_NOTES.md) or the [compatibility matrix](./README.md#release-compatibility-matrix).

2. Download latest [binary build](https://github.com/umee-network/umee/releases). The build is compatible with the latest Ubuntu LTS x86-64. You MUST also copy the `libwasmvm` (see note in 1. about libwasmvm version):

   ```sh
   wget https://raw.githubusercontent.com/CosmWasm/wasmvm/<version>/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
   ```

3. Use our released docker [umeed container](https://github.com/umee-network/umee/pkgs/container/umeed).

To test if the `libwasm` is correctly linked, run `umeed version`.

### Price Feeder

We are using Ojo Price Feeder. Please follow the [instructions](https://github.com/ojo-network/price-feeder/blob/umee/README.md). Make sure you use the latest release with the `umee/` prefix (eg: `umee/v2.4.0`).

NOTE: for self building and configuration examples you MUST use the [umee branch](https://github.com/ojo-network/price-feeder/tree/umee).

Copy the [`price-feeder.toml`](https://github.com/ojo-network/price-feeder/blob/umee/price-feeder.example.toml).

- For the provider config you can use our latest [umee-provider-config directory](https://github.com/ojo-network/price-feeder/tree/umee/umee-provider-config) as is.

## Running a node

1. Update the `app.toml` , `client.toml` and `config.toml` based on your preference. You MUST set non zero min gas prices in `app.toml`. Query `umeed q ugov min-gas-price` to see the what is the minimum acceptable value:

   ```toml
   # your app.toml file
   minimum-gas-prices = "0.1uumee"
   ```

## Joining the network.

Before joining the mainnet you should join a testnet!

### Testnet

1. Make sure your are able to run `umeed` and price feeder locally.
2. Join the [Discord server](https://discord.gg/4ZJAFvg9). Make sure you are in the Testnet group.
3. Follow the state sync [canon-4 instructions](https://mzonder.notion.site/UMEE-Start-from-STATE-SYNC-canon-4-f485563a089a436d9d1fe98f54af8737).
4. You can use the following peers in your `config.toml`:

   ```toml
   persistent_peers = "ee7d691781717cbd1bf6f965dc45aad19c7af05f@canon-4.network.umee.cc:10000,dfd1d83b668ff2e59dc1d601a4990d1bd95044ba@canon-4.network.umee.cc:10001"
   ```

5. Using discord, ping one of the UX Team members to send you testnet `uumee`.
6. Once your validator is setup (and you did self delegation), ping again UX Team members and send your validator address. We will do a delegation.
7. Make sure your Price Feeder is running correctly. If your [testnet window misses](https://canon.price-feeder.com/) are above 50% then something is wrong. Look for a help on Discord.

### Mainnet

1. Make sure you firstly tested your setup on Testnet.
2. Use one of the community [snapshots & instructions](https://github.com/obajay/StateSync-snapshots/tree/main/Projects/Umee).
3. Buy `uumee` to self delegate.
4. Make sure your Price Feeder is running correctly. If your [mainnet window misses](https://price-feeder.com/) are above 50% then something is wrong. Look for a help on Discord.

We recommend to use [Cosmovisor](./README.md#cosmovisor) for mainnet nodes.
