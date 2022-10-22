<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v3.1.0

This is a state machine breaking release. Coordinated update is required.

Updates:

- New `leverage/MsgSupplyCollateral` message which combines functionality of both supply and collaterization.
- New chain `/cosmos/base/node/v1beta1/config` query gRPC endpoint was integrated providing chain information such us `bond_denom`, `gas_prices`... See [cosmos-sdk/11582](https://github.com/cosmos/cosmos-sdk/issues/11582) for more details.

Please see the [CHANGELOG](https://github.com/umee-network/umee/blob/v3.1.0/CHANGELOG.md) for an exhaustive list of changes.

### Gravity Bridge

This is the second step for enabling Gravity Bridge.
We enable all messages. Peggo was updated to handle the bridge pause.

### Update instructions

- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Run latest Peggo (v1.2.1)
- Run latest Price Feeder (v1.0.0)
- Swap binaries.
- Restart the chain.

Validators are required to run Peggo in order to sync the Gravity Bridge messages.
