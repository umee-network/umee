<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

The Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.5.0

In this release we are adding validations to ibc-transfer msg receiver address and memo feilds. This release will resolve the latest incident of spam ibc-transfer transactions.

## v6.4.1

This release updates our dependencies and applies latest patches to the v6.4.x line. All validators must update to this patch release.

## v6.4.0

Highlights:

- Cosmos SDK v0.47.10 patch update.
- IBC Hooks: we integrated ICS20 Memo handling.
- Integrated Packet Forwarding Middleware.
- Update `uibc/MsgGovUpdateQuota` Msg type to handle the new inflow parameters.
- Update `uibc/QueryAllOutflowsResponse` to include denom symbol (token name) in every outflow.

[CHANGELOG](CHANGELOG.md)

### IBC Hooks

This release brings the first part of the seamless cross-chain money market transactions. At UX, we want to provide the best User Experience for handling lending and leverage. In this release, we support the following `x/leverage` messages:

- `MsgSupply`
- `MsgSupplyCollateral`
- `MsgLiquidate`

The operation can only use tokens that are part of the IBC transfer (after any intermediate deductions) and the supplier / liquidator must be the IBC recipient (acting on someone else's behalf is not allowed). Authz is not supported. The remaining tokens will be credited to the recipient.

Documentation: [x/uibc/README.md](https://github.com/umee-network/umee/blob/v6.4.0/x/uibc/README.md#ibc-ics20-hooks)

### Validators

**Upgrade Title** (for Cosmovisor): **v6.4**.

Update Price Feeder to `umee/2.4.1+`.

NOTE: after the upgrade, you should restart your Price Feeder. We observed that Price Feeder doesn't correctly re-established a connection after the chain upgrade.

#### libwasmvm update

Our dependencies have been updated. The binary requires `libwasmvm v1.5.2`. When you build the binary from source on the server machine you probably don't need any change. However, when you download a binary from GitHub, or from another source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.5.2/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.5.2/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

### Upgrade instructions

- Download latest binary or build from source.
- Make sure `libwasmvm.$(uname -m).so` is properly linked
  - Run the binary to make sure it works for you: `umeed version`
- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Swap binaries.
- Ensure latest Price Feeder (see [compatibility matrix](https://github.com/umee-network/umee/#release-compatibility-matrix)) is running and ensure your price feeder configuration is up-to-date.
- Restart the chain.
- Restart Price Feeder.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
