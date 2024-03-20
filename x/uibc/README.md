# IBC Transfer and Rate Limits for IBC Denoms

## Abstract

The `x/uibc` is a Cosmos Module providing:

- IBC Quota is an ICS-4 middleware for the ICS-20 token transfer app to apply quota mechanism.

## Content

- [IBC ICS20 Hooks](#ics20-ibc-hooks)
- [IBC ICS20 Quota](#isc20-ibc-quota)

## IBC ICS20 Hooks

The IBC ICS20 hooks are part of our [ICS20 middleware](https://github.com/umee-network/umee/blob/main/x/uibc/uics20/ibc_module.go#L25) that enables ICS-20 token transfers to trigger message execution. This functionality allows cross-chain calls that involve token movement. IBC hooks are useful for a variety of use cases, including cross-chain swaps, which are an extremely powerful primitive.

### Concepts

Users can define ICS20 hook instructions in ICS20 transfer Memo field, that will trigger procedure call once the transfer is successfully recorded in the UX Chain.

### Design

The ICS20 packet data Memo field (introduced in [IBC v3.4.0](https://medium.com/the-interchain-foundation/moving-beyond-simple-token-transfers-d42b2b1dc29b)) allows to attach arbitrary data to a token transfer. The hook execution will be triggered if and only if:

- the packet data `memo` field can be JSON deserialized into the [`umee/uibc/v1/ICS20Memo`](https://github.com/umee-network/umee/blob/v6.4.0-beta1/proto/umee/uibc/v1/uibc.proto#L14). This means that the JSON serialized object into the memo string must extend the ICS20Memo struct.
- `ICS20Memo.fallback_addr`, if defined, must be a correct bech32 Umee address.

The fallback address is optional. It used when the memo is valid, but the hook execution (messages) fail. We strongly recommend to always use it. Otherwise the funds can be stuck in the chain if the hook execution fails.
If memo has a correct structure, and fallback addr is defined but malformed, we cancel the transfer (otherwise we would not be able to use it correctly).

The hooks processing has the following flow:

- Check ICS20 Quota. If quota exceeds, the transfer is cancelled.
- Deserialize Memo into `ICS20Memo`. If fails, assume the memo doesn't have hook instructions and continue with a normal transfer.
- Validate `ICS20Memo.fallback_addr`. If fails, the transfer is cancelled.
- Unpack and validate `ICS20Memo.messsages` (procedures). If fails, continue with the transfer, and overwrite the recipient with the `fallback_addr` if it is defined.
- Execute the transfer.
- Execute the hooks messages. If they fail and `fallback_addr` is defined, then revert the transfer (and all related state changes and events) and use send the tokens to the `fallback_addr` instead.

### Proto

```protocol-buffer
message ICS20Memo {
  // messages is a list of `sdk.Msg`s that will be executed when handling ICS20 transfer.
  repeated google.protobuf.Any messages = 1;
  // fallback_addr [optional] is a bech23 account address used to overwrite the original ICS20
  // recipient as described in the Design section above.
  string fallback_addr = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Supported Messages

The `ICS20Memo` is a list of native `sdk.Message`. Only the following combinations are supported currently:

- `[MsgSupply]`
- `[MsgSupplyCollateral]`
- `[MsgLiquidate]`

Validation:

- the operator (defined as the message signer) in each message, must be the same as the ICS20 transfer recipient,
- messages must only use the subset of the transferred tokens.

NOTE: because the received amount of tokens may be different than the amount originally sent (relayers or hop chains may charge transfer fees), if the amount of tokens in each message exceeds the amount received, we adjust the token amount in the messages.

### Example

Example 1: valid Memo Hook. Send `400 ibc/C0737D24596F82E8BD5471426ED00BDB5DA34FF13BE2DC0B23F7B35EA992B5CD` (here, the denom is an IBC denom of remote chain representing `uumee`). Execute a hook procedure to supply and collateralize `312uumee` and the reminder amount will be credited to the recipient balance.

```json
{
  "fallback_addr": "umee10h9stc5v6ntgeygf5xf945njqq5h32r5r2argu",
  "messages": [
    {
      "@type": "/umee.leverage.v1.MsgSupplyCollateral",
      "supplier": "umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due",
      "asset": { "denom": "uumee", "amount": "312" }
    }
  ]
}
```

Command to make a transaction:

```sh
umeed tx ibc-transfer transfer transfer channel-1 umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due  \
  400ibc/C0737D24596F82E8BD5471426ED00BDB5DA34FF13BE2DC0B23F7B35EA992B5CD \
  --from umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due \
  --memo '{"fallback_addr":"umee10h9stc5v6ntgeygf5xf945njqq5h32r5r2argu","messages":[{"@type":"/umee.leverage.v1.MsgSupplyCollateral","supplier":"umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due","asset":{"denom":"uumee","amount":"312"}}]}'
```

Example 2: memo struct is correct, struct is correct, but it doesn't pass the validation. The procedure wants to use `uother` denom, but the transfer sends `uumee` coins. Since the `fallback_addr` is correctly defined, all coins will go to the `fallback_addr`.

```json
{
  "fallback_addr": "umee10h9stc5v6ntgeygf5xf945njqq5h32r5r2argu",
  "messages": [
    {
      "@type": "/umee.leverage.v1.MsgSupplyCollateral",
      "supplier": "umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due",
      "asset": { "denom": "uother", "amount": "312" }
    }
  ]
}
```

Command to make a transaction:

```sh
$ umeed tx ibc-transfer transfer transfer channel-1 umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due \
  400ibc/C0737D24596F82E8BD5471426ED00BDB5DA34FF13BE2DC0B23F7B35EA992B5CD \
  --from umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due \
  --memo '{"fallback_addr":"umee10h9stc5v6ntgeygf5xf945njqq5h32r5r2argu","messages":[{"@type":"/umee.leverage.v1.MsgSupplyCollateral","supplier":"umee1y6xz2ggfc0pcsmyjlekh0j9pxh6hk87ymc9due","asset":{"denom":"uother","amount":"312"}}]}'
```

### Compatibility with IBC Apps Hooks

The IBC Apps repo has [`ibc-hooks`](https://github.com/cosmos/ibc-apps/tree/main/modules/ibc-hooks) middleware, which has similar functionality and the Memo structure is compatible with the one defined here. IBC App hooks only support cosmwasm procedures: the instruction is defined in the `Memo.wasm` field and fallback is fully handled by the CW contract. This implementation is compatible with IBC Apps: in the future we can support IBC Apps `wasm` hooks, without breaking our `Memo` struct.

### Limitations

- The current protocol requires that the IBC receiver is same as the "operator" (supplier, liquidator) in the `Memo.messages`.

## IBC ICS20 Quota

Hack or lending abuse is impossible to stop once the funds leave the chain. One mitigation is to limit the IBC inflows and outflows and be able to stop a chain and recover the funds with a migration.

### Concepts

Inflow is an ICS-20 transaction of sending tokens to the Umee chain.
Outflow is an ICS-20 transaction sending tokens out of the Umee chain.

IBC Quota is an upper limit in USD amount.

### Design

All inflows and outflows are measured in token average USD value using our x/oracle `AvgKeeper`. The `AvgKeeper` aggregates TVWAP prices over 16h window.

We are tracking inflows and outflows for tokens which are registered in x/leverage Token Registry.
NOTE: we measure per token as defined in the x/leverage, not the IBC Denom Path (there can be multiple paths). Since creating a channel is permission less, we want to use the same quota token.
For inflows:

- `inflows`: metric per token.
- `inflow_sum` : sum of all `inflows` from the previous point.

Similarly to inflows, we measure outflows per token and aggregates (sum):

- `outflows`: metric per token.
- `outflow_sum`: sum of `outflows` from the previous point.

The metrics above are reset every `params.quota_duration` in Begin Blocker.
Example: if the reset was done at 14:00 UTC, then the next reset will be done `quota_duration` later. You can observe the reset with `/umee/uibc/v1/EventQuotaReset` event, which will contain `next_expire` attribute.

#### Outflow Quota

Inflows and outflows metrics above are used to **limit ICS-20 transfers** of tokens in the x/leverage Token Registry. The outflow transfer of token `X` is possible when:

1. Outflow quota after the transfer is not suppressed:
1. `outflow_sum <= params.total_quota`. For example if it's set to 1.6M USD then IBC outflows reaching the total quota will be 600k USD worth of ATOM, 500k USD worth of STATOM, 250k USD worth of UMEE and 250k USD worth JUNO.
1. `token_quota[X] <= params.token_quota` - the token X quota is not suppressed.
1. OR Outflow quota lifted by inflows is not reached:
1. `outflow_sum <= params.inflow_outflow_quota_base + params.inflow_outflow_quota_rate * inflow_sum`
1. `token_quota[X] <= params.inflow_outflow_token_quota_base + params.inflow_outflow_token_quota_rate * inflows[X]`

See `../../proto/umee/uibc/v1/quota.proto` for the list of all params.

If a any `total_quota` or `token_quota` parameter is set to zero then we consider it as unlimited.

Transfer is **reverted** whenever it breaks any quota.

Transfer of tokens, which are not registered in the x/leverage Token Registry are not subject to the quota limit.

#### ICS-20 Quota control

The ICS-20 quota mechanism is controlled by the `Params.IbcStatus`, which can have the following values:

- DISABLED: inflow and quota outflow checks are disabled, essentially allowing all ics-20 transfers.
- ENABLED: inflow and quota outflow checks are enabled (default value).
- TRANSFERS_PAUSED: all ICS-20 transfers are disabled.

### State

In the state we store:

- Module [parameters](../../proto/umee/uibc/v1/quota.proto#L11).
- Running sum of total outflow values, serialized as `sdk.Dec`.
- Running sum of per token outflow values, serialized as `sdk.Dec`.
- Next quota expire time (after which the quota reset happens).
- Running sum of total inflow values, serialized as `sdk.Dec`.
- Running sum of per token inflow values, serialized as `sdk.Dec`.

### Messages

The RPC [Messages](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/tx.proto#L16) provide an access to the x/gov to change the module parameters.

### Queries

The RPC [Queries](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/query.proto#L15) allow to query module parameters and current outflow sums.

### Events

All events with description are listed in the [events.proto](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/events.proto) file.
