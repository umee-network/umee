<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v4.0.0

This release contains the Historacle Upgrade, a pricing update which improves the way we treat quickly-changing prices in the leverage module.

> See the [Historacle Design Doc](/docs/design_docs/011-historacle-pricing.md) for a description of how these prices are calculated.
> See the [Leverage Module Spec](/x/leverage/README.md#historic-borrow-limit-value) for a description of how these prices are treated by the leverage protocol.

**Please Note** This upgrade requires the use of [Price Feeder V2.0.3](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv2.0.3) **AFTER** the Umee v4.0 Upgrade. Prior to this upgrade, you should stay on [Price Feeder V2.0.2](https://github.com/umee-network/umee/releases/tag/price-feeder%2Fv2.0.2).

Additional highlights:

- [1694](https://github.com/umee-network/umee/pull/1694) `MsgMaxWithdraw`, `MsgMaxBorrow` and `MsgRepay` won't return errors if there is nothing to withdraw, borrow or repay respectively. Leverage `ErrMaxWithdrawZero` and `ErrMaxBorrowZero` has been removed.

Please see the [CHANGELOG](/CHANGELOG.md#v4.0.0) for an exhaustive list of changes.

### Update instructions

- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Run latest Price Feeder (v2.0.3) - **updated**
- Swap binaries.
- Restart the chain.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

NOTE: BEFORE the upgrade, make sure the binary is working and libwasmvm is in your system. You can test it by running `./umeed-v4.0.0 --version`.

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
