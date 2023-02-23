<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v4.2.0

- new option is available in `app.toml`: `iavl-lazy-loading` (in general settings). When setting to `true`, lazy loading of iavl store will be enabled and improve start up time of archive nodes.

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
### Upgrade instructions

- Note: Skip this step if you build binary from source and are able to properly link libwasmvm.
  - Download `libwasmvm`:

```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Swap binaries.
- Restart the chain.
- Ensure latest Peggo (v1.4.0) is running
- Ensure latest Price Feeder (v2.1.0) is running

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

NOTE: BEFORE the upgrade, make sure the binary is working and libwasmvm is in your system. You can test it by running `./umeed-v4.1.0 --version`.

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
