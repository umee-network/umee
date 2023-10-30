# meToken Module

## Abstract

This document specifies the `x/metoken` module of the Umee chain.

meToken is a new Token that represents an Index composed of assets used for swaps and redemptions. It can be minted
during the swap of the Index accepted assets and burned to redeem any Index accepted asset. Each Index will have a
unique name for the meToken that represents it. Its price is determined by the average price of the assets in the Index.

The `metoken` module allows users to swap and redeem accepted assets for an Index meToken. The Index meToken will
maintain the parity between underlying assets given a specific configuration. The module transfers part of the
supplied assets to the `leverage` module in order to accrue interest.

The `metoken` module depends directly on `x/leverage` for supplying and withdrawing assets, `x/oracle` for assets
prices and the cosmos `x/bank` module for balance updates (coins).

## Contents

1. **[Concepts](#concepts)**
    - [Accepted Assets](#accepted-assets)
    - [Index Parameters](#index-parameters)
    - [Dynamic meToken Price](#dynamic-metoken-price)
      - [Initial Price](#initial-price)
    - [Swapping and Redeeming](#swapping-and-redeeming)
    - Important Derived Values:
        - [Dynamic Fee](#dynamic-fee)
    - [Reserves](#reserves)
      - [Reserves Re-balancing](#reserves-re-balancing)
    - [Interest](#interest)
      - [Claiming Interests](#claiming-interests)
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

This list is controlled by the chain governance. Assets that are not in the index registry are not available for swapping or redeeming for the Index meToken.

Once an asset is added to an Index, it cannot be removed. However, its `target_allocation` can be changed to incentivize or disincentivize its presence in the Index.

### Index Parameters

The Index will have the following parameters:

- meToken denom: a denom of the meToken Index that will be given to user in exchange for accepted assets.
- meToken max supply: the maximum amount of meTokens (in specific Index) that can be minted. A swap that requires to
  mint more meToken than this value will result in an error.
- Fees: every fee is calculated and charged to the user in the asset that is used in the operation. The calculation
  will be explained below, the following values will be used as parameters for that calculation:
  - Min fee: the minimum fee to be charged to the user. The applied fee will tend to decrease down to this value,
    when the accepted asset is undersupplied in the index. It must be less than Balanced and Max fees.
    Valid values: `[0-1]`.
  - Balanced fee: the fee to be charged to the user when the index is balanced. It must be greater than Min fee and
    lower than Max fee, it cannot be 0.
    Valid values: `[0-1]`.
  - Max fee: the maximum fee to be charged to the user. The applied fee will tend to increase up to this value,
    when the accepted asset is oversupplied in the index. It must be greater than Min and Balanced fees.
    Valid values: `[0-1]`.
- Accepted Assets: a list where each asset will have the following parameters:
  - Asset denom.
  - Reserve portion: the portion of swapped assets that will be transferred to `metoken` module as reserves, and the
    minimum portion that will be taken from `metoken` module reserves when a redemption occurs. Valid values: `[0-1]`.
  - Target allocation: the portion of an accepted asset the Index is targeting to have. The sum of
    `target_allocation` of every accepted asset in the Index should be equal to 1. Valid values: `[0-1]`.

### Dynamic meToken Price

Every meToken will have a dynamic price. It will be based on the underlying assets total value, divided by the
amount of minted meTokens, and calculated for every operation. The formula for the price is as follows:

``` yaml
metoken_price = 
  (asset1_amount * asset1_price + asset2_amount * asset2_price + assetN_amount * assetN_price) / metoken_minted

As an example, using the following Index:
 - 2.5        WETH at price:      1858.5 USD
 - 6140       USDT at price:     0.99415 USD 
 - 1.75013446 WBTC at price: 28140.50585 USD
 - 6 minted meTokens
 
The price of meToken would be:
metoken_price = (2.5 * 1858.5 + 6140 * 0.99415 + 1.75013446 * 28140.50585) / 6
metoken_price = 10000 USD
```

#### Initial price

The initial price for the first transaction will be determined by the average price of the underlying assets,
divided by the quantity of accepted assets in the Index, using the following formula:

``` yaml
metoken_price = (asset1_price + asset2_price + assetN_price) / N

As an example for an Index composed of:
 - USDC at price: 1.018 USD
 - USDT at price: 0.983 USD
 - IST  at price: 1.035 USD
 
The price of meToken for the initial swap would be:
metoken_price = (1.018 + 0.983 + 1.035) / 3
metoken_price = 1.012 USD
```

### Swapping and Redeeming

Users have the following actions available to them:

- Swap accepted asset for Index meToken. Every accepted asset can be swapped for the Index meToken. The exchange
  rate will be determined using prices from `x/oracle` module for the accepted assets and the Index meToken
  dynamic price. The user will need to pay a [Dynamic Fee](#dynamic-fee) for the swap. The fee will be charged in
  the accepted asset the Index meToken is offered for.

  Index meToken amount needed for the swap will be minted and transferred to the user's account, while the accepted
  asset for the swap will be transferred to the `leverage` module pools and the `metoken` module reserves.
  The portion to be transferred to each one is determined by the _Index Registry_ configuration of each accepted asset.

  In case the defined portion to transfer to the `x/leverage` is not possible, because of the `leverage` module max
  supply cap for a particular token, the remaining part will be transferred to `x/metoken` reserves.

- Redeem Index meToken for accepted asset. Index meToken can be redeemed for every accepted asset. The exchange
  rate will be determined using prices from `x/oracle` module for the accepted assets and the Index meToken dynamic
  price. The user will need to pay a [Dynamic Fee](#dynamic-fee) for the redemption. The fee will be charged in the
  accepted asset the Index meToken is redeemed for.

  Index meToken amount needed for the redemption will be withdrawn from the user's account and burned, while
  the chosen asset to redeem will be transferred from the `leverage` module pools and the `metoken` module reserves
  to the user's account. The portion to be withdrawn from each one is determined by the _Index Registry_
  configuration of each accepted asset.

  When it is not possible to withdraw the needed portion from the `leverage` module given its own constraints, the part
  taken from the reserves will increase in order to complete the redemption, if possible.

### Derived Values

Some important quantities that govern the behavior of the `metoken` module are derived from a combination of
parameters. The math and reasoning behind these values are presented below.

As a reminder, the following values are always available as a basis for calculations:

- Account Token balances, available through the `bank` module.
- Index parameters from the _Index Registry_.
- Total reserves of any tokens, saved in `metoken` module reserve balance.
- Total amount of any tokens transferred to the `leverage` module, stored in `metoken` module [State](#state).
- The price of every underlying asset taken from `oracle` module.

The more complex derived values must use the values above as basis.

#### Dynamic Fee

The fee to be applied for the swap or the redemption will be dynamic and based on the deviation from the
`target_allocation` of an asset and its current allocation in the Index. Every charged fee to the user will be
transferred to the `metoken` module balance and the value will be added to the [State](#state). In that way it is
possible to discriminate between the reserves, fees and interest saved on the `metoken` module balance.
The formula for calculating the dynamic fee is as follows:

``` text
dynamic_fee = balanced_fee + [allocation_delta * (balanced_fee / 100)]

If the dynamic_fee is lower than min_fee   -> dynamic_fee = min_fee
If the dynamic_fee is greater than max_fee -> dynamic_fee = max_fee

where: 
allocation_delta for the swap = (current_allocation - target_allocation) / target_allocation
allocation_delta for the redemption = (target_allocation - current_allocation) / target_allocation
```

Example for the meUSD index, and the following fee and accepted assets:

``` yaml
- Fee:
  - Min: 0.001
  - Balanced: 0.2
  - Max: 0.5

- USDT:
  - reserve_portion: 0.2
  - target_allocation: 0.33333
- USDC: 
  - reserve_portion: 0.2
  - target_allocation: 0.33333
- IST: 
  - reserve_portion: 0.2
  - target_allocation: 0.33333
  
After several swaps the index has 1200 USDT, 760 USDC and 3000 IST. Total supply: 4960.
  
Prices:
 - USDT = 0.998 USD
 - USDC = 1.0   USD
 - IST  = 1.02  USD
 - meUSD_price = (1200 * 0.998 + 760 * 1.0 + 3000 * 1.02) / 4960 = 1.011612903

Calculations for swap:
- USDT allocation_delta: (0.24193 - 0.33333)/0.33333 = -0.2742027
- USDC allocation_delta: (0.15322 - 0.33333)/0.33333 = -0.5403354
- IST  allocation_delta: (0.60483 - 0.33333)/0.33333 =  0.8145081

- USDT fee: 0.2 - 0.2742027*0.2 = 0.14515
- USDC fee: 0.2 - 0.5403354*0.2 = 0.09193
- IST  fee: 0.2 + 0.8145081*0.2 = 0.36290

Following this values, let's simulate a swap. A user wants to buy meUSD for 10 USDT.

meUSD = 10 * (0.998 / 1.011612903) * (1 - 0.14515)
meUSD = 8.433465976

The user will receive 8.433465976 meUSD. 1.14515 USDT is the fee and 8.5485 USDT will be split 
between the reserves and the liquidity for the leverage module based on the reserve_portion.
6.8388 USDT will be transferred to the leverage module and 1.7097 USDT will be transferred to reserves.

Calculations for redemption:
- USDT allocation_delta: (0.33333 - 0.24193)/0.33333 =  0.2742027
- USDC allocation_delta: (0.33333 - 0.15322)/0.33333 =  0.5403354
- IST  allocation_delta: (0.33333 - 0.60483)/0.33333 = -0.8145081

- USDT fee: 0.2 + 0.2742027*0.2 = 0.25484
- USDC fee: 0.2 + 0.5403354*0.2 = 0.30806
- IST  fee: 0.2 - 0.8145081*0.2 = 0.03709

Following this values, let's simulate a redemption. A user wants to sell 20 meUSD for IST.

IST = 20 * (1.011612903 / 1.02) * (1 - 0.03709)
IST = 19.09984668

The user will receive 19.09984668 IST. 0.7357004428 IST will be the total saved fee. 20 meUSD will be taken from the 
user's account and burned.
The total value to be withdrawn is 19.83554712 IST and it will be split between the reserves and the liquidity from the
leverage module based on the reserve_portion.
15.8684377 IST will be taken from leverage module and 3.967109424 IST from the reserves.
```

Another example with an edge case where the min and max fee are used:

``` yaml
- Fee:
  - Min: 0.01
  - Balanced: 0.3
  - Max: 0.8

- USDT:
  - reserve_portion: 0.3
  - target_allocation: 0.25
- USDC: 
  - reserve_portion: 0.3
  - target_allocation: 0.25
- IST: 
  - reserve_portion: 0.3
  - target_allocation: 0.25
- MSK:
  - reserve_portion: 0.3 
  - target_allocation: 0.25
  
After several swaps the index has 3500 USDT, 100 USDC, 300 IST and 0 MSK. Total supply: 3900. 
  
Prices:
 - USDT = 0.998   USD
 - USDC = 0.99993 USD
 - IST  = 1.02    USD
 - MSK  = 1.0     USD
 - meUSD_price = (3500 * 0.998 + 100 * 0.99993 + 300 * 1.02 + 0 * 1.0) / 3900 = 0.9997417949

Calculations for swap:
- USDT allocation_delta: (0.89743 - 0.25)/0.25 =  2.58972
- USDC allocation_delta: (0.02564 - 0.25)/0.25 = -0.89744
- IST  allocation_delta: (0.07692 - 0.25)/0.25 = -0.69232
- MSK  allocation_delta: (0.0 - 0.25)/0.25     = -1.0

- USDT fee: 0.3 + 2.58972*0.3 = 1.07691 -> This exceedes the max fee (0.8). In this case the max fee will be used.
- USDC fee: 0.3 - 0.89744*0.3 = 0.03076
- IST  fee: 0.3 - 0.69232*0.3 = 0.09230
- MSK  fee: 0.3 - 1.0*0.3     = 0       -> This is below the min fee (0.01). For this swap the min fee will be applied.

Following this values, let's simulate a swap. A user wants to buy meUSD for 10 MSK.

meUSD = 10 * (1.0 / 0.9997417949) * (1 - 0.01)
meUSD = 9.902556891

The user will receive 9.902556891 meUSD. 0.1 MSK will be the total saved fee and 9.9 MSK will be splitted 
between the reserves and the liquidity for the leverage module based on the reserve_portion.
6.93 MSK will be transferred to the leverage module and 2.97 MSK will be transferred to reserves.

Calculations for redemption:
- USDT allocation_delta: (0.25 - 0.89743)/0.25 = -2.58972
- USDC allocation_delta: (0.25 - 0.02564)/0.25 =  0.8974358
- IST  allocation_delta: (0.25 - 0.07692)/0.25 =  0.69232
- MSK  allocation_delta: (0.25 - 0.0)/0.25     =  1.0

- USDT fee: 0.3 - 2.58972*0.3 = -0.47691 -> This is below the min fee (0.01) and also the fee can't be negative. For this swap the min fee will be applied.
- USDC fee: 0.3 + 0.89744*0.3 = 0.56923
- IST  fee: 0.3 + 0.69232(0.3 = 0.50769
- MSK  fee: 0.3 + 1.0*0.3     = 0.6      -> In this case the redeption is not possible since there is no MSK liquidity anyway.

Following this values, let's simulate a redemption. A user wants to sell 20 meUSD for USDC.

USDC = 20 * (0.9997417949 / 0.99993) * (1 - 0.56923)
USDC = 8.613778424

The user will receive 8.613778424 USDC. 11.38245721 USDC will be the total saved fee. 20 meUSD will be taken from the 
user's account and burned.
The total value to be withdrawn is 19.99623563 USDC and it will be splitted between the reserves and the liquidity from the
leverage module based on the reserve_portion.
13.99736494 USDC will be taken from leverage module and 5.99887069 USDC from the reserves.
```

### Reserves

The `metoken` module will have its own reserves to stabilize the processing of the withdrawals. A portion of
every swap will be transferred to the reserves and a percentage of every withdrawal will be taken from the reserves.
This portion is determined by the parameters of every asset in the Index.

All the reserves will be saved to the `metoken` module balance along with all the fees and the claimed interest. The
amount of fees for every asset will be saved to the `metoken` module [State](#state) as well as the amount of the
interests claimed from the `leverage` module. The reserves are equal to `balance - fee - interest` for every asset.

#### Reserves Re-balancing

The frequency of the Reserves Re-balancing will be determined by module parameter `rebalancing_frequency`.
The workflow for every asset of each Index is as follows:

 1. Get the amount of Token transferred to the `leverage` module, stored in `metoken` module [State](#state).
 2. Get the amount of Token maintained in the `metoken` module balance and deduct it by the fee amount and the
    interests amount, both stored in `metoken` module [State](#state). The result is the amount of Token reserves.
 3. Check if the portion of reserves is below the desired and transfer the missing amount from `leverage` module to
    `metoken` reserves, or vice versa if required.
 4. Update `next_rebalancing_time`, stored in the `metoken` module [State](#state) adding the `rebalancing_frequency` to
  the current block time.

### Interest

Every supply of liquidity to `leverage` module will produce interests. The interest will be accruing based on the
settings of the supplied Token in the [Token Registry](https://github.com/umee-network/umee/tree/main/x/leverage#accepted-assets)
and the [Supplying APY](https://github.com/umee-network/umee/tree/main/x/leverage#supplying-apy). Its usage will be
decided in future iterations.

#### Claiming Interests

Every `claim_interests_frequency` a process that withdraws all the accrued interest of every asset supplied to the
`leverage` module will be triggered. This process consists of the following steps:

 1. Get the amount of Token existing in the `leverage` module, stored in the `leverage` module balance for `metoken`
    module address and Token denom.
 2. Get the amount of Token transferred to the `leverage` module, stored in `metoken` module [State](#state).
 3. Calculate the delta 1) - 2), this will be the accrued interest.
 4. Withdraw accrued interest from `leverage` module.
 5. Update the claimed interest in the `metoken` module [State](#state).
 6. Update `next_interest_claiming_time`, stored in the `metoken` module [State](#state) adding the
    `claim_interests_frequency` to the current block time.

## State

The `x/metoken` module keeps the following objects in state:

- Index Registry: `0x01 | index_name -> Index`
- Module Balances of every Index: `0x02 | metoken_denom -> Balance`, where `Balance` is:
  - `metoken_supply`: total meToken Supply.
  - `leveraged`: transferred to `leverage` module Amount.
  - `reserved`: total `metoken` module reserves.
  - `fees`: total `fees` saved in reserves.
  - `interests`: total `interest` claimed from `x/leverage`.
- Next Time (unix timestamp) for Reserves Re-balancing: `0x03 -> int64`
- Next Time (unix timestamp) for Claiming Interests: `0x04 -> int64`

The following serialization methods are used unless otherwise stated:

- `sdk.Dec.Marshal()` and `sdk.Int.Marshal()` for numeric types
- `cdc.Marshal` and `cdc.Unmarshal` for `gogoproto/types.Int64Value` wrapper around int64

## Queries

See (LINK TO BE ADDED) for list of supported queries.

## Messages

See (LINK TO BE ADDED) for list of supported messages.

## Update Registry Proposal

`Update-Registry` gov proposal will add the new index to index registry or update the existing index with new settings.

### CLI

```bash
umeed tx gov submit-proposal [path-to-proposal-json] [flags]
```

Example:

```bash
umeed tx gov submit-proposal /path/to/proposal.json --from umee1..

# Note `authority` will be gov module account address in proposal.json
umeed q auth module-accounts -o json | jq '.accounts[] | select(.name=="gov") | .base_account.address'
```

where `proposal.json` contains:

```json
{
  "messages": [
    {
      "@type": "/umee.metoken.v1.MsgGovUpdateRegistry",
      "authority": "umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp",
      "title": "Update the meToken Index Registry",
      "description": "Add me/USD Index, Update me/EUR Index",
      "add_indexes": [
        {
          "metoken_denom": "me/USD",
          "metoken_max_supply": "2000000",
          "fee": {
            "min": "0.01",
            "balanced": "0.2",
            "max": "0.6"
          },
          "accepted_assets": [
            {
              "asset_denom": "USDT",
              "reserve_portion": "0.2",
              "total_allocation": "0.333"
            },
            {
              "asset_denom": "USDC",
              "reserve_portion": "0.2",
              "total_allocation": "0.334"
            },
            {
              "asset_denom": "IST",
              "reserve_portion": "0.2",
              "total_allocation": "0.333"
            }
          ]
        }
      ],
      "update_indexes": [
        {
          "metoken_denom": "me/EUR",
          "metoken_max_supply": "2000000",
          "fee": {
            "min": "0.001",
            "balanced": "0.3",
            "max": "0.9"
          },
          "accepted_assets": [
            {
              "asset_denom": "EURC",
              "reserve_portion": "0.3",
              "total_allocation": "0.333"
            }
          ]
        }
      ]
    }
  ],
  "metadata": "",
  "deposit": "100uumee"
}
```

## Events

See (LINK TO BE ADDED) for list of supported events.

## Params

See (LINK TO BE ADDED) for list of supported module params.

## End Block

Every block, the `metoken` module runs the following steps in order:

- Re-balance Reserves if at or after `next_rebalancing_time`.
- Claim interests from the `leverage` module if at or after `next_interest_claiming_time`.
