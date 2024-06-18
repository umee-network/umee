<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

The Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.5.0

In this release, we are introducing validations for the IBC transfer message receiver address and memo fields. These enhancements aim to address and resolve the recent incident involving spam IBC transfer transactions.

- Maximum length for IBC transfer memo field: 32'768 characters
- Maximum length for IBC transfer receiver address field: 2'048 characters
- Bump `ibc-go` to v7.5.1.


### Validators

**Upgrade Title** (for Cosmovisor): **v6.5**.

Update Price Feeder to `umee/2.4.3+`.

NOTE: after the upgrade, you should restart your Price Feeder. We observed that Price Feeder doesn't correctly re-established a connection after the chain upgrade.

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
