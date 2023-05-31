<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v5.0.0

Highlights:

- Cosmwasm integration.
- Gravity Bridge phase-3: shutdown of the transfers. In this release we introduce valset burn mechanism,
  which will block the Ethereum smart contract for processing any further transactions, as well
  as sending transfers back to Ethereum. This follows the plan approved through in the
  [prop-67](https://www.mintscan.io/umee/proposals/67).
  NOTE: All validators must continue to use Peggo to not get slashed.
- Updated to the latest Cosmos SDK v0.46.12

See [CHANGELOG](https://github.com/umee-network/umee/blob/v5.0.0-rc1/CHANGELOG.md)

### Validators

#### libwasmvm update

Our dependencies have been updated. Now the binary requires `libwasmvm v1.2.3`. When you build the binary from source on the server machine you probably don't need any change. However when you download a binary from GitHub, or from other source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.2.3/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.2.3/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

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
- Ensure latest Price Feeder (see [compatibility matrix](https://github.com/umee-network/umee/#release-compatibility-matrix)) is running and check your price feeder config is up to date.
- Restart the chain.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
