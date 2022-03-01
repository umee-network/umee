<!-- markdownlint-disable MD041 -->
<!-- markdownlint-disable MD013 -->
![Logo!](assets/umee-logo.png)

[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://img.shields.io/badge/repo%20status-WIP-yellow.svg?style=flat-square)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/umee-network/umee)
[![Go Report Card](https://goreportcard.com/badge/github.com/umee-network/umee?style=flat-square)](https://goreportcard.com/report/github.com/umee-network/umee)
[![Version](https://img.shields.io/github/tag/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/blob/main/LICENSE)
[![Lines Of Code](https://img.shields.io/tokei/lines/github/umee-network/umee?style=flat-square)](https://github.com/umee-network/umee)
[![GitHub Super-Linter](https://img.shields.io/github/workflow/status/umee-network/umee/Lint?style=flat-square&label=Lint)](https://github.com/marketplace/actions/super-linter)

> A Golang implementation of the Umee network, a decentralized universal capital
facility in the Cosmos ecosystem.

Umee is a Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain. The platform specializes in
allowing staked assets from PoS blockchains to be used as collateral for borrowing
across blockchains. The platform uses a combination of algorithmically determined
interest rates based on market driven conditions. As a cross chain DeFi protocol,
Umee will allow a multitude of decentralized debt products.

## Table of Contents

- [Compatibility Matrix](#compatibility-matrix)
- [Active Networks](#active-networks)
  - [Public](#public)
  - [Private](#private)
- [Install](#install)

## Compatibility Matrix

| Version | Mainnet | Experimental | SDK Version | Peggo Version | Price Feeder Version |
|:-------:|:-------:|:------------:|:-----------:|---------------|----------------------|
|  v0.8.x |    ✗    |      ✓       |   v0.45.x   | v0.2.x        | v0.1.x               |
|  v1.x.x |    ✓    |      ✗       |   v0.45.x   | v0.2.x        | N/A                  |

## Active Networks

### Mainnet

[umee-1](networks/umee-1)

### Private

- [umee-alpha-mainnet-2](networks/umee-alpha-mainnet-2)

## Install

To install the `umeed` binary:

```shell
$ make install
```
