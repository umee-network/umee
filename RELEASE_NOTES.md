<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v5.2.0

This is maintenance release, which prepares the chain for the v6.0 release.

Highlights:

- [x/leverage] Allow setting multiple tokens with the same symbol name but different denom. This fixes the oracle voting miss counter on duplicate symbol denoms.
- Adding Amino support to `x/leverage.MsgLeveragedLiquidate`.
- Fix MsgBeginUnbonding counting existing unbondings against max unbond twice.
- Fixes an x/oracle RPC endpoint spelling: `/umee/oracle/v1/validators/{validator_addr}/aggregate_vote`

More: [v5.2.0 CHANGELOG](https://github.com/umee-network/umee/blob/v5.2.0/CHANGELOG.md).

### Validators

#### Peggo

You can kill Peggo (there is no need to run it any more).

#### Min Gas Prices

Since v4.2 release we request all validators set a `minimum-gas-prices` setting (in app `config/app.toml` file, general settings). We recommend `0.1uumee` which is equal the current Keplr _average_ setting:

```
minimum-gas-prices = "0.1uumee"
```

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
