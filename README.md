<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->

# ReFiUmee

This is a fork of [Umee](https://github.com/umee-network/umee/) blockchain (a Cosmos native Money Market) for the EthPrague hackathon.

## ReFi Module and changes for EthPrague

Changes:

- new `refileverage` module: new mechanism to collaterize and ReFi assets
- Price Feeder support for ReFi Assets.
- New governance messages.
- Mock the [Axelar](https://axelar.network/) bridge with the General Message Passing interface - light client based cross chain bridge connecting all major EVM chains and IBC chains.
- Go bindings for our [Aave Gho Facilitator](https://github.com/ReFi-DeFi-hack-ethprague-2023/gho-refi-faciliator).

## About Umee

Umee is a Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain. The platform specializes in
allowing staked assets from PoS blockchains to be used as collateral for borrowing
across blockchains. The platform uses a combination of algorithmically determined
interest rates based on market driven conditions. As a cross chain DeFi protocol,
Umee will allow a multitude of decentralized debt products.

## Build

To install the `umeed` binary:

```shell
$ make build
```

### Recommended Database Backend

We recommend to use RocksDB. It requires to install `rocksdb` system libraries.
We plan to migrate newer version of badgerdb, which brings lot of improvements and simplifies the setup.

To build with `rocksdb` enabled:

```bash
ENABLE_ROCKSDB=true COSMOS_BUILD_OPTIONS=rocksdb  make build
```

Once you generate config files, you need to update:

```bash
# app.toml / base configuration options
app-db-backend = "rocksdb"

# config.toml / base configuration options
db_backend = "rocksdb"
```

### Swagger

- To update the latest swagger docs, follow these steps

Generate the latest swagger:

```bash
 $ make proto-swagger-gen
 $ make proto-update-swagger-docs
```

Build the new binary or install the new binary with the latest swagger docs:

```bash
$ make build
# or
$ make install
```

Make sure to execute these commands whenever you want to update the swagger documentation.

- To enable it, modify the node config at `$UMEE_HOME/config/app.toml` to `api.swagger` `true`
- Run the node normally `umeed start`
- Enter the swagger docs `http://localhost:1317/swagger/`

### Cosmovisor

> [Docs](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor)
> Note: `cosmovisor` only works for upgrades in the `umeed`, for off-chain processes updates like `peggo` or `price-feeder`, manual steps are required.

- `cosmovisor` is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals. If it sees a proposal that gets approved, `cosmovisor` can automatically download
  the new binary, stop the current binary, switch from the old binary to the new one, and finally restart the node with the new binary.

- Install it with go

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
```

- Create folders for Cosmovisor

```shell
mkdir -p ~/.umee/cosmovisor/genesis/bin
mkdir -p ~/.umee/cosmovisor/upgrades

cp <path-to-umeed-binary> ~/.umee/cosmovisor/genesis/bin
```

- For the usual use of `cosmovisor`, we recommend setting theses env variables

```shell
export DAEMON_NAME=umeed
export DAEMON_HOME={NODE_HOME}
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
export DAEMON_PREUPGRADE_MAX_RETRIES=3
```

- If you didn't build binary from source in the machine, you have to download the respective `libwasmvm` into your machine.

```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

- To use `cosmovisor` for starting `umeed` process, instead of calling `umeed start`, use `cosmovisor run start [umeed flags]`

## Liquidators

A guide to running liquidations on Umee can be found [here](./x/leverage/LIQUIDATION.md)
