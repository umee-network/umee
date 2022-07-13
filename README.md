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
  - [Private](#private)
- [Install](#install)

## Releases

See [Release procedure](CONTRIBUTING.md#release-procedure) for more information about the release model.

### Release Compatibility Matrix

| Version | Mainnet | Experimental | SDK Version | IBC Version | Peggo Version | Price Feeder Version |
| :-----: | :-----: | :----------: | :---------: | :---------: | :-----------: | :------------------: |
| v0.8.x  |    ✗    |      ✓       |   v0.45.x   |   v2.0.x    |    v0.2.x     |        v0.1.x        |
| v1.x.x  |    ✓    |      ✗       |   v0.45.x   |   v2.0.x    |    v0.2.x     |         N/A          |
| v2.x.x  |    ✗    |      ✓       |   v0.45.x   |   v2.3.x    |    v0.2.x     |        v0.2.x        |

## Active Networks

### Public

- [umee-1](networks/umee-1)
- [umeemania-1](networks/umeemania-1)

### Private

- [umee-betanet-1](networks/umee-betanet-1)
- [umee-betanet-2](networks/umee-betanet-2)

## Install

To install the `umeed` binary:

```shell
$ make install
```
