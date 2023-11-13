<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.2.0

Highlights:

- Umee chain upgrades to the latest stable Cosmos SDK v0.47
- The `gov` module in in Cosmos SDK v0.47 has been updated to support a minimum proposal deposit at submission time. It is determined by a new parameter called `MinInitialDepositRatio`. When multiplied by the existing `MinDeposit` parameter, it produces the necessary proportion of coins needed at the proposal submission time. The motivation for this change is to prevent proposal spamming.
  We set `MinInitialDepositRatio` to 10%.`

### Validators

- Upgrade Price Feeder: TODO

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
