# Release Notes

## Improvements

- (deps) Update ibc-go to [v1.0.1](https://github.com/cosmos/ibc-go/releases/tag/v1.0.1)
  - Fixes a security vulnerability identified in transfer application (no funds are at risk)
- (deps) Update gravity-bridge to [v0.1.23](https://github.com/PeggyJV/gravity-bridge/releases/tag/v0.1.23)
  - Fixes typo in `send-to-ethereum` command
  - The `client` CLI is no longer needed; you can use `gorc deploy erc20` and `gorc eth-to-cosmos`

## Client Breaking Changes

- (gorc) The `metrics` config now only contains a single `listen_addr` field, e.g. `"127.0.0.1:3000"`
- (gorc) The former `cosmos.gas_price` has moved to a `[cosmos.gas_price]` stanza
  - Example:

    ```toml
    [cosmos.gas_price]
    amount = 0.0001
    denom = "uumee"
    ```
