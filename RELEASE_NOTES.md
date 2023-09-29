<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.1.0

- Enable [meToken module](https://github.com/umee-network/umee/blob/main/x/metoken/README.md): allows to create an
  index composed of a list of assets and realize operations such as swap and redeem between the assets and the
  index token.
- Bump go version to 1.21.
- Add spot price fields to account summary, and ensure all other fields use leverage logic prices.
- Fix avg params storage for x/oracle.

## v6.0.2

This fixes a crash shortly after the 6.0.1 upgrade. The crash occurred at height `8427849` but this binary works even if you switch to it immediately after the gov upgrade. Patch must be applied **as soon as possible**.

## v6.0.1

This is a bug fix release for the `leverage.MsgGovUpdateSpecialAssets` handler.
We also added `umeed q ugov emergency-group` CLI query. Users were able to query the Emergency Group address using REST.

[CHANGELOG](CHANGELOG.md)

## v6.0.0

Highlights:

- We introduce [Special Assets](https://github.com/umee-network/umee/blob/v6.0.0-beta2/x/leverage/README.md#special-asset-pairs): a new primitive to optimize positions in x/leverage.
- New [inflation mechanism](./docs/design_docs/012-umee-inflation-v2.md).
- [Emergency Groups](#emergency-groups).
- Full Gravity Bridge removal. We don't include GB module any more in Umee.
- New `MsgLeveragedLiquidate.MaxRepay` which allows to limit the liquidation size using the leveraged liquidation mechanism.
- Renamed ugov `EventMinTxFees` to `EventMinGasPrice`.

### New Inflation Mechanism

The Upgrade Handler sets the following values to the Umee `x/ugov` Inflation Cycle parameters:

- `max_supply = 21e18uumee` (21 billions UMEE)
- `inflation_cycle = time.Hour * 24 * 365 * 2` (2 years)
- `inflation_reduction_rate = 2500 basis points` (25%)

The new Inflation Cycle will start on 2023-10-15 15:00 UTC. This will mark the first inflation reduction from the current rates:

- `inflation_min` 7% → 5.25%
- `inflation_max` 14% → 10.5%

The x/staking Bonded Goal stays the same: 33.00%.

### Emergency Groups

Currently, any parameter update requires going through a standard governance process, which takes 4 days. In a critical situation we need to act immediately:

- Control IBC Quota parameters (eg disable IBC)
- apply safe updates to oracle, leverage or incentive module parameters.

Emergency Group can trigger safe parameter updates at any time as a standard transaction. The Emergency Group address is controlled by the Umee Chain governance (`x/gov`) and can be disabled at any time.

### Validators

#### libwasmvm update

Our dependencies have been updated. Now the binary requires `libwasmvm v1.3.0`. When you build the binary from source on the server machine you probably don't need any change. However when you download a binary from GitHub, or from other source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.3.0/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.3.0/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

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
