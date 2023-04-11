# IBC Transfer and Rate Limits for IBC Denoms

## Abstract

The `x/uibc` is a Cosmos Module providing:

- IBC Denom Metadata Tracker for [ICS-20](https://github.com/cosmos/ibc/tree/main/spec/app/ics-020-fungible-token-transfer) transferred tokens to backfill denom metadata into the x/bank standard Cosmos SDK module.
- IBC Quota is an ICS-4 middleware for the ICS-20 token transfer app to apply quota mechanism.

## Content

- [IBC Denom Metadata Tracker](#ibc-denom-metadata-tracker)
- [IBC Quota](#ibc-quota)

## IBC Denom Metadata Tracker

`x/bank.types.Metadata` is a structure which provides essential information about denom, such as display denom name, description, symbol, list of units (unit name and decimal exponent), and the default unit (`Base`).

ICS-20 is a x/bank token transfer protocol over IBC.
The core implementation doesn't create bank `Metadata` when a new token is transferred for the very first time. It's worth to note that token received through IBC is identified by the port ID, channel ID and the source denom ID.
The purpose of the `x/uibc/ics20` module is to wrap the core IBC module and create a denom `Metadata` whenever it is missing. Look at the [`TrackDenomMetadata`](ics20/keeper/keeper.go) function for more details.

### Considerations

The IBC ICS-20 doesn't carry any metadata information, so we only fill up the base denom. Importantly, we don't know about the `Exponent`, and we set `Exponent := 0`. In many cases this is wrong, and should be overwritten by chain governance.

## IBC Quota

Hack or lending abuse is impossible to stop once the funds leave the chain. One mitigation is to limit the IBC inflows and outflows and be able to stop a chain and recover the funds with a migration.

### Concepts

Inflow is an ICS-20 transaction of sending tokens to the Umee chain.
Outflow is an ICS-20 transaction sending tokens out of the Umee chain.

IBC Quota is an upper limit in USD amount.

### Design

#### Outflows

All outflows are measured in token average USD value using our x/oracle `AvgKeeper`. The `AvgKeeper` aggregates TVWAP prices over 16h window.

We define 2 Quotas for ICS-20 transfers. Each quota only tracks tokens x/leverage Token Registry.

- `Params.TokenQuota`: upper limit of a sum of all outflows per token. Initially it's set to 0.6m USD per token. It limits the outflows value for each token.
  NOTE: we measure per token as defined in the x/leverage, not the IBC Denom Path (there can be multiple paths). Since creating a channel is permission less, we want to use same quota token.
- `Params.TotalQuota`: upper limit of a sum of all token outflows combined. Initially it's set to 1m USD. Example of IBC outflows reaching the total quota: 300k USD worth of ATOM, 200k USD worth of STATOM, 250k USD worth of UMEE and 250k USD worth JUNO.

If a quota parameter is set to zero then we consider it as unlimited.

All quotas are reset in `BeginBlocker` whenever a time difference between the new block, and the previous reset is more than `Params.QuotaDuration` in seconds (initially set to 24h).

Transfer is reverted whenever it breaks any quota.

Transfer of tokens, which are not registered in the x/leverage Token Registry are not subject to the quota limit.

#### Inflows

We only allow inflows of tokens registered in x/leverage Token Registry. Other inflow transfers will be rejected.

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

### Messages

The RPC [Messages](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/tx.proto#L16) provide an access to the x/gov to change the module parameters.

### Queries

The RPC [Queries](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/query.proto#L15) allow to query module parameters and current outflow sums.

### Events

All events with description are listed in the [events.proto](https://github.com/umee-network/umee/blob/main/proto/umee/uibc/v1/events.proto) file.
