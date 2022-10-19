<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v3.1.0

This is a state machine breaking release. Coordinated update is required.

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
