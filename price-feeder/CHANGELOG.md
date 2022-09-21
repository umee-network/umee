<!-- markdownlint-disable MD013 MD024 -->
<!-- markdown-link-check-disable -->

<!--
Changelog Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github PR referenced in the following format:

* (<tag>) [#<PR-number>](https://github.com/umee-network/umee/pull/<PR-number>) <changelog entry>

Types of changes (Stanzas):

Features: for new features.
Improvements: for changes in existing functionality.
Deprecated: for soon-to-be removed features.
Bug Fixes: for any bug fixes.
Client Breaking: for breaking Protobuf, CLI, gRPC and REST routes used by clients.
API Breaking: for breaking exported Go APIs used by developers.
State Machine Breaking: for any changes that result in a divergent application state.

To release a new version, ensure an appropriate release branch exists. Add a
release version and date to the existing Unreleased section which takes the form
of:

## [<version>](https://github.com/umee-network/umee/releases/tag/<version>) - YYYY-MM-DD

Once the version is tagged and released, a PR should be made against the main
branch to incorporate the new changelog updates.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Bugs

- [1428](https://github.com/umee-network/umee/pull/1428) Update umeed version to an actual tag.

## [v1.0.0](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv1.0.0) - 2022-09-19

### Features

- [1328](https://github.com/umee-network/umee/pull/1328) Add bitget provider.
- [1339](https://github.com/umee-network/umee/pull/1339) Add mexc provider.

### Bugs

- [1338](https://github.com/umee-network/umee/pull/1338) Fix websocket reconnections on remote closures.

## [v0.3.0](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.3.0) - 2022-08-31

### Bugs

- [1084](https://github.com/umee-network/umee/pull/1084) Initializes block height before subscription to fix an error message that appeared on the first few ticks.
- [1244](https://github.com/umee-network/umee/pull/1244) Add verification for quote in conversion rate.
- [1264](https://github.com/umee-network/umee/pull/1264) Convert osmosis candle timestamp from seconds to milliseconds.
- [1262](https://github.com/umee-network/umee/pull/1262) Add verification for quote in tvwap map.
- [1268](https://github.com/umee-network/umee/pull/1268) Don't panic when a provider has only out-of-date candles.
- [1291](https://github.com/umee-network/umee/pull/1291) Set sdk version during build time.

### Improvements

- [#1121](https://github.com/umee-network/umee/pull/1121) Use the cosmos-sdk telemetry package instead of our own.
- [#1032](https://github.com/umee-network/umee/pull/1032) Update the accepted tvwap period from 3 minutes to 5 minutes.
- [#978](https://github.com/umee-network/umee/pull/978) Cleanup the oracle package by moving deviation & conversion logic.
- [#1175](https://github.com/umee-network/umee/pull/1175) Add type ProviderName.
- [#1255](https://github.com/umee-network/umee/pull/1255) Move TickerPrice and CandlePrice to types package
- [#1374](https://github.com/umee-network/umee/pull/1374) Add standard for telemetry metrics.

### Features

- [#1038](https://github.com/umee-network/umee/pull/1038) Adds the option for validators to override API endpoints in our config.
- [#1002](https://github.com/umee-network/umee/pull/1002) Add linting to the price feeder CI.
- [#1170](https://github.com/umee-network/umee/pull/1170) Restrict price feeder quotes to USD, USDT, USDC, ETH, DAI, and BTC.
- [#1175](https://github.com/umee-network/umee/pull/1175) Add ProviderName type to facilitate the reading of maps.
- [#1215](https://github.com/umee-network/umee/pull/1215) Moved ProviderName to Name in provider package.
- [#1274](https://github.com/umee-network/umee/pull/1274) Add option to set config by env variables.
- [#1299](https://github.com/umee-network/umee/pull/1299) Add FTX as a provider.

## [v0.2.5](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.2.5) - 2022-07-28

### Bugs

- [1177](https://github.com/umee-network/umee/pull/1177) Update a deprecated osmosis api endpoint.

### Improvements

- [#1179](https://github.com/umee-network/umee/pull/1179) Improve logs when unable to find prices.

## [v0.2.4](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.2.4) - 2022-07-14

### Features

- [1110](https://github.com/umee-network/umee/pull/1110) Add the ability to detect deviations with multi-quoted prices, ex. using BTC/USD and BTC/ETH at the same time.
- [#998](https://github.com/umee-network/umee/pull/998) Make deviation thresholds configurable for stablecoin support.

## [v0.2.3](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.2.3) - 2022-06-30

### Improvements

- [#1069](https://github.com/umee-network/umee/pull/1069) Subscribe to node event EventNewBlockHeader to have the current chain height.

## [v0.2.2](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.2.2) - 2022-06-27

### Improvements

- [#1050](https://github.com/umee-network/umee/pull/1050) Cache x/oracle params to decrease the number of queries to nodes.

### Features

- [#925](https://github.com/umee-network/umee/pull/925) Require stablecoins to be converted to USD to protect against depegging.

## [v0.2.1](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.2.1) - 2022-04-06

### Improvements

- [#766](https://github.com/umee-network/umee/pull/766) Update deps to use umee v2.0.0.

## [v0.2.0](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.2.0) - 2022-04-04

### Features

- [#730](https://github.com/umee-network/umee/pull/730) Update the mock provider to use a new spreadsheet which uses randomness.

### Improvements

- [#684](https://github.com/umee-network/umee/pull/684) Log errors when providers are unable to unmarshal candles and tickers, instead of either one.
- [#732](https://github.com/umee-network/umee/pull/732) Set oracle functions to public to facilitate usage in other repositories.

### Bugs

- [#732](https://github.com/umee-network/umee/pull/732) Fixes an issue where filtering out erroneous providers' candles wasn't working.

## [v0.1.4](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.1.4) - 2022-03-24

### Features

- [#648](https://github.com/umee-network/umee/pull/648) Add Coinbase as a provider.
- [#679](https://github.com/umee-network/umee/pull/679) Add a configurable provider timeout, which defaults to 100ms.

### Bug Fixes

- [#675](https://github.com/umee-network/umee/pull/675) Add necessary input validation to SubscribePairs in the price feeder.

## [v0.1.3](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.1.3) - 2022-03-21

### Features

- [#649](https://github.com/umee-network/umee/pull/649) Add "GetAvailablePairs" function to providers.

## [v0.1.2](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.1.2) - 2022-03-08

### Features

- [#592](https://github.com/umee-network/umee/pull/592) Add subscribe ticker function to the following providers: Binance, Huobi, Kraken, and Okx.
- [#601](https://github.com/umee-network/umee/pull/601) Use TVWAP formula for determining prices when available.
- [#609](https://github.com/umee-network/umee/pull/609) TVWAP faulty provider detection.

### Bug Fixes

- [#607](https://github.com/umee-network/umee/pull/607) Fix kraken provider timestamp unit.

### Refactor

- [#610](https://github.com/umee-network/umee/pull/610) Split subscription of ticker and candle channels.

## [v0.1.1](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.1.1) - 2022-03-01

### Features

- [#502](https://github.com/umee-network/umee/pull/502) Faulty provider detection: discard prices that are not within 2𝜎 of others.
- [#536](https://github.com/umee-network/umee/pull/536) Force a minimum of three providers per asset.
- [#522](https://github.com/umee-network/umee/pull/522) Add Okx as a provider.
- [#551](https://github.com/umee-network/umee/pull/551) Update Binance provider to use WebSocket.
- [#569](https://github.com/umee-network/umee/pull/569) Update Huobi provider to use WebSocket.
- [#540](https://github.com/umee-network/umee/pull/536) Use environment vars / standard input for the keyring password instead of the config file.
- [#580](https://github.com/umee-network/umee/pull/580) Update Kraken provider to use WebSocket.

### Bug Fixes

- [#552](https://github.com/umee-network/umee/pull/552) Stop requiring telemetry during config validation.
- [#573](https://github.com/umee-network/umee/pull/573) Strengthen CORS settings.
- [#574](https://github.com/umee-network/umee/pull/574) Stop registering metrics endpoint if telemetry is disabled.

### Refactor

- [#587](https://github.com/umee-network/umee/pull/587) Clean up logs from price feeder providers.

## [v0.1.0](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.1.0) - 2022-02-07

### Features

- Initial release!!!
