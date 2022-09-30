<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->

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

## [v3.0.2](https://github.com/umee-network/umee/releases/tag/v3.0.2) - 2022-09-29

### Fixes

- [1460](https://github.com/umee-network/umee/pull/1460) Bump Gravity Bridge.

## [v3.0.1](https://github.com/umee-network/umee/releases/tag/v3.0.1) - 2022-09-28

### Fixes

- [1450](https://github.com/umee-network/umee/pull/1450) fix: token registry cache which caused v3.0.0 halt.

## [v3.0.0](https://github.com/umee-network/umee/releases/tag/v3.0.0) - 2022-09-22

### State Machine Breaking

- [1326](https://github.com/umee-network/umee/pull/1326) Setting protocol controlled min gas price.
- [1401](https://github.com/umee-network/umee/pull/1401) Increased free gas oracle tx limit from 100k to 140k.
- [1411](https://github.com/umee-network/umee/pull/1411) Set min gas price to zero for v3 release

### API Breaking

- [1029](https://github.com/umee-network/umee/pull/1029) Removed MsgSetCollateral(addr,denom,bool), and replaced with MsgAddCollateral(addr,coin) and MsgRemoveCollateral(addr,coin)
- [1023](https://github.com/umee-network/umee/pull/1023) Restrict MsgWithdraw to only uToken input (no base token auto-convert)
- [1106](https://github.com/umee-network/umee/pull/1106) Rename Lend to Supply, including MsgLendAsset, Token EnableLend, docs, and internal functions. Also QueryLoaned similar queries to QuerySupplied.
- [1113](https://github.com/umee-network/umee/pull/1113) Rename Amount field to Asset when sdk.Coin type in Msg proto.
- [1122](https://github.com/umee-network/umee/pull/1122) Rename MsgWithdrawAsset, MsgBorrowAsset, MsgRepayAsset, MsgAddCollateral, and MsgRemoveCollateral to MsgWithdraw, MsgBorrow, MsgRepay, MsgCollateralize, MsgDecollateralize.
- [1123](https://github.com/umee-network/umee/pull/1123) Shorten all leverage and oracle query structs by removing the Request suffix.
- [1125](https://github.com/umee-network/umee/pull/1125) Refactor: remove proto getters in x/leverage and x/oracle proto types.
- [1126](https://github.com/umee-network/umee/pull/1126) Update proto json tag from `APY` to `apy`.
- [1118](https://github.com/umee-network/umee/pull/1118) MsgLiquidate now has reward denom instead of full coin
- [1130](https://github.com/umee-network/umee/pull/1130) Update proto json tag to lower case.
- [1140](https://github.com/umee-network/umee/pull/1140) Rename MarketSize query to TotalSuppliedValue, and TokenMarketSize to TotalSupplied.
- [1188](https://github.com/umee-network/umee/pull/1188) Remove all individual queries which duplicate market_summary fields.
- [1199](https://github.com/umee-network/umee/pull/1199) Move all queries which require address input (e.g. `supplied`, `collateral_value`, `borrow_limit`) into aggregate queries `acccount_summary` or `account_balances`.
- [1236](https://github.com/umee-network/umee/pull/1236) Add more response fields to leverage messages.
- [1222](https://github.com/umee-network/umee/pull/1222) Add leverage parameter DirectLiquidationFee.
- [1333](https://github.com/umee-network/umee/pull/1333) Remove first (addr) argument on all CLI commands, using `--from` flag instead.

### Features

- [1147](https://github.com/umee-network/umee/pull/1147) Add SlashWindow oracle query.
- [913](https://github.com/umee-network/umee/pull/913) Add LendEnabled, BorrowEnabled, and Blacklist to Token struct.
- [913](https://github.com/umee-network/umee/pull/913) Changed update registry gov proposal to add and update tokens, but never delete them.
- [918](https://github.com/umee-network/umee/pull/918) Add MarketSummary query to CLI.
- [1068](https://github.com/umee-network/umee/pull/1068) Add a cache layer for token registry.
- [1096](https://github.com/umee-network/umee/pull/1096) Add `max_collateral_share` to the x/leverage token registry.
- [1094](https://github.com/umee-network/umee/pull/1094) Added TotalCollateral query.
- [1099](https://github.com/umee-network/umee/pull/1099) Added TotalBorrowed query.
- [1157](https://github.com/umee-network/umee/pull/1157) Added `PrintOrErr` util function optimizing the CLI code flow.
- [1118](https://github.com/umee-network/umee/pull/1118) MsgLiquidate rewards base assets instead of requiring an addtional MsgWithdraw
- [1159](https://github.com/umee-network/umee/pull/1159) Add `max_supply_utilization` and `min_collateral_liquidity` to the x/leverage token registry.
- [1188](https://github.com/umee-network/umee/pull/1188) Add `liquidity`, `maximum_borrow`, `maximum_collateral`, `minimum_liquidity`, `available_withdraw`, `available_collateralize`, and `utoken_supply` fields to market summary.
- [1203](https://github.com/umee-network/umee/pull/1203) Add Swagger docs.
- [1212](https://github.com/umee-network/umee/pull/1212) Add `util/checkers` utility package providing common check / validation functions.
- [1217](https://github.com/umee-network/umee/pull/1217) Integrated Cosmos SDK v0.46
  - Adding Cosmos SDK x/group module.
  - Increased Gov `MaxMetadataLen` from 255 to 800 characters.
- [1220](https://github.com/umee-network/umee/pull/1220) Submit oracle prevotes / vote txs via the CLI.
- [1222](https://github.com/umee-network/umee/pull/1222) Liquidation reward_denom can now be either token or uToken.
- [1238](https://github.com/umee-network/umee/pull/1238) Added bad debts query.
- [1323](https://github.com/umee-network/umee/pull/1323) Oracle cli - Add validator address override option.
- [1329](https://github.com/umee-network/umee/pull/1329) Implement MaxCollateralShare.
- [1330](https://github.com/umee-network/umee/pull/1330) Implemented MaxSupplyUtilization.
- [1319](https://github.com/umee-network/umee/pull/1319) Implemented MaxSupply.
- [1331](https://github.com/umee-network/umee/pull/1331) Implemented MinCollateralLiquidity.
- [1343](https://github.com/umee-network/umee/pull/1343) RepayBadDebt and Liquidate automatically clear blacklisted collateral.
- [1379](https://github.com/umee-network/umee/pull/1379) Add `mininumCommissionRate` update to all validators.
- [1395](https://github.com/umee-network/umee/pull/1395) Require compile-time flag to enable liquidation_targets query.

### Improvements

- [935](https://github.com/umee-network/umee/pull/935) Fix protobuf linting
- [940](https://github.com/umee-network/umee/pull/940)(x/leverage) Renamed `Keeper.DeriveBorrowUtilization` to `SupplyUtilization` (see #926)
- [959](https://github.com/umee-network/umee/pull/959) Improve ModuleBalance calculation
- [962](https://github.com/umee-network/umee/pull/962) Streamline AccrueAllInterest
- [967](https://github.com/umee-network/umee/pull/962) Use taylor series of e^x for more accurate interest at high APY.
- [987](https://github.com/umee-network/umee/pull/987) Streamline x/leverage CLI tests
- [1012](https://github.com/umee-network/umee/pull/1012) Improve negative time elapsed error message
- [1236](https://github.com/umee-network/umee/pull/1236) Improve leverage event fields.
- [1294](https://github.com/umee-network/umee/pull/1294) Simplify window progress query math.
- [1300](https://github.com/umee-network/umee/pull/1300) Improve leverage test suite and error specificity.
- [1322](https://github.com/umee-network/umee/pull/1322) Improve complete liquidation threshold and close factor.
- [1332](https://github.com/umee-network/umee/pull/1332) Improve reserve exhaustion event and log message.
- [1362](https://github.com/umee-network/umee/pull/1362) Remove inefficent BorrowAmounts and CollateralAmounts leverage invariants.
- [1363](https://github.com/umee-network/umee/pull/1332) Standardize leverage KVStore access andincrease validation.
- [1385](https://github.com/umee-network/umee/pull/1385) Update v1.1-v3.0 upgrade plan name

### Bug Fixes

- [1018](https://github.com/umee-network/umee/pull/1018) Return nil if negative time elapsed from the last block happens.
- [1156](https://github.com/umee-network/umee/pull/1156) Propagate context correctly.
- [1288](https://github.com/umee-network/umee/pull/1288) Safeguards LastInterestTime against time reversals and unintended interest from hard forks.
- [1357](https://github.com/umee-network/umee/pull/1357) Interptex x/0 collateral liquidity as 100%
- [1383](https://github.com/umee-network/umee/pull/1383) Remove potential panic during FeeAndPriority error case.
- [1405](https://github.com/umee-network/umee/pull/1405) No longer skip MinCollateralLiquidity < 1 validation.

## [v2.0.2](https://github.com/umee-network/umee/releases/tag/v2.0.2) - 2022-05-13

### Features

- [860](https://github.com/umee-network/umee/pull/860) Add IBC upgrade and upgrade-client gov proposals.
- [894](https://github.com/umee-network/umee/pull/894) Add MarketSummary query

### Improvements

- [866](https://github.com/umee-network/umee/pull/866) Make the x/oracle keeper's GetExchangeRateBase method more efficient.

### API Breaking

- [870](https://github.com/umee-network/umee/pull/870) Move proto v1beta1 to v1.
- [903](https://github.com/umee-network/umee/pull/903) (leverage) Renamed `Keeper.CalculateLiquidationLimit` to `CalculateLiquidationThreshold`.

## [v2.0.1](https://github.com/umee-network/umee/releases/tag/v2.0.1) - 2022-04-25

### Features

- [835](https://github.com/umee-network/umee/pull/835) Add miss counter query to oracle cli.

### Bug Fixes

- [829](https://github.com/umee-network/umee/pull/829) Fix `umeed tx leverage liquidate` command args.

### Improvements

- [781](https://github.com/umee-network/umee/pull/781) Oracle module unit test cleanup.
- [782](https://github.com/umee-network/umee/pull/782) Add unit test to `x/oracle/types/denom.go` and `x/oracle/types/keys.go`.
- [786](https://github.com/umee-network/umee/pull/786) Add unit test to `x/oracle/...`.
- [798](https://github.com/umee-network/umee/pull/798) Increase `x/oracle` unit test coverage.
- [803](https://github.com/umee-network/umee/pull/803) Add `cover-html` command to makefile.

## [v2.0.0](https://github.com/umee-network/umee/releases/tag/v2.0.0) - 2022-04-06

### Improvements

- [754](https://github.com/umee-network/umee/pull/754) Update go.mod to use `/v2` import path.
- [723](https://github.com/umee-network/umee/pull/723) Add leverage parameter SmallLiquidationSize, which determines the USD value at which a borrow is considered small enough to be liquidated in a single transaction.
- [711](https://github.com/umee-network/umee/pull/711) Clarify error message for negative elapsed time case.

### Features

- Convexity upgrade!!!

## [v1.1.0](https://github.com/umee-network/umee/releases/tag/v1.1.0) - 2022-09-08

### State Machine Breaking

- [1358](https://github.com/umee-network/umee/pull/1358/files) Disable Gravity Bridge bridge messages.

### Improvements

- [#1355](https://github.com/umee-network/umee/pull/1355) Update tooling to go1.19 and CI to the latest setup (based on v3).

## [v1.0.4](https://github.com/umee-network/umee/releases/tag/v1.0.4) - - 2022-09-08

### Improvements

- [#1353](https://github.com/umee-network/umee/pull/1353) Gravity Bridge update

## [v1.0.3](https://github.com/umee-network/umee/releases/tag/v1.0.3) - 2022-02-17

### State Machine Breaking

- [#560](https://github.com/umee-network/umee/pull/560) Use Gravity Bridge fork that disables slashing completely.

## [v1.0.2](https://github.com/umee-network/umee/releases/tag/v1.0.2) - 2022-02-16

### Features

- [#556](https://github.com/umee-network/umee/pull/556) Refactor the `debug addr` command to convert addresses between any Bech32 HRP.

## [v1.0.1](https://github.com/umee-network/umee/releases/tag/v1.0.1) - 2022-02-07

### Bug Fixes

- [#517](https://github.com/umee-network/umee/pull/517) Fix makefile `build` and `install` targets to support Ledger devices.

## [v1.0.0](https://github.com/umee-network/umee/releases/tag/v1.0.0) - 2022-02-07

### Features

- Initial release!!!
