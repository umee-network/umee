<!-- markdownlint-disable MD041 -->
![Logo!](assets/umee-small-logo.png)

[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://img.shields.io/badge/repo%20status-WIP-yellow.svg?style=flat-square)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/umee-network/umee)
[![Go Report Card](https://goreportcard.com/badge/github.com/umee-network/umee?style=flat-square)](https://goreportcard.com/report/github.com/umee-network/umee)
[![Version](https://img.shields.io/github/tag/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/blob/main/LICENSE)
[![Lines Of Code](https://img.shields.io/tokei/lines/github/umee-network/umee?style=flat-square)](https://github.com/umee-network/umee)
[![GitHub Super-Linter](https://github.com/umee-network/umee/workflows/Run%20super-linter/badge.svg)](https://github.com/marketplace/actions/super-linter)

> A Golang implementation of the Umee network, a decentralized universal capital
facility in the Cosmos ecosystem.

Umee is a Universal Capital Facility that can collateralize assets on one blockchain
towards borrowing assets on another blockchain. The platform specializes in
allowing staked assets from PoS blockchains to be used as collateral for borrowing
across blockchains. The platform uses a combination of algorithmically determined
interest rates based on market driven conditions. As a cross chain DeFi protocol,
Umee will allow a multitude of decentralized debt products.

## Table of Contents

- [Dependencies](#dependencies)
- [Active Networks](#active-networks)
- [Install](#install)

## Dependencies

- [Go 1.16+](https://golang.org/dl/)
- [Cosmos SDK v0.44.0+](https://github.com/cosmos/cosmos-sdk/releases)
- [Starport](https://docs.starport.network/intro/install.html)

## Active Networks

### Betanet

- Chain ID: `umee-betanet-2`
- Gravity Bridge Orchestrator Version: [`v0.2.10`](https://github.com/PeggyJV/gravity-bridge/releases/tag/v0.2.10)
- Gravity Ethereum Network: `Görli`
- Gravity Contract Address: [`0xc846512f680a2161D2293dB04cbd6C294c5cFfA7`](https://rinkeby.etherscan.io/address/0xc846512f680a2161D2293dB04cbd6C294c5cFfA7)
- Genesis: [genesis.json](https://raw.githubusercontent.com/umee-network/umee/main/networks/umee-betanet-2/genesis.json)
- Genesis Hash: `a0214294429982a0b2772648ae1f45b8dab9ec33d89f3fb1bfd35465a2164fa5`
  - `$ cat genesis.json | jq -S -c -M '' | sha256sum`
- Peers:
  - `a9a84866786013f75138388fbf12cdfc425bd39c@137.184.69.184:26656`
  - `684dd9ce7746041d0453322808cc5b238861e386@137.184.65.210:26656`
  - `c4c425c66d2941ce4d5d98185aa90d2330de5efd@143.244.166.155:26656`
  - `eb42bdbd821fad7bd0048a741237625b4d954d18@143.244.165.138:26656`

#### Deployed Tokens

| Token | Display |                   Address                  |
|:-----:|:-------:|:------------------------------------------:|
| uumee |   umee  | [`0x29889b8e4785eEEb625848a9Fdc599Fb4569e292`](https://rinkeby.etherscan.io/address/0x29889b8e4785eEEb625848a9Fdc599Fb4569e292)|

## Install

To install the `umeed` binary:

```shell
$ make install
```
