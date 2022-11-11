<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->

![Logo!](assets/umee-logo.png)

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/umee-network/umee)
[![Go Report Card](https://goreportcard.com/badge/github.com/umee-network/umee?style=flat-square)](https://goreportcard.com/report/github.com/umee-network/umee)
[![Version](https://img.shields.io/github/tag/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/blob/main/LICENSE)
[![Lines Of Code](https://img.shields.io/tokei/lines/github/umee-network/umee?style=flat-square)](https://github.com/umee-network/umee)
[![GitHub Super-Linter](https://img.shields.io/github/workflow/status/umee-network/umee/Lint?style=flat-square&label=Lint)](https://github.com/marketplace/actions/super-linter)

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
- [Active Networks](#active-networks)
  - [Public](#public)
- [Install](#install)
  - [Swagger](#swagger)

## Releases

See [Release procedure](CONTRIBUTING.md#release-procedure) for more information about the release model.

### Release Compatibility Matrix

| Umee Version | Mainnet | Experimental | Cosmos SDK |  IBC   | Peggo   | Price Feeder |       Gravity Bridge       |
| :----------: | :-----: | :----------: | :--------: | :----: | :-----: | :----------: | :------------------------: |
|    v0.8.x    |    ✗    |      ✓       |  v0.45.x   | v2.0.x | v0.2.x  |    v0.1.x    |                            |
|    v1.x.x    |    ✓    |      ✗       |  v0.45.x   | v2.0.x | v0.2.x  |     N/A      | umee/v1 module/v1.4.x-umee |
|    v2.x.x    |    ✗    |      ✓       |  v0.45.x   | v2.3.x | v0.2.x  |    v0.2.x    |   umee/v2 module/v1.4.x    |
|    v3.x.x    |    ✓    |      ✗       |  v0.46.x   | v5.0.x | v1.3.x+ |    v1.0.x    | umee/v3 module/v1.5.x-umee |

## Active Networks

### Public

- [umee-1](networks/umee-1)
- [canon-2](networks/canon-2)

## Install

To install the `umeed` binary:

```shell
$ make install
```

### Swagger

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

- For the usual use of `cosmovisor`, we recomend setting theses env variables

```shell
export DAEMON_NAME=umeed
export DAEMON_HOME={NODE_HOME}
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
export DAEMON_PREUPGRADE_MAX_RETRIES=3
```

- To use `cosmovisor` for starting `umeed` process, instead of calling `umeed start`, use `cosmovisor run start [umeed flags]`
