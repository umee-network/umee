<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v4.2.0

The main highlight of this release is new `x/uibc` module which introduces IBC Quota functionality.
We are using it to limit daily (quota resets every 24 hours) IBC outflows in USD value based on our Avg price oracle update. Our current daily quota is 1M USD of total outflows and 600k USD of per token outflows. We only track the `x/leverage` registered tokens.
Example IBC outflows reaching the daily quota:

- 600k USD worth of ATOM (this is the max of a single token we can send per day), 400k USD worth of UMEE.
- 300k USD worth of ATOM, 200k USD worth of STATOM, 250k USD worth of UMEE and 250k USD worth JUNO.

Other highlights:

- Oracle: Avg prices. We compute average price based on ~16h rolling window.
- New option is available in `app.toml`: `iavl-lazy-loading` (in general settings). When setting to `true`, lazy loading of iavl store will be enabled and improve start up time of archive nodes.
- Migration from Tendermint to [CometBFT](https://github.com/cometbft/cometbft).

See [CHANGELOG](https://github.com/umee-network/umee/blob/v4.2.0/CHANGELOG.md) for a full list of changes.

### Validators

Given recent spam transactions in Umee, we request all validators set a `minimum-gas-prices` setting (in app `config/app.toml` file, general settings). We recommend `0.1uumee` which is equal the current Keplr _average_ setting:

```
minimum-gas-prices = "0.1uumee"
```

In next release we will be enforcing the minimum setting.

#### Peggo

Related to min gas price updates, you MUST also set the related parameter when starting Peggo:

```
--cosmos-gas-prices="0.1uumee"
```

### Upgrade instructions

- Note: Skip this step if you build binary from source and are able to properly link libwasmvm.
  - Download `libwasmvm`:

```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

- Download latest binary or build from source.
- Run the binary to make sure it works for you: `umeed --version`
- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Swap binaries.
- Restart the chain.
- Ensure latest Peggo (v1.4.0) is running
- Ensure latest Price Feeder (v2.1.0) is running and check your price feeder config is up to date.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

NOTE: BEFORE the upgrade, make sure the binary is working and libwasmvm is in your system. You can test it by running `./umeed-v4.1.0 --version`.

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.

## v4.1.0

This release contains several fixes designed to make lending and borrowing more resilient during price outages. Short summary of changes is available in the [Changelog](./CHANGELOG.md)

- [Price Feeder V2.1.0](https://github.com/umee-network/umee/releases/tag/price-feeder/v2.1.0) is recommended for use with this release. Upgrading price feeder can be done immediately by any validators who have not already switched. It does not need to be simultaneously with the chain upgrade.

**Please Note:**

Building from source will automatically link the `libwasmvm.x86_64.so` created as a part of the build process (you must build on the same host as you run the binary, or copy the `libwasmvm.x86_64.so` your lib directory).

If you build on system different than Linux amd64, then you need to download appropriate version of libwasmvm (eg from [CosmWasm/wasmvm Releases](https://github.com/CosmWasm/wasmvm/releases)) or build it from source (you will need Rust toolchain).

Otherwise you have to download `libwasmvm`. Please check [Supported Platforms](https://github.com/CosmWasm/wasmvm/tree/main/#supported-platforms). Example:

```bash
wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -P /lib/
```
