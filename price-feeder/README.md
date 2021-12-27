# Oracle Price Feeder

The `price-feeder` tool is an extension of Umee's `x/oracle` module, both of
which are based on Terra's [x/oracle](https://github.com/terra-money/core/tree/main/x/oracle)
module and [oracle-feeder](https://github.com/terra-money/oracle-feeder). The
core differences are as follows:

- All exchange rates are quoted in USD or USD stablecoins.
- No need or use of reference exchange rates (e.g. Luna).
- No need or use of ToBin tax.
- The `price-feeder` combines both `feeder` and `price-server` into a single
  Golang-based application for better UX, testability and integration.

## Background

The `price-feeder` tool is responsible for performing the following:

1. Fetching and aggregating exchange rate price data from various providers, e.g.
   Binance and Osmosis, based on operator configuration. These exchange rates
   are exposed via an API and are used to feed into the main oracle process.
2. Taking aggregated exchange rate price data and submitting those exchange rates
   on-chain to Umee's `x/oracle` module following Terra's [Oracle](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html)
   specification.

## Usage
