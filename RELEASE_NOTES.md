<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v5.0.0

Highlights:

- Updated to the latest Cosmos SDK v0.46.12

See [CHANGELOG](https://github.com/umee-network/umee/blob/v5.0.0-rc1/CHANGELOG.md)

#### Min Gas Prices

Since v4.2 release we request all validators set a `minimum-gas-prices` setting (in app `config/app.toml` file, general settings). We recommend `0.1uumee` which is equal the current Keplr _average_ setting:

```
minimum-gas-prices = "0.1uumee"
```

You MUST also set the related parameter when starting Peggo `--cosmos-gas-prices="0.1uumee"`

### Upgrade instructions

- Download latest binary or build from source.
- Make sure `libwasmvm.$(uname -m).so` is properly linked
  - Run the binary to make sure it works for you: `umeed version`
- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Swap binaries.
- Ensure latest Peggo (v1.4.0) is running
- Ensure latest Price Feeder (v2.1.1) is running and check your price feeder config is up to date. Price Feeder was moved to the new repository: [ojo-network/price-feeder](https://github.com/ojo-network/price-feeder/tree/umee).
- Restart the chain.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
