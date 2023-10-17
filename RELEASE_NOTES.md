<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.1.0

Highlights:

- [meToken](https://github.com/umee-network/umee/blob/main/x/metoken/README.md) module: allows to create an index composed of a list of assets and realize operations such as swap and redeem between the assets and the index token.
- Bump go version to 1.21.
- Add spot price fields to account summary, and ensure all other fields use leverage logic prices.
- Fix avg params storage for x/oracle.
- Emergency Groups are able to do security adjustments in x/metoken.

[CHANGELOG](CHANGELOG.md)

### Emergency Groups

Currently, any parameter update requires going through a standard governance process, which takes 4 days. In a critical situation we need to act immediately:

- Control IBC Quota parameters (eg disable IBC)
- apply safe updates to oracle, leverage or incentive module parameters.

Emergency Group can trigger safe parameter updates at any time as a standard transaction. The Emergency Group address is controlled by the Umee Chain governance (`x/gov`) and can be disabled at any time.

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
