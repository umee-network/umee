<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v3.2.0

This is a state machine breaking release. Coordinated update is required.

Highlights:

- IBC update to v5.1
- `QueryLiquidationTargets` RPC is now available when the node is run with `--enable-liquidator-query`. The LIQUIDATOR build flag has been removed. NOTE: this query should not be enabled for nodes with public API. The query involves intensive computation and can impact node stability when used by an attacker.
- Introduced experimental features, available when build with `experimental` flag. This flag must not be used on mainnet.

Please see the [CHANGELOG](https://github.com/umee-network/umee/blob/v3.2.0/CHANGELOG.md) for an exhaustive list of changes.

### Gravity Bridge

This is the final step for enabling Gravity Bridge. We enable slashing.
Validators must run Peggo and must process claims to not be slashed.

### Github Release

New experimental features which are part of the linked binary changed the build process. Umeed doesn't support static CGO build (with `CGO_ENABLED=1`). Github Actions only support build using Linux on amd64, we can not make a cross platform build using Github Actions. So our Github release only contains source code archive and amd64 Linux binary.

Moreover to run the provided binary, you need to have `libwasmvm.x86_64.so v1.1.1` in your system lib directory.

Building from source will automatically link the `libwasmvm.x86_64.so` created as a part of the build process (you must build on same host as you run the binary, or copy the `libwasmvm.x86_64.so` your lib directory).

Please check [Supported Platforms](https://github.com/CosmWasm/wasmvm/tree/v1.1.1/#supported-platforms) for `libwasmvm`

### Update instructions

- Note: Skip this step if you build binary from source 
    - Download `libwasmvm` 
```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Run latest Peggo (v1.3.0) - **updated**
- Run latest Price Feeder (v2.0.0) - **updated**
- Swap binaries.
- Restart the chain.

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).
- If you use Cosmovisor, you have to download the respective `libwasmvm` into your machine. 
```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

NOTE: As described in the previous section, you need to have `libwasmvm.x86_64.so` correctly linked to the binary. BEFORE the upgrade, make sure the binary is working. You can test it by running `./umeed-v3.2.0 --version`.
