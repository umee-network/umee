<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.2.0-beta

Highlights:

- Umee chain upgrades to the latest stable Cosmos SDK v0.47
- The `gov` module in in Cosmos SDK v0.47 has been updated to support a minimum proposal deposit at submission time. It is determined by a new parameter called `MinInitialDepositRatio`. When multiplied by the existing `MinDeposit` parameter, it produces the necessary proportion of coins needed at the proposal submission time. The motivation for this change is to prevent proposal spamming.
  We set `MinInitialDepositRatio` to 10%.`
- Added `meToken` WASM queries.
- IBC Quota v2 mechanism
  1. Outflows quota has been increased to `$1.6M` for total outflows and `$1.2M` per token outflows.
  2. new lifting conditions is added: IBC outflows are possible if (1.) fails, but
     - sum outflows of all tokens <= `$1M +  InflowOutflowQuotaRate * sum_of_all_inflows`;
     - and token outflows <= `$0.9M + InflowOutflowQuotaRate * token_inflows`.
  See [IBC Quota Design](./x/uibc/README.md#design) for more details.

[CHANGELOG](CHANGELOG.md)

### Validators

#### Price Feeder

Price Feeder `< umee/v2.3.0` is not compatible with Cosmos SDK v0.47. Validators must update to `umee/v2.3.0` or newer.

#### libwasmvm update

Our dependencies have been updated. Now the binary requires `libwasmvm v1.5.0`. When you build the binary from source on the server machine you probably don't need any change. However when you download a binary from GitHub, or from other source, make sure you update the `/usr/lib/libwasmvm.<cpu_arch>.so`. For example:

- copy from `$GOPATH/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.5.0/internal/api/libwasmvm.$(uname -m).so`
- or download from github `wget https://raw.githubusercontent.com/CosmWasm/wasmvm/v1.5.0/internal/api/libwasmvm.$(uname -m).so -O /lib/libwasmvm.$(uname -m).so`

You don't need to do anything if you are using our Docker image.

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
