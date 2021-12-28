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

## Providers

The list of current supported providers:

- [Binance](https://www.binance.com/en)
- [Huobi](https://www.huobi.com/en-us/)
- [Kraken](https://www.kraken.com/en-us/)
- [Osmosis](https://app.osmosis.zone/)

## Usage

The `price-feeder` tool runs off of a single configuration file. This configuration
file defines what exchange rates to fetch and what providers to get them from.
In addition, it defines the oracle's keyring and feeder account information.
Please see the [example configuration](price-feeder.example.toml) for more details.

```shell
$ price-feeder /path/to/price_feeder_config.toml
```

## Configuration

### `server`

The `server` section contains configuration pertaining to the API served by the
`price-feeder` process such the listening address and various HTTP timeouts.

### `currency_pairs`

### `account`

The `account` section contains the oracle's feeder and validator account information.
These are used to sign and populate data in pre-vote and vote oracle messages.

### `keyring`

The `keyring` section contains Keyring related material used to fetch the key pair
associated with the oracle account that signs pre-vote and vote oracle messages.

### `rpc`

The `rpc` section contains the Tendermint and Cosmos application gRPC endpoints.
These endpoints are used to query for on-chain data that pertain to oracle
functionality and for broadcasting signed pre-vote and vote oracle messages.
