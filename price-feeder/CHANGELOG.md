<!-- markdownlint-disable MD013 MD024 -->

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

### Features

- [#540](https://github.com/umee-network/umee/pull/536) Use environment vars / standard input for the keyring password instead of the config file.
- [#522](https://github.com/umee-network/umee/pull/522) Add Okx as a provider.
- [#536](https://github.com/umee-network/umee/pull/536) Force a minimum of three providers per asset.
- [#502](https://github.com/umee-network/umee/pull/502) Faulty provider detection: discard prices that are not within 2𝜎 of others.
- [#551](https://github.com/umee-network/umee/pull/551) Update Binance provider to use WebSocket.
- [#569](https://github.com/umee-network/umee/pull/569) Update Huobi provider to use WebSocket.
- [#580](https://github.com/umee-network/umee/pull/580) Update Kraken provider to use WebSocket.

### Bug Fixes

- [#552](https://github.com/umee-network/umee/pull/552) Stop requiring telemetry during config validation.
- [#574](https://github.com/umee-network/umee/pull/574) Stop registering metrics endpoint if telemetry is disabled.
- [#573](https://github.com/umee-network/umee/pull/573) Strengthen CORS settings.

## [v0.1.0](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv0.1.0) - 2022-02-07

### Features

- Initial release!!!
