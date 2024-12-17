<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

The Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.7.2

Highlights:

- cosmos-sdk update to v0.47.15
- fix the max_withdraw query
- deps upgrade

[CHANGELOG](CHANGELOG.md)

### Validators

Update Price Feeder to `umee/2.4.4+`.

NOTE: after the upgrade, you should restart your Price Feeder. We observed that Price Feeder doesn't correctly re-establish a connection after the chain upgrade.

#### libwasmvm update

Our dependencies have been updated. The binary requires `libwasmvm v1.5.5`. When you build the binary from source on the server machine you probably don't need any change. However, when you download a binary from GitHub, or from another source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.5.5/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.5.5/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

### Upgrade instructions

- Download latest binary or build from source.
- Make sure `libwasmvm.$(uname -m).so` is properly linked
  - Run the binary to make sure it works for you: `umeed version`
- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Swap binaries.
- Ensure latest Price Feeder (see [compatibility matrix](https://github.com/umee-network/umee/#release-compatibility-matrix)) is running and ensure your price feeder configuration is up-to-date.
- Restart the chain.
- Restart Price Feeder.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
