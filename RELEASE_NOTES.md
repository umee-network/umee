<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

The Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.5.0

Highlights:

- Cosmos SDK v0.47.11 update.
- Adding new `auction` module to our app.
- Removing `crisis` module from our app.

[CHANGELOG](CHANGELOG.md)

### Auction module

We propose a new Cosmos SDK module, that will provide mechanism for protocol owned auctions.

UX Chain will now auction a portion of collected fees and introduce a token burning mechanism, unlocking a way to a potentially deflationary UX token.

Documentation: [x/auction/README.md](https://github.com/umee-network/umee/blob/v6.5.0/x/auction/README.md)

### Validators

**Upgrade Title** (for Cosmovisor): **v6.5**.

Update Price Feeder to `umee/2.4.3+`.

NOTE: after the upgrade, you should restart your Price Feeder. We observed that Price Feeder doesn't correctly re-established a connection after the chain upgrade.

#### libwasmvm update

Our dependencies have been updated. The binary requires `libwasmvm v1.5.2`. When you build the binary from source on the server machine you probably don't need any change. However, when you download a binary from GitHub, or from another source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.5.2/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.5.2/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

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
