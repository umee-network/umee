<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v4.1.0

- new option is available in `app.toml`: `iavl-lazy-loading` (in general settings). When setting to `true`, lazy loading of iavl store will be enabled and improve start up time of archive nodes.

## v4.0.0

This release contains the Historacle Upgrade, a pricing update which improves the way we treat quickly-changing prices in the leverage module.

- See the [Historacle Design Doc](/docs/design_docs/011-historacle-pricing.md) for a description of how these prices are calculated.
- See the [Leverage Module Spec](/x/leverage/README.md#historic-borrow-limit-value) for a description of how these prices are treated by the leverage protocol.

**Please Note:**

- This upgrade requires the use of [Price Feeder V2.0.3](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv2.0.3) **AFTER** the Umee v4.0 Upgrade. Prior to this upgrade, you should stay on [Price Feeder V2.0.2](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv2.0.2).
- To run the provided binary, you **have to have `libwasmvm.x86_64.so v1.1.1`** in your system lib directory.

Building from source will automatically link the `libwasmvm.x86_64.so` created as a part of the build process (you must build on the same host as you run the binary, or copy the `libwasmvm.x86_64.so` your lib directory).

If you build on system different than Linux amd64, then you need to download appropriate version of libwasmvm (eg from [CosmWasm/wasmvm Releases](https://github.com/CosmWasm/wasmvm/releases)) or build it from source (you will need Rust toolchain).

Otherwise you have to download `libwasmvm`. Please check [Supported Platforms](https://github.com/CosmWasm/wasmvm/tree/main/#supported-platforms). Example:

```bash
wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -P /lib/
```

Additional highlights:

- [1694](https://github.com/umee-network/umee/pull/1694) `MsgMaxWithdraw`, `MsgMaxBorrow` and `MsgRepay` won't return errors if there is nothing to withdraw, borrow or repay respectively. Leverage `ErrMaxWithdrawZero` and `ErrMaxBorrowZero` has been removed.

Please see the [CHANGELOG](/CHANGELOG.md#v4.0.0) for an exhaustive list of changes.

### Update instructions

- Note: Skip this step if you build binary from source and are able to properly link libwasmvm.
  - Download `libwasmvm`:

```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Run latest Price Feeder (v2.0.3) - **updated**
- Swap binaries.
- Restart the chain.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

NOTE: BEFORE the upgrade, make sure the binary is working and libwasmvm is in your system. You can test it by running `./umeed-v4.0.0 --version`.

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
