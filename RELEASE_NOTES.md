<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v3.2.0

This is a state machine breaking release. Coordinated update is required.

Highlights:

- .

Please see the [CHANGELOG](https://github.com/umee-network/umee/blob/v3.1.0/CHANGELOG.md) for an exhaustive list of changes.

### Gravity Bridge

This is the final step for enabling Gravity Bridge. We enable slashing.
Validators must run Peggo and must process claims to not be slashed.

### Update instructions

- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Run latest Peggo (v1.2.1)
- Run latest Price Feeder (v1.0.0)
- Swap binaries.
- Restart the chain.
