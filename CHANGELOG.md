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
appropriate stanza (see below). Each entry should ideally include the Github
PR referenced in the following format:

* [PR-number](https://github.com/umee-network/umee/pull/PR-number) description

Types of changes (Stanzas):

State Machine Breaking: for any changes that result in a divergent application state.
Features: for new features.
Improvements: for changes in existing functionality.
Deprecated: for soon-to-be removed features.
Bug Fixes: for any bug fixes.
Client Breaking: for breaking Protobuf, CLI, gRPC and REST routes used by clients.
API Breaking: for breaking exported Go APIs used by developers.

To release a new version, ensure an appropriate release branch exists. Add a
release version and date to the existing Unreleased section which takes the form
of:

## [version](https://github.com/umee-network/umee/releases/tag/version) - YYYY-MM-DD

Once the version is tagged and released, a PR should be made against the main
branch to incorporate the new changelog updates.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## v6.5.0

### Bug Fixes

- [2551](https://github.com/umee-network/umee/pull/2551) Restrict length of IBC transfer memo and receiver fields.

## v6.4.1 - 2024-04-30

### Improvements

- [daef10a](https://github.com/umee-network/umee/commit/daef10ad4f774aae915ad33f4ab1134695146785) Update dependencies.

## v6.4.0 - 2024-03-21

### Features

- [2459](https://github.com/umee-network/umee/pull/2459), [2461](https://github.com/umee-network/umee/pull/2461) uibc: handle `params.ics20_hooks` switch (enabled / disabled).

### Bug Fixes

- [2462](https://github.com/umee-network/umee/pull/2462) (x/leverage) Take `MaxModuleWithdraw` into account when computing user `MaxWithdraw`.

## v6.4.0-beta1 - 2024-03-11

### Features

- [2408](https://github.com/umee-network/umee/pull/2408) New `converter` helper app.
- [2349](https://github.com/umee-network/umee/pull/2349), [2437](https://github.com/umee-network/umee/pull/2437), [2411](https://github.com/umee-network/umee/pull/2411), [2442](https://github.com/umee-network/umee/pull/2442), [2443](https://github.com/umee-network/umee/pull/2443) IBC ICS20 memo handlers.
- [2381](https://github.com/umee-network/umee/pull/2381) Integrated Packet Forward Middleware.
- [2433](https://github.com/umee-network/umee/pull/2433) Noop Axelar GMP handler.

### Improvements

- [2328](https://github.com/umee-network/umee/pull/2328) Add UX and `uux` denom aliases for UMEE.
- [2388](https://github.com/umee-network/umee/pull/2388) Adjust interest rate algorithm and associated token parameter validation rules.

### Bug Fixes

- [2384](https://github.com/umee-network/umee/pull/2384) Fix `sdkclient` acc sequence setting.
- [2417](https://github.com/umee-network/umee/pull/2417) Fix the ibc inflows storing of registered tokens when sender chain is source chain.

## v6.3.0 - 2024-01-03

### Improvements

- [2363](https://github.com/umee-network/umee/pull/2363) Upgrade Cosmos SDK to v0.47.7.
- [2370](https://github.com/umee-network/umee/pull/2370) Add missing params to `uibc/MsgGovUpdateQuota`.
- [2374](https://github.com/umee-network/umee/pull/2374) Add symbol name to the x/uibc QueryAllOutflowsResponse `outflows` entry.
- [2377](https://github.com/umee-network/umee/pull/2377) Add symbol name to the x/uibc QueryInflowsResponse `inflows` entry.

### Features

- [2352](https://github.com/umee-network/umee/pull/2352) new `oracle/MissCounters` query
- [2352](https://github.com/umee-network/umee/pull/2352) new `uibc/Inflows` query.
- [2349](https://github.com/umee-network/umee/pull/2349) UIBC: adding ICS 20 memo handler (disabled).
- [2369](https://github.com/umee-network/umee/pull/2369) Add query `leverage/RegisteredTokensWithMarkets` to fetch Registered Tokens and their Market Summaries for frontend in fewer queries.

### Bug Fixes

- [2358](https://github.com/umee-network/umee/pull/2358) metoken endblocker should be before oracle.
- [2368](https://github.com/umee-network/umee/pull/2368) Fix inflow amount calculation. Previously, the inflow amount of the token was being overridden by the new inflow amount.
- [2375](https://github.com/umee-network/umee/pull/2375) Ensure Umee and SDK account sequence setting changes the calling client.

### API Breaking

- [2375](https://github.com/umee-network/umee/pull/2375) Rename Umee and SDK client `WithAccSeq` to `SetAccSeq`.

## v6.2.0 - 2023-12-01

### Bug Fixes

- [2315](https://github.com/umee-network/umee/pull/2315) Improve reliability of MaxBorrow, MaxWithdraw when special asset pairs present.
- [2346](https://github.com/umee-network/umee/pull/2346) Fix an issue where metokens were not included in historic data.
- [2365](https://github.com/umee-network/umee/pull/2365) Add fee to metoken price and balances query.

### Improvements

- [2299](https://github.com/umee-network/umee/pull/2299) Upgrade Cosmos SDK to v0.47.
- [2301](https://github.com/umee-network/umee/pull/2301) use gov/v1 MinInitialDepositRatio and set it to 0.1.
- [2341](https://github.com/umee-network/umee/pull/2341) inspect query also returns a list of accounts whose positions could not be calculated

### Breaking Changes

- [2332](https://github.com/umee-network/umee/pull/2332) Rename: `TotalOutflowSum` to `OutflowSum`, `TotalInflowSum` to `InflowSum`.

## v6.1.0 - 2023-10-17

### Improvements

- [2285](https://github.com/umee-network/umee/pull/2285) Upgrade Cosmos SDK to v0.46.15.
- [2288](https://github.com/umee-network/umee/pull/2288) Change `consensus.block.max_size` to 6MB.

### Bug Fixes

- [2276](https://github.com/umee-network/umee/pull/2276) e2e test reliability.
- [2278](https://github.com/umee-network/umee/pull/2278) Fix the store upgrade for metoken.

## v6.1.0-beta1 - 2023-09-29

### Features

- [2264](https://github.com/umee-network/umee/pull/2264) Emitting event when IBC Quota is reset.
- [2254](https://github.com/umee-network/umee/pull/2254) Oracle: added timestamp to prices, added `QueryExgRatesWithTimestamp` method.
- [2246](https://github.com/umee-network/umee/pull/2246) Emergency Group: support metoken index updates.

### API-Breaking

- [2267](https://github.com/umee-network/umee/pull/2267) `BorrowLimit` field in QueryAccountSummaryResponse can be nil on missing borrow price (behavior now matches `LiquidationThreshold` field).
- [2254](https://github.com/umee-network/umee/pull/2254) Oracle: `DenomExchangeRate` is renamed to `ExchangeRate`.

### Improvements

- [2256](https://github.com/umee-network/umee/pull/2256) unify and refactor `client` package.
- [2261](https://github.com/umee-network/umee/pull/2261) Use go 1.21
- [2267](https://github.com/umee-network/umee/pull/2267) Leverage transactions accept spot prices up to 3 minutes old, and leverage queries use most recent spot price when required.
- [2263](https://github.com/umee-network/umee/pull/2263) Add spot price fields to account summary.
- [2270](https://github.com/umee-network/umee/pull/2270) Increase free oracle tx limit to 200k gas.
- [2285](https://github.com/umee-network/umee/pull/2285) Leveraged liquidate works during low liquidity.

### Features

- [2269](https://github.com/umee-network/umee/pull/2269) Enable x/metoken module.

### Bug Fixes

- [2260](https://github.com/umee-network/umee/pull/2260) fix(oracle): avg params storage.

## [v6.0.2](https://github.com/umee-network/umee/releases/tag/v6.0.2) - 2023-09-20

### Bug Fixs

- [2257](https://github.com/umee-network/umee/pull/2257) fix(oracle): missing avg params.

## [v6.0.1](https://github.com/umee-network/umee/releases/tag/v6.0.1) - 2023-09-15

### Features

- [2245](https://github.com/umee-network/umee/pull/2245) cli x/ugov: emergency group query.

### Bug Fixes

- [2247](https://github.com/umee-network/umee/pull/2247) `leverage.GovUpdateSpecialAssets`: set missing `LiquidationThreshold` attribute.

## [v6.0.0](https://github.com/umee-network/umee/releases/tag/v6.0.0) - 2023-09-14

### Features

- [2128](https://github.com/umee-network/umee/pull/2128) CW transaction and query handlers for the incentive module.
- [2129](https://github.com/umee-network/umee/pull/2129) Emergency Group x/ugov proto.
- [2145](https://github.com/umee-network/umee/pull/2145) UMEE v2 Inflation.
- [2146](https://github.com/umee-network/umee/pull/2146) Add store `GetTimeMs` and `SetTimeMs`.
- [2157](https://github.com/umee-network/umee/pull/2157) Add `x/metoken` module.
- [2150](https://github.com/umee-network/umee/pull/2150), [2178](https://github.com/umee-network/umee/pull/2178) `x/leverage` Special Asset Pairs.
- [2145](https://github.com/umee-network/umee/pull/2145) Add New `Inflation Parms` to x/ugov proto and added `inflation rate` change logic to umint
- [2159](https://github.com/umee-network/umee/pull/2159) Add hard market cap for token emission.
- [2155](https://github.com/umee-network/umee/pull/2155) `bpmath`: basis points math package.
- [2166](https://github.com/umee-network/umee/pull/2166) Basis Points: `MulDec`
- [2170](https://github.com/umee-network/umee/pull/2170) Add `SupplyFromModule` and `WithdrawToModule` to leverage keeper.
- [2177](https://github.com/umee-network/umee/pull/2177) metoken queries to cosmwasm and stargate queries.
- [2187](https://github.com/umee-network/umee/pull/2187) New CMD: `ibc_denom`. It creates ibc denom by base denom and channel-id.
- [2188](https://github.com/umee-network/umee/pull/2188) Emergency Group support for `x/leverage`.
- [2191](https://github.com/umee-network/umee/pull/2191) Emergency Group support for IBC Quota.
- [2242](https://github.com/umee-network/umee/pull/2242) New `MsgLeveragedLiquidate.MaxRepay` which allows to limit the liquidation size using the leveraged liquidation mechanism.

### Improvements

- [2135](https://github.com/umee-network/umee/pull/2135) Remove Gravity Bridge module.
- [2209](https://github.com/umee-network/umee/pull/2209) Move leverage module params from paramspace to regular leverage module state.

### State Machine Breaking

- [2145](https://github.com/umee-network/umee/pull/2145) new UMEE inflation schedule.
- [2176](https://github.com/umee-network/umee/pull/2176) x/ugov module v2: adding token emission params.

### Bug Fixes

- [2185](https://github.com/umee-network/umee/pull/2185) `x/oracle` end_blocker panic.

### API Breaking

- [2135](https://github.com/umee-network/umee/pull/2135) Gravity Bridge API is removed.
- [2140](https://github.com/umee-network/umee/pull/2140) Renamed ugov EventMinTxFees to EventMinGasPrice.
- [2165](https://github.com/umee-network/umee/pull/2165) Use underscore for message part in the web gRPC path format:
  - `/umee/ugov/v1/min-gas-price` --> `/umee/ugov/v1/min_gas_price`
  - `/umee/ugov/v1/emergency-group` --> `/umee/ugov/v1/emergency_group`
  - `/umee/uibc/v1/all-outflows` --> `/umee/uibc/v1/all_outflows`
- [2169](https://github.com/umee-network/umee/pull/2169) Update numeric store getters (`util/store` package) to return bool if value is missing.
- [2180](https://github.com/umee-network/umee/pull/2180) Rename leverage `Keeper.ExchangeToken -> ToUToken`, `Keeper.ExchangeUToken -> ToToken` and `Keeper.ExchangeUTokens -> ToTokens`.
- [2183](https://github.com/umee-network/umee/pull/2183) Move `ToUTokenDenom`, `StripUTokenDenom` and `HasUTokenPrefix` from `leverage/keeper` to `coin` package.
- [2203](https://github.com/umee-network/umee/pull/2203) Rework proposal messages. Remove Title from `ugov/MsgGovSetIBCStatus`, `ugov/MsgGovUpdateQuota`, `leverage/MsgGovUpdateRegistry`
- [2234](https://github.com/umee-network/umee/pull/2234) Remove "Get" prefix from cli/ ref tests.

## [v5.2.0](https://github.com/umee-network/umee/releases/tag/v5.2.0) - 2023-08-31

### Improvements

- [2134](https://github.com/umee-network/umee/pull/2134) Bump CometBFT to 34.29.
- [2196](https://github.com/umee-network/umee/pull/2196) Adding Amino support to `x/leverage.MsgLeveragedLiquidate`.

### State Machine Breaking

- [2197](https://github.com/umee-network/umee/pull/2197) Allowing duplicate symbols on leverage token registry. Fix the oracle voting miss counter on duplicate symbol denoms.

### Bug Fixes

- [2148](https://github.com/umee-network/umee/pull/2148) Fix MsgBeginUnbonding counting existing unbondings against max unbond twice.
- [2148](https://github.com/umee-network/umee/pull/2148) Fix MsgLeverageLiquidate CLI not actually allowing wildcard denoms.
- [2197](https://github.com/umee-network/umee/pull/2197) Allowing duplicate symbols on leverage token registry. Fix the oracle voting miss counter on duplicate symbol denoms.

### API Breaking

- [2212](https://github.com/umee-network/umee/pull/2212) Fixes an x/oracle RPC endpoint spelling, changing "/umee/oracle/v1/valdiators/{validator_addr}/aggregate_vote" to "/umee/oracle/v1/validators/{validator_addr}/aggregate_vote"

## [v5.1.0](https://github.com/umee-network/umee/releases/tag/v5.1.0) - 2023-07-07

### Bug Fixes

- [2125](https://github.com/umee-network/umee/pull/2125) Fix v5.1 upgrade handler.

## [v5.1.0-rc1](https://github.com/umee-network/umee/releases/tag/v5.1.0-rc1) - 2023-06-30

### Features

- [2121](https://github.com/umee-network/umee/pull/2121) Allow `MsgLeveragedLiquidate` to auto-select repay and reward denoms if request fields left blank.
- [2114](https://github.com/umee-network/umee/pull/2114) Add borrow factor to `x/leverage`
- [2102](https://github.com/umee-network/umee/pull/2102) and [2106](https://github.com/umee-network/umee/pull/2106) Add `MsgLeveragedLiquidate` to `x/leverage`
- [2085](https://github.com/umee-network/umee/pull/2085) Add `inspect` query to leverage module, which msut be enabled on a node by running with `-l` liquidator query flag.
- [1952](https://github.com/umee-network/umee/pull/1952) Add `x/incentive` module.
- [2015](https://github.com/umee-network/umee/pull/2015), [2050](https://github.com/umee-network/umee/pull/2050) Add `x/ugov` module.
- [2078](https://github.com/umee-network/umee/pull/2078) Upgrade `ibc-go` to v6.2.

### Improvements

- [2057](https://github.com/umee-network/umee/pull/2057) Cosmwasm QA tests.

## [v5.0.0](https://github.com/umee-network/umee/releases/tag/v5.0.0) - 2023-06-07

### Improvements

- [2076](https://github.com/umee-network/umee/pull/2076) Cosmwasm: registering `cosmwasm_1_2` capability.
- [2083](https://github.com/umee-network/umee/pull/2083) Update `wasmvm` to 1.2.4.

### Fixes

- [2089](https://github.com/umee-network/umee/pull/2089) MsgSupplyCollateral no longer fails when market is below MinCollateralLiquidity.

## [v5.0.0-rc1](https://github.com/umee-network/umee/releases/tag/v5.0.0-rc1) - 2023-05-31

### Features

- [1888](https://github.com/umee-network/umee/pull/1888) Created `/sdkclient` and `/client` (umee client) packages to easy the E2E tests and external tools. Essentially, you can import that client and broadcast transactions easily.
- [1993](https://github.com/umee-network/umee/pull/1993) Updated our Cosmos SDK fork to 0.46.12 and included an option to disable colored logs.
- [2071](https://github.com/umee-network/umee/pull/2071) Update GB and enable Valset Burn mechanism.
- [1547](https://github.com/umee-network/umee/pull/1547), [1997](https://github.com/umee-network/umee/pull/1997), [2058](https://github.com/umee-network/umee/pull/2058), [2063](https://github.com/umee-network/umee/pull/2063) Cosmwasm integration.

### Improvements

- [1989](https://github.com/umee-network/umee/pull/1989) Leverage: fix the duplicate symbol denoms issue on adding new tokens to registry.
- [1989](https://github.com/umee-network/umee/pull/1989) Updated go version from 1.19 to 1.20
- [2001](https://github.com/umee-network/umee/pull/2001) Refactoring CLI integration test suite.
- [2009](https://github.com/umee-network/umee/pull/2009) Use DavidAnson/markdownlint action for Markdown linting.
- [2010](https://github.com/umee-network/umee/pull/2010) New `util/store` generic functions to load all values from a store.
- [2021](https://github.com/umee-network/umee/pull/2021) `uibc/quota/keeper` unit test refactor.
- [2024](https://github.com/umee-network/umee/pull/2024) New upgrade registration methods.
- [2037](https://github.com/umee-network/umee/pull/2037) Speedup e2e by removing docker cache.
- [2038](https://github.com/umee-network/umee/pull/2038) Unify util tsdk and store packages.

### Fixes

- [1982](https://github.com/umee-network/umee/pull/1982) Fix the build version (`umeed version`).
- [2052](https://github.com/umee-network/umee/pull/2052) Allow liquidation threshold == collateral weight in token validation.
- [2072](https://github.com/umee-network/umee/pull/2072) Fix an int64 overflow when computing module liquidity for high-exponent assets.

## [v4.4.2](https://github.com/umee-network/umee/releases/tag/v4.4.2) - 2023-06-08

- [2090](https://github.com/umee-network/umee/pull/2090) Bump Cosmos SDK to v0.46.13 and CometBFT to v0.34.28 and IAVL to v0.19.6.

## [v4.4.1](https://github.com/umee-network/umee/releases/tag/v4.4.1) - 2023-05-25

### Improvements

- [2064](https://github.com/umee-network/umee/pull/2064) Bump `ibc-go` to v6.1.1.

## [v4.4.0](https://github.com/umee-network/umee/releases/tag/v4.4.0) - 2023-05-05

### State Machine Breaking

- [2022](https://github.com/umee-network/umee/pull/2022) Disable IBC ICS-20 inflow of only x/leverage registered tokens.

## [v4.3.0](https://github.com/umee-network/umee/releases/tag/v4.3.0) - 2023-04-05

### Features

- [1963](https://github.com/umee-network/umee/pull/1963) ICA Host integration.
- [1953](https://github.com/umee-network/umee/pull/1953) IBC: accept only inflow of leverage registered tokens
- [1967](https://github.com/umee-network/umee/pull/1967) Gravity Bridge phase out phase-2: disable Umee -> Ethereum transfers.
- [1967](https://github.com/umee-network/umee/pull/1967) Gravity Bridge phase out phase-2: disable Umee -> Ethereum transfers.

### Improvements

- [1959](https://github.com/umee-network/umee/pull/1959) Update IBC to v6.1
- [1962](https://github.com/umee-network/umee/pull/1962) Increasing unit test coverage for `x/leverage`, `x/oracle`
  and `x/uibc`
- [1913](https://github.com/umee-network/umee/pull/1913), [1974](https://github.com/umee-network/umee/pull/1974) uibc: quota status check.
- [1973](https://github.com/umee-network/umee/pull/1973) UIBC: handle zero Quota Params.

### Fixes

- [1929](https://github.com/umee-network/umee/pull/1929) Leverage: `MaxWithdraw` now accounts for `MinCollateralLiquidity`
- [1957](https://github.com/umee-network/umee/pull/1957) Leverage: Reserved amount per block now rounds up.
- [1956](https://github.com/umee-network/umee/pull/1956) Leverage: token liquidation threshold must be bigger than collateral_weight.
- [1954](https://github.com/umee-network/umee/pull/1954) Leverage: `MaxBorrow` now accounts for
  `MinCollateralLiquidity` and `MaxSupplyUtilization`
- [1968](https://github.com/umee-network/umee/pull/1968) Leverage: fix type cast of AdjustedBorrow in ExportGenesis

## [v4.2.0](https://github.com/umee-network/umee/releases/tag/v4.2.0) - 2023-03-15

### Features

- [1867](https://github.com/umee-network/umee/pull/1867) Allow `/denom` option on registered tokens query to get only a single token by `base_denom`.
- [1568](https://github.com/umee-network/umee/pull/1568) IBC ICS20 transfer quota. New Cosmos SDK module and IBC ICS20 middleware to limit IBC token outflows.
- [1764](https://github.com/umee-network/umee/pull/1764) New `util.Panic` helper function.
- [1725](https://github.com/umee-network/umee/pull/1725) historacle: average prices.

### Improvements

- [1744](https://github.com/umee-network/umee/pull/1744) docs: testing guidelines.
- [1771](https://github.com/umee-network/umee/pull/1771) CI: add experimental e2e tests on docker image.
- [1788](https://github.com/umee-network/umee/pull/1788) deprecated use of `sdkerrors`.
- [1835](https://github.com/umee-network/umee/pull/1835) CI: use experimental for default CI tests.
- [1864](https://github.com/umee-network/umee/pull/1864) testing: mock gen integration.

### Fixes

- [1767](https://github.com/umee-network/umee/pull/1767) Oracle: Fix `GetTickerPrice()` and `GetCandlePrice()`.

## [v4.1.0](https://github.com/umee-network/umee/releases/tag/v4.1.0) - 2023-02-15

### Features

- [1808](https://github.com/umee-network/umee/pull/1808) Blacklisted tokens automatically cleared from token registry if they have not yet been supplied.

### Fixes

- [1707](https://github.com/umee-network/umee/pull/1707) Oracle: Enforce voting threshold param in oracle endblocker.
- [1736](https://github.com/umee-network/umee/pull/1736) Blacklisted tokens no longer add themselves back to the oracle accept list.
- [1807](https://github.com/umee-network/umee/pull/1807) Fixes BNB ibc denom in 4.1 migration
- [1812](https://github.com/umee-network/umee/pull/1812) MaxCollateralShare now works during partial oracle outages when certain conditions are safe.
- [1821](https://github.com/umee-network/umee/pull/1821) Allow safe leverage operations during partial oracle outages.
- [1845](https://github.com/umee-network/umee/pull/1845) Fix validator power calculation during oracle ballot counting.
- [1851](https://github.com/umee-network/umee/pull/1851) Oracle: ballot sorting.
- [1852](https://github.com/umee-network/umee/pull/1852) Oracle: power vote calculation.

## [v4.0.1](https://github.com/umee-network/umee/releases/tag/v4.0.1) - 2023-02-10

### Fixes

- [1800](https://github.com/umee-network/umee/pull/1800) Handle non-capitalized assets when calling the historacle data.

## [v4.0.0](https://github.com/umee-network/umee/releases/tag/v4.0.0) - 2023-01-20

### API Breaking

- [1683](https://github.com/umee-network/umee/pull/1683) MaxWithdraw query now returns `sdk.Coins`, not `sdk.Coin` and will be empty (not zero coin) when returning a zero amount. Denom field in query is now optional.
- [1694](https://github.com/umee-network/umee/pull/1694) `MsgMaxWithdraw`, `MsgMaxBorrow` and `MsgRepay` won't return errors if there is nothing to withdraw, borrow or repay respectively. Leverage `ErrMaxWithdrawZero` and `ErrMaxBorrowZero` has been removed.

### Fixes

- [1680](https://github.com/umee-network/umee/pull/1680) Add amino support for MsgMaxWithdraw.
- [1694](https://github.com/umee-network/umee/pull/1694) `leverage.MaxBorrow` return zero instead of failing when there is no more to borrow.
- [1710](https://github.com/umee-network/umee/pull/1710) Skip blacklisted tokens in MaxBorrow and MaxWithdraw queries.
- [1717](https://github.com/umee-network/umee/pull/1717) Oracle: Add blockNum to median and median deviation queries.

### Features

- [1548](https://github.com/umee-network/umee/pull/1548) Historacle prices and medians keeper proof of concept.
- [1580](https://github.com/umee-network/umee/pull/1580), [1632](https://github.com/umee-network/umee/pull/1632), [1657](https://github.com/umee-network/umee/pull/1657) Median tracking for historacle pricing.
- [1630](https://github.com/umee-network/umee/pull/1630) Incentive module proto.
- [1588](https://github.com/umee-network/umee/pull/1588) Historacle proto.
- [1653](https://github.com/umee-network/umee/pull/1653) Incentive Msg Server interface implementation.
- [1654](https://github.com/umee-network/umee/pull/1654) Leverage historacle integration.
- [1685](https://github.com/umee-network/umee/pull/1685) Add medians param to Token registry.
- [1683](https://github.com/umee-network/umee/pull/1683) Add MaxBorrow query and allow returning all denoms from MaxWithdraw.
- [1690](https://github.com/umee-network/umee/pull/1690) Add MaxBorrow message type.
- [1711](https://github.com/umee-network/umee/pull/1711) Add historic pricing information to leverage MarketSummary query.
- [1723](https://github.com/umee-network/umee/pull/1723) Compute borrow limits using the lower of either spot or historic price for each collateral token, and the higher of said prices for borrowed tokens. Remove extra spot/historic only fields in account summary.
- [1694](https://github.com/umee-network/umee/pull/1694) Add new sdkutil package to enhance common Cosmos SDK functionality. Here, the `ZeroCoin` helper function.

## [v3.3.0](https://github.com/umee-network/umee/releases/tag/v3.3.0) - 2022-12-20

### Features

- [1642](https://github.com/umee-network/umee/pull/1642) Added QueryMaxWithdraw and MsgMaxWithdraw.
- [1633](https://github.com/umee-network/umee/pull/1633) MarketSummary query now displays symbol price instead of base price for readability.

### Improvements

- [1659](https://github.com/umee-network/umee/pull/1659) Update to Cosmos SDK 0.46.7 and related dependencies (#1659)

### Fixes

- [1640](https://github.com/umee-network/umee/pull/1640) Migrate legacy x/leverage gov handler proposals to the new `MsgGovUpdateRegistry` messages.
- [1650](https://github.com/umee-network/umee/pull/1650) Fixes bug with reserves in ExportGenesis.
- [1642](https://github.com/umee-network/umee/pull/1642) Added missing CLI for QueryBadDebts.
- [1633](https://github.com/umee-network/umee/pull/1633) Increases price calculation precision for high exponent. assets.
- [1645](https://github.com/umee-network/umee/pull/1645) Fix: docker build & release.
- [1650](https://github.com/umee-network/umee/pull/1650) export genesis tracks reserves.

## [v3.2.0](https://github.com/umee-network/umee/releases/tag/v3.2.0) - 2022-11-25

Since `umeed v3.2` there is a new runtime dependency: `libwasmvm.x86_64.so v1.1.1` is required.
Building from source will automatically link the `libwasmvm.x86_64.so` created as a part of the build process (you must build on same host as you run the binary, or copy the `libwasmvm.x86_64.so` your lib directory).

### Features

- [1555](https://github.com/umee-network/umee/pull/1555) Updates IBC to v5.1.0 that adds adds optional memo field to `FungibleTokenPacketData` and `MsgTransfer`.
- [1577](https://github.com/umee-network/umee/pull/1577) Removes LIQUIDATOR build flag and adds `--enable-liquidator-query` or `-l` runtime flag to `umeed start`. See [README.md](README.md) file for more details.

### State Machine Breaking

- [1555](https://github.com/umee-network/umee/pull/1555) Enable GB Slashing.

### API Breaking

- [1578](https://github.com/umee-network/umee/pull/1578) Reorganize key constructors in x/leverage/types and x/oracle/types.

## [v3.1.0](https://github.com/umee-network/umee/releases/tag/v3.1.0) - 2022-10-22

### Features

- [1513](https://github.com/umee-network/umee/pull/1513) New query service exposing chain information via new RPC route. See [cosmos-sdk/11582](https://github.com/cosmos/cosmos-sdk/issues/11582).

### State Machine Breaking

- [1474](https://github.com/umee-network/umee/pull/1474) Enabled all Gravity Bridge claims.
- [1479](https://github.com/umee-network/umee/pull/1479) Add MsgSupplyCollateral.

### Fixes

- [1471](https://github.com/umee-network/umee/pull/1471) Fix slash window progress query.
- [1501](https://github.com/umee-network/umee/pull/1501) Cosmos SDK patch release.

## [v3.0.3](https://github.com/umee-network/umee/releases/tag/v3.0.3) - 2022-10-21

### Fixes

- [1511](https://github.com/umee-network/umee/pull/1511) Cosmos SDK patch release for Umee v3.0.3.

## [v3.0.2] - 2022-09-29

### Fixes

- [1460](https://github.com/umee-network/umee/pull/1460) Bump Gravity Bridge.

## [v3.0.1] - 2022-09-28

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
- [1118](https://github.com/umee-network/umee/pull/1118) MsgLiquidate rewards base assets instead of requiring an additional MsgWithdraw
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
- [1362](https://github.com/umee-network/umee/pull/1362) Remove inefficient BorrowAmounts and CollateralAmounts leverage invariants.
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
