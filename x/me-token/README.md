# Me-Token Module

## Abstract

This document specifies the `x/me-token` module of the Umee chain.

The `me-token` module allows users to swap and redeem stable assets for an index Token. This index Token will 
maintain the parity between underlying assets given a specific configuration.

The `me-token` module depends directly on `x/leverage` for supplying and withdrawing assets, and the cosmos `x/bank` 
module as these all affect account balances.

## Contents

1. **[Concepts](#concepts)**
    - [Accepted Assets](#accepted-assets)
    - [Swapping and Redeeming](#swapping-and-redeeming)
    - Important Derived Values:
        - [Incentivizing Displacement](#incentivizing-displacement)
        - [Exchange Rate](#exchange-rate)
    - [Reserves](#reserves)
2. **[State](#state)**
3. **[Queries](#queries)**
4. **[Messages](#messages)**
5. **[Update Registry Proposal](#update-registry-proposal)**
6. **[Events](#events)**
7. **[Parameters](#params)**
8. **[EndBlock](#end-block)**

## Concepts

### Accepted Assets

At the foundation of the `me-token` module is the _Index Registry_, which contains a list of indexes with their accepted assets and other parameters.

This list is controlled by governance and an emergency group. Assets that are not in the index registry are not available for swapping or redeeming for the index's Token.

Once added to the index registry, assets cannot be removed. In the rare case where an asset would need to be phased out, its target allocation can be changed to zero.

### Swapping and Redeeming

Users have the following actions available to them:

- Swap accepted asset for index Token with the [Incentivizing Displacement](#incentivizing-displacement) applied to 
  the initial 1:1 [Exchange Rate](#exchange-rate) to calculate the index Token amount.
  
  Calculated index Token amount will be minted and transferred to the user's account, meanwhile the accepted asset for 
  the swap will be transferred to the `leverage` module pools and the `me-token` module reserves. The portion to be 
  transferred to each one is determined by the _Index Registry_ configuration of each accepted asset.

- Redeem index Token for accepted asset with the [Incentivizing Displacement](#incentivizing-displacement) applied to
  the initial 1:1 [Exchange Rate](#exchange-rate) to calculate the asset amount.

  The index Token amount will be withdrawn from the user's account and burned, meanwhile the chosen asset to redeem 
  will be transferred from the `leverage` module pools and the `me-token` module reserves to the user's account.

  When is not possible to withdraw the needed portion from the `leverage` module given its own constraints, the part 
  taken from the reserves will increase in order to complete the redemption.

### Derived Values

Some important quantities that govern the behavior of the `me-token` module are derived from a combination of 
parameters. The math and reasoning behind these values will appear below.

As a reminder, the following values are always available as a basis for calculations:

- Account Token balances, available through the `bank` module.
- Index parameters from the _Index Registry_.
- Total reserves of any Token denomination, saved in `me-token` module reserve balance.
- Total amount of any Token denomination transferred to the `leverage` module, stored in `me-token` module [State](#state).

The more complex derived values must use the values above as a basis.

#### Incentivizing Displacement

The index has a parameter `max_incentivizing_displacement` and each accepted asset for the index has a parameter 
`target_allocation`. The `incetivizing_displacement` is calculated to affect the `exchange_rate` to incentivize or 
decentivize the swap and the redemption of a Token when its amount doesn't meet the `target_allocation`, using the 
following formula: 

`denom_incentivizing_displacement = index_max_incentivizing_displacement * [ (denom_target_allocation - 
denom_current_allocation) / (denom_target_allocation / 100)  / 100 ]`

Examples:
For the meUSD index the accepted assets with their `target_allocation` are: 

- USDT: 0.333
- USDC: 0.334
- IST: 0.333.

After several swaps the index has 1200 USDT, 760 USDC and 3000 IST. The incentive displacement for each Token is:

- USDT: 0.02733
- USDC: 0.05419
- IST:  -0.08168

#### Exchange Rate

The exchange rate is used for the swap and redemption of an accepted asset for index Token and is calculated as follows:

`exchange_rate(denom) = 1 + incentivizing_displacement(denom)`

Given the example above if a user realizes a swap of an accepted asset for meUSD he will receive the 
following amounts of it:

- 1.02733 meUSD for 1 USDT
- 1.05419 meUSD for 1 USDC
- 0.91832 meUSD for 1 IST

Following the same example if a user realizes a redemption of meUSD for an accepted asset he will receive the 
following amount of it:

- 0.97339 USDT for 1 meUSD
- 0.94859 USDC for 1 meUSD
- 1.08894 IST for 1 meUSD

### Reserves

The `me-token` module will have its own reserves to stabilize the processing of the withdrawals. A percentage of 
every swap will be transferred to the reserves and a percentage of every withdrawal will be taken from the reserves. 
This percentage is determined by the parameters of every asset.

#### Reserves Re-balancing

The reserves will be re-balanced twice a day. The workflow for every Token of each Index is as follows:

- Get the amount of Token is transferred to the `leverage` module, stored in `me-token` module [State](#state).
- Get the amount of Token is maintained in the `me-token` module reserves.
- Check if the portion of reserves is bellow the desired and transfer the missing amount from `leverage` module to 
  `me-token` reserves.
- Update `rebalancing_block`, stored in the `me-token` module [State](#state).

## State

The `x/me-token` module keeps the following objects in state:

- Registered Index (Index settings)\*: `0x01 | index_name -> Index`
- Transferred to `leverage` module Amount: `0x02 | denom -> sdkmath.Int`
- Next Block for Reserves Re-balancing: `0x03 -> sdkmath.Int`
- Total Token Supply: `0x04 | denom -> sdk.Int`

The following serialization methods are used unless otherwise stated:

- `sdk.Dec.Marshal()` and `sdk.Int.Marshal()` for numeric types
- `[]byte(denom) | 0x00` for asset denominations (strings)

## Queries

See `todo: add link for queries proto file after creation` for list of supported queries.

## Messages

See `todo: add link for messages proto file after creation` for list of supported messages.

## Update Registry Proposal

`Update-Registry` gov proposal will add the new index to index registry or update the existing index with new settings.

### CLI
```bash
TBD
```

Example:

```bash
TBD
```

where `proposal.json` contains:

```json
{
TBD
}
```

## Events

See `todo: add link for events proto file after creation` for list of supported events.

## Params

See `todo: add link for params proto file after creation` for list of supported module params.

## End Block

Every block, the `me-token` module runs the following steps in order:

- Re-balance Reserves if at `rebalancing_block`