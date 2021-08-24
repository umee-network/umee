![Logo!](assets/umee-small-logo.png)

[![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://img.shields.io/badge/repo%20status-WIP-yellow.svg?style=flat-square)](https://www.repostatus.org/#wip)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://godoc.org/github.com/umee-network/umee)
[![Go Report Card](https://goreportcard.com/badge/github.com/umee-network/umee?style=flat-square)](https://goreportcard.com/report/github.com/umee-network/umee)
[![Version](https://img.shields.io/github/tag/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/releases/latest)
[![License: Apache-2.0](https://img.shields.io/github/license/umee-network/umee.svg?style=flat-square)](https://github.com/umee-network/umee/blob/main/LICENSE)
[![Lines Of Code](https://img.shields.io/tokei/lines/github/umee-network/umee?style=flat-square)](https://github.com/umee-network/umee)

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
- [Cosmos SDK v0.43.0+](https://github.com/cosmos/cosmos-sdk/releases)
- [Starport](https://docs.starport.network/intro/install.html)

## Active Networks

### Betanet

- Chain ID: `umee-betanet-1`
- Gravity Contract Address: `0xc846512f680a2161D2293dB04cbd6C294c5cFfA7`
- Peers:
  - `a9a84866786013f75138388fbf12cdfc425bd39c@137.184.69.184:26656`
  - `684dd9ce7746041d0453322808cc5b238861e386@137.184.65.210:26656`

#### Deployed Tokens

| Token 	| Display 	|                   Address                  	|
|:-----:	|:-------:	|:------------------------------------------:	|
| uumee 	|   umee  	| 0x29889b8e4785eEEb625848a9Fdc599Fb4569e292 	|

## Install

To install the `umeed` binary:

```shell
$ make install
```
