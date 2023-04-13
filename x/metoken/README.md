# MeToken Module

## Abstract

This document specifies the `x/metoken` module of the Umee chain.

The `metoken` module allows users to swap and redeem accepted assets for an Index meToken. The Index meToken will 
maintain the parity between underlying assets given a specific configuration.

The `metoken` module depends directly on `x/leverage` for supplying and withdrawing assets, and the cosmos `x/bank` 
module as these all affect account balances.

## Contents

1. **[Concepts](#concepts)**
    - [Accepted Assets](#accepted-assets)
    - [Index Parameters](#index-parameters)
    - [Swapping and Redeeming](#swapping-and-redeeming)
    - Important Derived Values:
        - [Dynamic Fee](#dynamic-fee)
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

At the foundation of the `metoken` module is the _Index Registry_, which contains a list of Indexes with their meTokens, accepted assets and other parameters.

This list is controlled by governance. Assets that are not in the index registry are not available for swapping or redeeming for the Index meToken.

In an extreme case an asset can be deleted from the _Index Registry_, this will disallow to swap given asset for the 
Index meToken or redeem a meToken for that asset.

### Index Parameters

The Index will have the following parameters:

 - MeToken denom: the denomination of the Index meToken that will be given to user in exchange for accepted assets.
 - MeToken max supply: the maximum amount of Index meTokens that can be minted. A swap that requires to mint more 
   Index meToken than this value will result in an error. 
 - Fee:
   - Min fee: the minimum fee to be charged to the user. The applied fee will tend to decrease down to this value, 
     when the accepted asset is undersupplied in the index. It must be less than Balanced and Max fees. Valid values 
     0-∞.
   - Balanced fee: the fee to be charged to the user when the index is balanced. It must be greater than Min fee and 
     lower than Max fee, it cannot be 0.
   - Max fee: the maximum fee to be charged to the user. The applied fee will tend to increase up to this value,
     when the accepted asset is oversupplied in the index. It must be greater than Min and Balanced fee. If the value 
     is 0, no Max fee will be applied.
 - Accepted Assets: a list where each asset will have the following parameters:
   - Asset denom: the denomination of the underlying asset.
   - Reserve portion: the portion of swapped assets that will be transferred to `metoken` module as reserves, and the 
     minimum portion that will be taken from `metoken` module reserves when a redemption occurs.
   - Target allocation: the portion of an accepted asset the Index is targeting to have. This value will be 
     calculated in the addition and the update of an Index, as well as when an asset is deleted from it. The value 
     would be 1/amount of assets in the Index.

### Swapping and Redeeming

Users have the following actions available to them:

- Swap accepted asset for Index meToken. Every accepted asset can be swapped for the Index meToken with 1:1 exchange 
  rate. The user will need to pay a [Dynamic Fee](#dynamic-fee) for the swap. The fee will be charged in the 
  accepted asset the Index meToken is swapped for.
  
  Index meToken amount needed for the swap will be minted and transferred to the user's account, while the accepted 
  asset for the swap will be transferred to the `leverage` module pools and the `metoken` module reserves. 
  The portion to be transferred to each one is determined by the _Index Registry_ configuration of each accepted asset.

- Redeem Index meToken for accepted asset. Index meToken can be redeemed for every accepted asset with 1:1 exchange 
  rate. The user will need to pay a [Dynamic Fee](#dynamic-fee) for the redemption. The fee will be charged in the
  accepted asset the Index meToken is redeemed for.

  Index meToken amount needed for the redemption will be withdrawn from the user's account and burned, while 
  the chosen asset to redeem will be transferred from the `leverage` module pools and the `metoken` module reserves 
  to the user's account. The portion to be withdrawn from each one is determined by the _Index Registry_ 
  configuration of each accepted asset.

  When it is not possible to withdraw the needed portion from the `leverage` module given its own constraints, the part 
  taken from the reserves will increase in order to complete the redemption, if possible.

### Derived Values

Some important quantities that govern the behavior of the `metoken` module are derived from a combination of 
parameters. The math and reasoning behind these values will appear below.

As a reminder, the following values are always available as a basis for calculations:

- Account Token balances, available through the `bank` module.
- Index parameters from the _Index Registry_.
- Total reserves of any Token denomination, saved in `metoken` module reserve balance.
- Total amount of any Token denomination transferred to the `leverage` module, stored in `metoken` module [State](#state).

The more complex derived values must use the values above as basis.

#### Dynamic Fee

The fee to be applied for the swap or the redemption will be dynamic and based on the deviation from the 
`target_allocation` of an asset and its current allocation in the Index. The fee will be transferred to the 
`metoken` module reserves. The formula for calculating the dynamic fee is as follows:

```
dynamic_fee = balanced_fee + [allocation_delta_percentage * (balanced_fee / 100)]

If the dynamic_fee is lower than min_fee   -> dynamic_fee = min_fee
If the dynamic_fee is greater than max_fee -> dynamic_fee = max_fee
```

Example for the meUSD index, and the following fee and accepted assets: 

```
- Fee:
  - Min: 0.001
  - Balanced: 0.2
  - Max: 0.5

- USDT:
  - target_allocation: 0.33333
- USDC: 
  - target_allocation: 0.33333
- IST: 
  - target_allocation: 0.33333

After several swaps the index has 1200 USDT, 760 USDC and 3000 IST. Total supply: 4960. 

Calculations for swap:
- USDT deviation from targeted allocation in percentages: (0.24193 - 0.33333)/(0.33333 / 100) = -0.0914  / 0.0033333 = -27.42027%
- USDC deviation from targeted allocation in percentages: (0.15322 - 0.33333)/(0.33333 / 100) = -0.18011 / 0.0033333 = -54.03354%
- IST  deviation from targeted allocation in percentages: (0.60483 - 0.33333)/(0.33333 / 100) =  0.2715  / 0.0033333 =  81.45081%

- USDT swap fee: 0.2 + [-27.42027*(0.2/100)] = 0.14515
- USDC swap fee: 0.2 + [-54.03354*(0.2/100)] = 0.09193
- IST  swap fee: 0.2 + [81.45081*(0.2/100)]  = 0.36290

Calculations for redemption:
- USDT deviation from targeted allocation in percentages: (0.33333 - 0.24193)/(0.33333 / 100) =  0.0914 / 0.0033333 = 27.42027%
- USDC deviation from targeted allocation in percentages: (0.33333 - 0.15322)/(0.33333 / 100) = 0.18011 / 0.0033333 = 54.03354%
- IST  deviation from targeted allocation in percentages: (0.33333 - 0.60483)/(0.33333 / 100) = -0.2715 / 0.0033333 = 81.45081%

- USDT swap fee: 0.2 + [27.42027*(0.2/100)]  = 0.25484
- USDC swap fee: 0.2 + [54.03354*(0.2/100)]  = 0.30806
- IST  swap fee: 0.2 + [-81.45081*(0.2/100)] = 0.03709
```

Another example with an edge case where the min and max fee are used:
```
- Fee:
  - Min: 0.01
  - Balanced: 0.3
  - Max: 0.8

- USDT:
  - target_allocation: 0.25
- USDC: 
  - target_allocation: 0.25
- IST: 
  - target_allocation: 0.25
- MSK: 
  - target_allocation: 0.25

After several swaps the index has 3500 USDT, 100 USDC, 300 IST and 0 MKS. Total supply: 3900. 

Calculations for swap:
- USDT deviation from targeted allocation in percentages: (0.89743 - 0.25)/(0.25 / 100) =  0.64743  / 0.0025 =  258.972%
- USDC deviation from targeted allocation in percentages: (0.02564 - 0.25)/(0.25 / 100) = -0.22436  / 0.0025 = -89.744%
- IST  deviation from targeted allocation in percentages: (0.07692 - 0.25)/(0.25 / 100) = -0.17308  / 0.0025 = -69.232%
- MKS  deviation from targeted allocation in percentages: (0.0 - 0.25)/(0.25 / 100)     = -0.25     / 0.0025 = -100%

- USDT swap fee: 0.3 + [258.972*(0.3/100)]   = 1.07691 -> This exceedes the max fee (0.8). In this case the max fee will be used.
- USDC swap fee: 0.3 + [-89.744*(0.3/100)]   = 0.03076
- IST  swap fee: 0.3 + [-69.232*(0.3/100)]   = 0.09230
- MKS  swap fee: 0.3 + [-100*(0.3/100)]      = 0       -> This is below the min fee (0.01). For this swap the min fee will be applied.

Calculations for redemption:
- USDT deviation from targeted allocation in percentages: (0.25 - 0.89743)/(0.25 / 100) = -0.64743  / 0.0025 = -258.972%
- USDC deviation from targeted allocation in percentages: (0.25 - 0.02564)/(0.25 / 100) =  0.22436  / 0.0025 =  89.74358%
- IST  deviation from targeted allocation in percentages: (0.25 - 0.07692)/(0.25 / 100) =  0.17308  / 0.0025 =  69.232%
- MKS  deviation from targeted allocation in percentages: (0.25 - 0.0)/(0.25 / 100)     =  0.25     / 0.0025 =  100%

- USDT swap fee: 0.3 + [-258.972*(0.3/100)] = -0.47691 -> This is below the min fee (0.01) and also the fee can't be negative. For this swap the min fee will be applied.
- USDC swap fee: 0.3 + [89.744*(0.3/100)]   = 0.56923
- IST  swap fee: 0.3 + [69.232*(0.3/100)]   = 0.50769
- MKS  swap fee: 0.3 + [100*(0.3/100)]      = 0.6      -> In this case the redeption is not possible since there is no MKS liquidity anyway.
```

### Reserves

The `metoken` module will have its own reserves to stabilize the processing of the withdrawals. A portion of 
every swap will be transferred to the reserves and a percentage of every withdrawal will be taken from the reserves. 
This portion is determined by the parameters of every asset. Every charged fee will be also transferred to the 
`metoken` module reserves.

#### Reserves Re-balancing

The reserves will be re-balanced twice a day. The frequency will be controlled by updates of the `rebalancing_block`.
The workflow for every asset of each Index is as follows:

- Get the amount of Token transferred to the `leverage` module, stored in `metoken` module [State](#state).
- Get the amount of Token maintained in the `metoken` module reserves.
- Check if the portion of reserves is bellow the desired and transfer the missing amount from `leverage` module to 
  `metoken` reserves.
- Update `rebalancing_block`, stored in the `metoken` module [State](#state).

## State

The `x/metoken` module keeps the following objects in state:

- Index Registry: `0x01 | index_name -> Index`
- Transferred to `leverage` module Amount: `0x02 | denom -> sdkmath.Int`
- Next Block for Reserves Re-balancing: `0x03 -> sdkmath.Int`
- Total Token Supply: `0x04 | denom -> sdk.Int`

The following serialization methods are used unless otherwise stated:

- `sdk.Dec.Marshal()` and `sdk.Int.Marshal()` for numeric types
- `[]byte(denom) | 0x00` for asset denominations (strings)

## Queries

See [metoken query proto](https://github.com/umee-network/umee/tree/main/proto/umee/metoken/v1/query.proto) for list of 
supported queries.

## Messages

See [metoken tx proto](https://github.com/umee-network/umee/tree/main/proto/umee/metoken/v1/tx.proto) for list of 
supported messages.

## Update Registry Proposal

`Update-Registry` gov proposal will add the new index to index registry or update the existing index with new settings.

### CLI
```bash
umeed tx gov submit-proposal [path-to-proposal-json] [flags]
```

Example:

```bash
umeed tx gov submit-proposal /path/to/proposal.json --from umee1..

// Note `authority` will be gov module account address in proposal.json
umeed q auth module-accounts -o json | jq '.accounts[] | select(.name=="gov") | .base_account.address'
```

where `proposal.json` contains:

```json
{
  "messages": [
    {
      "@type": "/umeenetwork.umee.metoken.v1.MsgGovUpdateRegistry",
      "authority": "umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp",
      "title": "Update the meToken Index Registry",
      "description": "Add meUSD Index, Update meEUR Index, Delete CNYT from meCNYT Index",
      "add_indexes": [
        {
          "metoken_denom": "meUSD",
          "metoken_max_supply": "2000000",
          "fee": {
            "min": "0.01",
            "balanced": "0.2",
            "max": "0.6"
          },
          "accepted_assets": [
            {
              "asset_denom": "USDT",
              "reserve_portion": "0.2"
            },
            {
              "asset_denom": "USDC",
              "reserve_portion": "0.2"
            },
            {
              "asset_denom": "IST",
              "reserve_portion": "0.2"
            }
          ]
        }
      ],
      "update_indexes": [
        {
          "metoken_denom": "meEUR",
          "metoken_max_supply": "2000000",
          "fee": {
            "min": "0.001",
            "balanced": "0.3",
            "max": "0.9"
          },
          "accepted_assets": [
            {
              "asset_denom": "EURC",
              "reserve_portion": "0.3"
            }
          ]
        }
      ],
      "delete_tokens": [
        {
          "metoken_denom": "meCNYT",
          "deleted_asset": "CNYT"
        }
      ]
    }
  ],
  "metadata": "",
  "deposit": "100uumee"
}
```

## Events

See [metoken events proto](https://github.com/umee-network/umee/tree/main/proto/umee/metoken/v1/events.proto) for 
list of supported events.

## Params

See [metoken module proto](https://github.com/umee-network/umee/tree/main/proto/umee/metoken/v1/metoken.proto) for 
list of supported module params.

## End Block

Every block, the `metoken` module runs the following steps in order:

- Re-balance Reserves if at `rebalancing_block`.