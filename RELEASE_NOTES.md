<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->

# Release Notes

## v3.0.0

v3.0.0 follows the _umeemania_ testnet release (v2.0.x) which presented our **lending** and **oracle** functionality.
Since umeemania, we spent time to hardened

### Highlights since v1.x

- `x/leverage` module, allowing anyone to:
  - supply liquidity and earn rewards,
  - collaterize the supplied liquidity to be able to borrow,
  - borrow (will charge the user borrow interest),
  - any umee holder governs the `x/leverage` parameters like: token registry and borrow / supply settings. See more about the parameters in [leverage.proto](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/leverage.proto) file.
- `x/oracle` module - a decentralized price oracle for the `x/leverage` module, as well as any app built in the Umee blockchain. In the future version we will offer oracle queries through IBC. Any Umee holders will govern oracle parameters defined in [oracle.proto](https://github.com/umee-network/umee/blob/main/proto/umee/oracle/v1/oracle.proto) file.

#### x/leverage settings

- TODO: important information about queries, and node parameters, eg: https://github.com/umee-network/umee/issues/1400

### Gravity Bridge

In `v1.1.x` (existing mainnet) we disable Gravity Bridge (GB) module due to Ethereum PoS migration (_the merge_).
This release is the first step to enable GB. We start by enabling validators update end evidence messages (`ValsetConfirm`), but the bridge messages: batch creation, claims (both ways: Ethereum->Cosmos and Cosmos->Ethereum) remain disabled.

See (TODO:) Gravity Bridge Release Notes.

### Update notes

Each validator must run:

- Peggo (Gravity Bridge Orchestrator).
- Price Feeder.

Instructions: https://umeeversity.umee.cc/validators/mainnet-validator.html

Not running above services, or missbehave is slasheable.
