<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v4.4.0

Highlights:

- TODO

<!-- TODO: See [CHANGELOG](https://github.com/umee-network/umee/blob/v4.3.0/CHANGELOG.md) for a full list of changes. -->

### Validators

We changed our GitHub release configuration, and now provide binaries in `.tar.gz` tarball.

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

## v4.3.0

Highlights:

- Gravity Bridge Shutdown Phase 2. Following [Prop-67](https://www.mintscan.io/umee/proposals/67) we are disabling Umee -> Ethereum token transfers. Ethereum -> Umee transfers are still possible, and we encourage everyone to move the tokens back to Umee. In May we are planning the complete shut down. See more in the [blog post](https://umee.cc/blog/bridgemigration).
- IBC updated to `ibc-go v6.1`. That also triggered our wasmvm dependency update (see `libwasmvm` update in Validators section)
- ICA Host integration.
- IBC ICS20: we will only accept tokens (denoms) which are registered in the x/leverage token registry. You can check the supported tokens by `umeed q leverage registered-tokens` or by visiting [umee/leverage/v1/registered_tokens](https://umee-api.polkachu.com/umee/leverage/v1/registered_tokens).

See [CHANGELOG](https://github.com/umee-network/umee/blob/v4.3.0/CHANGELOG.md) for a full list of changes.

### Validators

### libwasmvm update

Our dependencies have been updated. Now the binary requires `libwasmvm v1.2.1`. When you build the binary from source on the server machine you probably don't need any change. However when you download a binary from GitHub, or from other source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.2.1/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.2.1/internal/api/libwasmvm.$(uname -m).so -O /usr/lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

### Min Gas Prices

Same as with v4.2 release. We request all validators set a `minimum-gas-prices` setting (in app `config/app.toml` file, general settings). We recommend `0.1uumee` which is equal the current Keplr _average_ setting:

```
minimum-gas-prices = "0.1uumee"
```

You MUST also set the related parameter when starting Peggo `--cosmos-gas-prices="0.1uumee"`

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
