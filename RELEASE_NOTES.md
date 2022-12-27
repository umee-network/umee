<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v3.3.0

- For the mainnet, this release includes update from v3.1.x → v3.3.x. Please also look at the [`v3.2` Release Notes](https://github.com/umee-network/umee/blob/v3.2.0/RELEASE_NOTES.md), notably the **Gravity Bridge Slashing**.
- For the Canon-2 testnet, this release includes update from v3.2.x → v3.3.x

v3.2.0 was not released on mainnet due to a bug in x/leverage gov messages migration to the new format which utilizes x/gov/v1 authorization system. The bug caused legacy token registry updates to break x/gov proposal queries. In v3.3 we fix that bug.

Additional highlights:

- Added `QueryMaxWithdraw` and `MsgMaxWithdraw` to allow user easily withdraw previously supplied tokens from the module back to the user balance.
- Updated Cosmos SDK to v0.46.7

Please see the [CHANGELOG](https://github.com/umee-network/umee/blob/v3.3.0/CHANGELOG.md) for an exhaustive list of changes.

### Github Release

Sinice `v3.2.0` new experimental features (disabled by default) are part of the linked binary. That changed the build process. Umeed officially doesn't support static CGO build (with `CGO_ENABLED=1`) any more. Github Actions only support build using Linux on amd64 -- we can not make a cross platform build using Github Actions (possible solution is to do it through Qemu emulator). So our Github release only contains source code archive and amd64 Linux binary.

To run the provided binary, you **have to have `libwasmvm.x86_64.so v1.1.1`** in your system lib directory.

Building from source will automatically link the `libwasmvm.x86_64.so` created as a part of the build process (you must build on the same host as you run the binary, or copy the `libwasmvm.x86_64.so` your lib directory).

If you build on system different than Linux amd64, then you need to download appropriate version of libwasmvm (eg from [CosmWasm/wasmvm Relases](https://github.com/CosmWasm/wasmvm/releases)) or build it from source (you will need Rust toolchain).

Otherwise you have to download `libwasmvm`. Please check [Supported Platforms](https://github.com/CosmWasm/wasmvm/tree/main/#supported-platforms). Example:

```bash
wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -P /lib/
```

### Update instructions

- Note: Skip this step if you build binary from source and are able to properly link libwasmvm.
  - Download `libwasmvm`:

```bash
$ wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.1.1/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so
```

- Wait for software upgrade proposal to pass and trigger the chain upgrade.
- Run latest Peggo (v1.4.0) - **updated**
- Run latest Price Feeder (v2.0.2) - **updated**
- Swap binaries.
- Restart the chain.

There is a new option available in `app.toml` (in Base Configuration). Set `iavl-disable-fastnode` to `true` if you want to disable fastnode cache and reduce RAM usage (default is `false`).

```
# IAVLDisableFastNode enables or disables the fast node feature of IAVL.
# Default is false.
iavl-disable-fastnode = false
```

You can use Cosmovisor → see [instructions](https://github.com/umee-network/umee/#cosmovisor).

- If you use Cosmovisor, and you didn't build binary from source in the validator machine, you have to download the respective `libwasmvm` into your machine. See the previous section for more details.

NOTE: BEFORE the upgrade, make sure the binary is working and libwasmvm is in your system. You can test it by running `./umeed-v3.3.0 --version`.

#### Docker

Docker images are available in [ghcr.io umee-network](https://github.com/umee-network/umee/pkgs/container/umeed) repository.
