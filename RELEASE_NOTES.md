<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.3.0

Highlights:

- Cosmos SDK v0.47.7 patch update.
- New queries: `oracle/MissCounters`, `uibc/Inflows`, `uibc/QuotaExpires`, `leverage/RegisteredTokenMarkets`
- Update `uibc/MsgGovUpdateQuota` Msg type to handle the new inflow parameters.
- Update `uibc/QueryAllOutflowsResponse` to include denom symbol (token name) in every outflow.

[CHANGELOG](CHANGELOG.md)

### Validators

**Upgrade Title** (for Cosmovisor): **v6.3**.

### Upgrade instructions

- Download latest binary or build from source.
- Make sure `libwasmvm.$(uname -m).so` is properly linked
  - Run the binary to make sure it works for you: `umeed version`
- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Swap binaries.
- Ensure latest Price Feeder (see [compatibility matrix](https://github.com/umee-network/umee/#release-compatibility-matrix)) is running and check your price feeder config is up to date.
- Restart the chain.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
