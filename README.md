<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->

![Logo!](assets/umee-logo.png)

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/umee-network/umee)
[![Go Report Card](https://goreportcard.com/badge/github.com/umee-network/umee?style=flat-square)](https://goreportcard.com/report/github.com/umee-network/umee)
[![Version](https://img.shields.io/github/tag/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/blob/main/LICENSE)
[![GitHub Super-Linter](https://img.shields.io/github/actions/workflow/status/umee-network/umee/lint.yml?branch=main)](https://github.com/marketplace/actions/super-linter)

> A Golang implementation of the Umee network, a decentralized universal capital
> facility in the Cosmos ecosystem.

Umee is a Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain. The platform specializes in
allowing staked assets from PoS blockchains to be used as collateral for borrowing
across blockchains. The platform uses a combination of algorithmically determined
interest rates based on market driven conditions. As a cross chain DeFi protocol,
Umee will allow a multitude of decentralized debt products.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Releases](#releases)
  - [Release Compatibility Matrix](#release-compatibility-matrix)
    - [Price Feeder](#price-feeder)
    - [libwasmvm](#libwasmvm)
  - [Active Networks](#active-networks)
- [Build](#build)
  - [Docker build](#docker-build)
  - [Recommended Database Backend](#recommended-database-backend)
  - [Swagger](#swagger)
  - [Cosmovisor](#cosmovisor)
- [Validators](#validators)
- [Liquidators](#liquidators)

## Releases

See [Release procedure](CONTRIBUTING.md#release-procedure) for more information about the release model.

### Release Compatibility Matrix

| Umee Version | Mainnet | Cosmos SDK |  IBC   |  Peggo  |  Price Feeder  |       Gravity Bridge       | libwasmvm |
| :----------: | :-----: | :--------: | :----: | :-----: | :------------: | :------------------------: | :-------: |
|    v0.8.x    |    ✗    |  v0.45.x   | v2.0.x | v0.2.x  |     v0.1.x     |                            |           |
|    v1.x.x    |    ✓    |  v0.45.x   | v2.0.x | v0.2.x  |      N/A       | umee/v1 module/v1.4.x-umee |           |
|    v2.x.x    |    ✗    |  v0.45.x   | v2.3.x | v0.2.x  |     v0.2.x     |   umee/v2 module/v1.4.x    |           |
|   v3.0-1.x   |    ✓    |  v0.46.x   | v5.0.x | v1.3.x+ |     v1.0.x     | umee/v3 module/v1.5.x-umee |           |
|  v3.1.0-cw1  |    ✗    |  v0.46.x   | v5.0.x | v1.3.x+ |     v2.0.x     | umee/v3 module/v1.5.x-umee |           |
|    v3.2.x    |    ✓    |  v0.46.6+  | v5.1.x | v1.3.x+ |     v2.0.x     |   umee/v3 v1.5.3-umee-3    |  v1.1.1   |
|    v3.3.x    |    ✓    |  v0.46.6+  | v5.1.x | v1.3.x+ |     v2.0.2     |   umee/v3 v1.5.3-umee-3    |  v1.1.1   |
|    v4.0.x    |    ✓    |  v0.46.6+  | v5.1.x | v1.3.x+ |     v2.0.3     |   umee/v4 v1.5.3-umee-4    |  v1.1.1   |
|    v4.1.x    |    ✓    |  v0.46.7+  | v5.2.x | v1.3.x+ |     v2.1.0     |   umee/v4 v1.5.3-umee-4    |  v1.1.1   |
|    v4.2.x    |    ✓    | v0.46.10+  | v5.2.x | v1.3.x+ |  umee/v2.1.1   |   umee/v4 v1.5.3-umee-4    |  v1.1.1   |
|    v4.3.x    |    ✓    | v0.46.11+  | v6.1.x | v1.3.x+ |  umee/v2.1.1   |   umee/v4 v1.5.3-umee-6    |  v1.2.1   |
|    v4.4.x    |    ✓    | v0.46.11+  | v6.1.x | v1.3.x+ |  umee/v2.1.4+  |   umee/v4 v1.5.3-umee-6    |  v1.2.3   |
|    v5.0.x    |    ✓    | v0.46.13+  | v6.2.x | v1.3.x+ |  umee/v2.1.4+  |   umee/v4 v1.5.3-umee-8    |  v1.2.4   |
|    v5.1.x    |    ✓    | v0.46.13+  | v6.2.x |   ---   |  umee/v2.1.6+  |   umee/v4 v1.5.3-umee-10   |  v1.2.4   |
|    v5.2.x    |    ✓    | v0.46.13+  | v6.2.x |   ---   |  umee/v2.1.6+  |   umee/v4 v1.5.3-umee-10   |  v1.2.4   |
|    v6.0.x    |    ✓    | v0.46.14+  | v6.2.x |   ---   | umee/v2.1.6-1+ |            ---             |  v1.3.0   |
|    v6.1.x    |    ✓    | v0.46.15+  | v6.2.x |   ---   |  umee/v2.1.7+  |            ---             |  v1.3.0   |
|    v6.2.x    |    ✓    |  v0.47.6+  | v7.2.x |   ---   |  umee/v2.3.0   |            ---             |  v1.5.0   |
|    v6.3.x    |    ✓    |  v0.47.7+  | v7.3.1 |   ---   |  umee/v2.3.0+  |            ---             |  v1.5.0   |
|    v6.4.x    |    x    | v0.47.10+  | v7.3.2 |   ---   |  umee/v2.4.1+  |            ---             |  v1.5.2   |
|    v6.5.x    |    x    | v0.47.10+  | v7.3.2 |   ---   |  umee/v2.4.3+  |            ---             |  v1.5.2   |

#### Price Feeder

Since `Price Feeder v2.1.0` the recommended oracle price feeder has been moved to this [repository](https://github.com/ojo-network/price-feeder/tree/umee) with the version prefix `umee/`.

#### libwasmvm

When you build the binary from source on the server machine you probably don't need any change. Building from source automatically link the `libwasmvm.$(uname -m).so` created as a part of the build process.

However when you download a binary from GitHub, or from another source, make sure you have the required version of `libwasmvm.<cpu_arch>.so` (should be in your lib directory, e.g.: `/usr/local/lib/`). You can get it:

- from your build machine: copy `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v<version>/internal/api/libwasmvm.$(uname -m).so`
- or download from CosmWasm GitHub `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v<version>/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

### Active Networks

Public:

- [umee-1](networks/umee-1) (mainnet)
- canon-4 (testnet)

## Build

To install the `umeed` binary:

```shell
$ make build
```

### Docker build

```bash
docker build -t umee-network/umeed -f contrib/images/umeed.dockerfile .

# start bash
docker run -it --name umeed umee-network/umeed bash

# or start the node if you already have a node directory setup
docker run -it --name umeed umee-network/umeed umeed start
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
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.5.0/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

- To use `cosmovisor` for starting `umeed` process, instead of calling `umeed start`, use `cosmovisor run start [umeed flags]`

## Validators

Please follow [Validator Instructions](./docs/VALIDATOR.md) for setting up a validator node.

## Liquidators

A guide to running liquidations on Umee can be found [here](./x/leverage/LIQUIDATION.md)
