# Release Notes

Release v0.2.0 of the Umee application. The release includes the following changes:

## Features

- Cosmos SDK version bumped to [v0.44.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.0)
- Full IBC compatibility based on ibc-go [v1.2.0](https://github.com/cosmos/ibc-go/releases/tag/v1.2.0)
- Gravity Bridge version bumped to [v0.2.7](https://github.com/PeggyJV/gravity-bridge/releases/tag/v0.2.7)

## Breaking Changes

- The `gorc` process for Gravity Bridge contains modifications to the configuration.
  See the release for further details.

e.g.

```toml
[cosmos.gas_price]
amount = 0.00001
denom = "uumee"

[metrics]
listen_addr = "127.0.0.1:3000"
```
