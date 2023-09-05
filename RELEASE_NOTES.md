<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.0.0

Highlights:

- TODO
- New inflation mechanism (see below).

[v6.0.0 CHANGELOG](https://github.com/umee-network/umee/blob/v6.0.0/CHANGELOG.md).

### New Inflation Mechanism

- Set max inflation to 12 billions UMEE (the supply today is around 12.2 billions).

The **min and max inflation rates** decrease by `25%` every **two years**. We call it inflation cycles. In each inflation cycle, the yearly total amount of newly emitted tokens is a variable - At the end of each inflation cycle we need to decrease by 25%: the `min_inflation_rate`, `max_inflation_rate` and the `current_inflation_rate`. - Inflation rate change speed is changed from 1 year to **6 months**.

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
