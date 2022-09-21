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
  - [Transaction Tips](https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/RELEASE_NOTES.md#transaction-tips-and-sign_mode_direct_aux)
  - [SIGN_MODE_DIRECT_AUX](https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/RELEASE_NOTES.md#transaction-tips-and-sign_mode_direct_aux)
  - transaction prioritization
- IBC v5.0

#### x/leverage settings

The leverage module is by default compiled without support for the `liquidation_targets` query.

Validators should NOT enable this query on their nodes - it is inefficient due to iterating over all borrower accounts, and can delay time-sensitive consensus operations when a sufficient number of addresses must be checked.

To run a node capable of supporting a liquidator, enable the query at compile time using `LIQUIDATOR=true make install`.

### Gravity Bridge

In `v1.1.x` (current mainnet) we disabled Gravity Bridge (GB) module due to Ethereum PoS migration (_the merge_).
This release is the first step to re-enable GB. We start by enabling validators update and evidence messages (`MsgValsetConfirm` and `MsgValsetUpdatedClaim`), but the bridge messages: batch creation, claims (both ways: Ethereum->Cosmos and Cosmos->Ethereum) remain disabled.

Validators are expected to run Peggo and update the valiator set in Gravity smart contract.

See [Gravity Bridge](https://github.com/umee-network/Gravity-Bridge/blob/module/v1.5.3-umee-1/module/RELEASE_NOTES.md) Release Notes.

### Update notes

Each validator MUST:

- Run Peggo (Gravity Bridge Orchestrator) v1.0.x.
- Run [Price Feeder](https://github.com/umee-network/umee/tree/main/price-feeder).
- Update `config.toml` file by setting `mempool.version="v1"`:

  ```
  [mempool]
  version = "v1"
  ```

Instructions: [umeeversity/validator](https://umeeversity.umee.cc/validators/mainnet-validator.html)

Failure to run Peggo and Price Feeder results in being slashed, as do certain types of misbehavior such as consistently submitting incorrect prices.
