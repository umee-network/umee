<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD024 -->

# Release Notes

## v3.0.0

v3.0.0 improves upon the _umeemania_ testnet release (v2.0.x) which introduced our **lending** and **oracle** functionality.

### Highlights since v1.x

- `x/leverage` module, which allows anyone to:
  - supply liquidity (and earn interest)
  - collateralize the supplied assets to enable borrowing
  - borrow (and pay interest)
  - participate in governance of `x/leverage` [parameters](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/leverage.proto) file.
- `x/oracle` module - a decentralized price oracle for the `x/leverage` module, as well as any app built in the Umee blockchain. UMEE holders set `x/oracle` [parameters](https://github.com/umee-network/umee/blob/main/proto/umee/oracle/v1/oracle.proto) by governance.
- Cosmos v0.46 upgrade, which features:
  - [`x/group`](https://tutorials.cosmos.network/tutorials/understanding-group/) module
  - [`x/nft`](https://github.com/cosmos/cosmos-sdk/tree/v0.46.1/x/nft/spec) module
  - [Transaction Tips](Transaction Tips and SIGN_MODE_DIRECT_AUX)
  - [SIGN_MODE_DIRECT_AUX](Transaction Tips and SIGN_MODE_DIRECT_AUX)
  - transaction prioritization
- IBC v5.0

#### Fees

All transactions, except oracle messages, are required to pay gas. We implemented a consensus controlled `protocol_min_gas_price = 0.05uumee`. All **validators must** set their `minimum-gas-prices` settings in `app.yml` to a value at least `0.05uumee` (otherwise the node won't start). Transactions with gas price smaller then `protocol_min_gas_price` will fail during the DeliverTx (transaction execution) phase.
Oracle transactions are free only if they are composed from the prevote and vote messages and have gas limit <= 140'000 gas.

#### x/leverage settings

The leverage module is by default compiled without support for the `liquidation_targets` query.

Validators should NOT enable this query on their nodes - it is inefficient due to iterating over all borrower accounts, and can delay time-sensitive consensus operations when a sufficient number of addresses must be checked.

To run a node capable of supporting a liquidator, enable the query at compile time using `LIQUIDATOR=true make install`.

### Gravity Bridge

In `v1.1.x` (current mainnet) we disabled Gravity Bridge (GB) module due to Ethereum PoS migration (_the merge_).
This release is the first step to re-enable GB. We start by enabling validators update end evidence messages (`ValsetConfirm`), but the bridge messages: batch creation, claims (both ways: Ethereum->Cosmos and Cosmos->Ethereum) remain disabled.

See (TODO:) Gravity Bridge Release Notes.

### Update notes

Each validator must run:

- Peggo (Gravity Bridge Orchestrator).
- Price Feeder.

Instructions: https://umeeversity.umee.cc/validators/mainnet-validator.html

Failure to run Peggo and Price Feeder results in being slashed, as do certain types of misbehavior such as consistently submitting incorrect prices.
