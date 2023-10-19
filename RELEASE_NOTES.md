<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->
<!-- markdownlint-disable MD040 -->

# Release Notes

Release Procedure is defined in the [CONTRIBUTING](CONTRIBUTING.md#release-procedure) document.

## v6.1.0

Highlights:

- [meToken](https://github.com/umee-network/umee/blob/main/x/metoken/README.md) module: allows to create an index composed of a list of assets and realize operations such as swap and redeem between the assets and the index token.
- Bump go version to 1.21.
- Add spot price fields to account summary, and ensure all other fields use leverage logic prices.
- Fix avg params storage for x/oracle.
- Emergency Groups are able to do security adjustments in x/metoken.
- Reduced `consensus.block.max_size` to 4 MB.

[CHANGELOG](CHANGELOG.md)

### meTokens

_meTokens_ is an abbreviation for Multi-Variant Elastic Tokens. We describe them as a general-purpose packaged asset builder. Simply put, they are tools to index multiple assets, which allows holders to build composable indexes that spread the risk associated with each asset bundled in an index.

MeTokens offer a way to build DeFi alternatives to TradFi and Web2 financial products. Think of a better, more liquid version of ETFs or mortgage-backed securities. The scope of buildable products will vary from builder to builder as the protocol allows for the creation of some of the most diverse use cases in the Cosmos Ecosystem.

### Validators

Following recent [asa-2023-002](https://forum.cosmos.network/t/amulet-security-advisory-for-cometbft-asa-2023-002/11604) security advisory we decreased the max block size to 4 MB.
Based on Notional [analysis](https://github.com/notional-labs/placid#reduce-the-size-of-your-chains-mempool-in-bytes), mempool size should be adjusted as well. We recommend to set it to `15 * 4 MB` (15 x block size):

```toml
max_txs_bytes = 60000000
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
