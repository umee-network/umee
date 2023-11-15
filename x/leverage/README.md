# Leverage Module

## Abstract

This document specifies the `x/leverage` module of the Umee chain.

The leverage module allows users to supply and borrow assets, and implements various features to support this, such as a token accept-list, a dynamic interest rate module, incentivized liquidation of undercollateralized debt, and automatic reserve-based repayment of bad debt.

The leverage module depends directly on `x/oracle` for asset prices, and interacts indirectly with `x/uibc`, and the cosmos `x/bank` module as these all affect account balances.

## Contents

1. **[Concepts](#concepts)**
   - [Accepted Assets](#accepted-assets)
     - [uTokens](#utokens)
   - [Supplying and Borrowing](#supplying-and-borrowing)
   - [Reserves](#reserves)
   - Important Derived Values:
     - [Adjusted Borrow Amounts](#adjusted-borrow-amounts)
     - [uToken Exchange Rate](#utoken-exchange-rate)
     - [Supply Utilization](#supply-utilization)
     - [Token Price](#token-price)
     - [Collateral Weight and Borrow Factor](#collateral-weight-and-borrow-factor)
     - [Borrow Limit](#borrow-limit)
     - [Max Borrow](#max-borrow)
     - [Liquidation Threshold](#liquidation-threshold)
     - [Borrow APY](#borrow-apy)
     - [Supplying APY](#supplying-apy)
     - [Close Factor](#close-factor)
     - [Total Supplied](#total-supplied)
2. **[State](#state)**
3. **[Queries](#queries)**
4. **[Messages](#messages)**
5. **[Update Registry Proposal](#update-registry-proposal)**
6. **[Events](#events)**
7. **[Parameters](#params)**
8. **[EndBlock](#end-block)**
   - [Bad Debt Sweeping](#sweep-bad-debt)
   - [Interest Accrual](#accrue-interest)

## Concepts

### Accepted Assets

At the foundation of the `leverage` module is the _Token Registry_, which contains a list of accepted types.

This list is controlled by governance. Assets that are not in the token registry are nor available for borrowing or supplying.

Once added to the token registry, assets cannot be removed. In the rare case where an asset would need to be phased out, it can have supplying or borrowing disabled, or in extreme cases, be ignored by collateral and borrowed value calculations using a blacklist.

#### uTokens

Every base asset has an associated _uToken_ denomination.

uTokens do not have parameters like the `Token` struct does, and they are always represented in account balances with a denom of `UTokenPrefix + token.BaseDenom`. For example, the base asset `uumee` is associated with the uToken denomination `u/uumee`.

### Supplying and Borrowing

Users have the following actions available to them:

- `MsgSupply` accepted asset types to the module, receiving _uTokens_ in exchange.

  Suppliers earn interest at an effective rate of the asset's [Supplying APY](#supplying-apy) as the [uToken Exchange Rate](#utoken-exchange-rate) increases over time.

  Supplying will fail if a token has reached its `max_supply`.

- `MsgCollateralize` or `MsgDecollateralize` a uToken as collateral for borrowing.

  Collaterized _uTokens_ are stored in the `leverage` module and they cannot be transferred until they are decollaterized or liquidated. Decolaterized _uTokens_  are returned back to the user's account. A user cannot decollateralize a uToken if it would reduce their [Borrow Limit](#borrow-limit) below their total borrowed value.

  If the user is undercollateralized (borrowed value > borrow limit), collateral is eligible for liquidation and cannot be decollateralized until the user's borrows are healthy again.

  Collateralize can fail if it would violate the module's `min_collateral_liquidity` for the token.

- `MsgSupplyCollateral` to combine the effects of `MsgSupply` and `MsgCollateralize`.
  Care should be taken by undercollateralized users when supplying token amounts too small to restore the health of their borrows, as the newly supplied assets will be eligible for liquidation immediately.

- `MsgWithdraw` supplied assets by turning in uTokens of the associated denomination.
  Withdraw respects the [uToken Exchange Rate](#utoken-exchange-rate).
  A user can always withdraw non-collateral uTokens, but can only withdraw collateral uTokens if it would not reduce their [Borrow Limit](#borrow-limit) below their total borrowed value.
  Users may also be preventing from withdrawing both non-collateral and collateral uTokens if it would violate the module's `min_collateral_liquidity`.

- `MsgMaxWithdraw` supplied assets by automatically calculating the maximum amount that can be withdrawn.
  This amount is calculated taking into account the available uTokens and collateral the user has, their borrow limit, and the available liquidity and collateral that can be withdrawn from the module respecting the `min_collateral_liquidity` and `max_supply_utilization` of the `Token`.

- `MsgBorrow` assets of an accepted type, up to their [Borrow Limit](#borrow-limit).
  Interest will accrue on borrows for as long as they are not paid off, with the amount owed increasing at a rate of the asset's [Borrow APY](#borrow-apy).
  Borrow can fail if it would violate the module's `max_supply_utilization` or `min_collateral_liquidity`.

- `MsgMaxBorrow` borrows assets by automatically calculating the maximum amount that can be borrowed. This amount is calculated taking into account the user's borrow limit and the module's available liquidity respecting the `min_collateral_liquidity` and `max_supply_utilization` of the `Token`.

- `MsgRepay` assets of a borrowed type, directly reducing the amount owed.

  Repayments that exceed a borrower's amount owed in the selected denomination succeed at paying the reduced amount rather than failing outright.

- `MsgLiquidate` repays undercollateralized borrows of a different user whose total borrowed value is greater than their [Liquidation Threshold](#liquidation-threshold) and receives some of their collateral as a reward.

  The liquidator must select a reward denomination present in the borrower's uToken collateral. Liquidation is limited by [Close Factor](#close-factor) and available balances, and will succeed at a reduced amount rather than fail outright when possible.

  If a borrower is way past their borrow limit, incentivized liquidation may exhaust all of their collateral and leave some debt behind. When liquidation exhausts the last of a borrower's collateral, its remaining debt is marked as _bad debt_ in the module, so it can be repaid using reserves.

- `MsgLeverageLiquidate` liquidates an account, but instead of repaying tokens using the liquidator's balance it borrows them instead. Additionally, the reward received is collateralized instantly.

  This allows more convenient liquidation where the liquidator does not need to keep balances of all potential repay tokens on hand, and can instead leverage a single type of collateral.

  From the module's point of view, no token or uToken transfers take place in this transaction, and the module's total borrowed and collateral amounts do not change.
  This makes leveraged liquidations immune to token liquidity exhaustion and harmless to module health measures like supply utilization and collateral liquitity.

  In addition, this transaction forces the liquidator to repay the maximum amount allowed by the user's `Close Factor` and the chosen repay and reward denoms.
  The transaction will fail if they would come close to exceeding their borrow limit in doing this, currently requiring `Borrowed Value / Borrow Limit < 0.9`.
  This is to prevent the liquidator from entering a position that is hard to unwind, and become at risk for liquidation in turn.

### Reserves

A portion of accrued interest on all borrows (determined per-token by the parameter `ReserveFactor`) is set aside as a reserves, which are automatically used to pay down bad debt.

Rather than being stored in a separate account, the `ReserveAmount` of any given token is stored in the module's state, after which point the module respects the reserved amount by treating part of the balance of the `leverage` module account as off-limits.

For example, if the module contains `1000 uumee` and `100 uumee` are reserved, then only `900 uumee` are available for Borrow and Withdraw transactions. If `40 uumee` of reserves are then used to pay off a bad debt, the module account will have `960 uumee` with `60 uumee` reserved, keeping the available balance at `900 uumee`.

### Oracle Rewards

At the same time reserves are accrued, an additional portion of borrow interest accrued is transferred from the `leverage` module account to the `oracle` module account to fund its reward pool. Because the transfer happens instantaneously and the accounts are separate, there is no need to module state to track the amounts.

### Derived Values

Some important quantities that govern the behavior of the `leverage` module are derived from a combination of parameters, borrow values, and oracle prices. The math and reasoning behind these values will appear below.

As a reminder, the following values are always available as a basis for calculations:

- Account token and uToken balances, available through the `bank` module.
- Total supply of any uToken denomination, stored in `leverage` module [State](#state).
- The `leverage` module account balance, available through the `bank` module.
- Collateral uToken amounts held in the `leverage` module account for individual borrowers, stored in `leverage` module [State](#state).
- Borrowed denominations and _adjusted amounts_ for individual borrowers, stored in `leverage` module state).
- _Interest scalars_ for all borrowed denominations, which are used with adjusted borrow amounts
- Total _adjusted borrows_ summed over all borrower accounts.
- Leverage module [Parameters](#params)
- Token parameters from the Token Registry

The more complex derived values must use the values above as a basis.

#### Adjusted Borrow Amounts

Borrow amounts stored in state are stored as `AdjustedBorrow` amounts, which can be converted to and from actual borrow amounts using the following relation:

> `AdjustedBorrow(denom,user)` \* `InterestScalar(denom)` = `BorrowedAmount(denom,user)`

When interest accrues on borrow positions, the `InterestScalar` of the denom is increased and the adjusted borrow amounts remain unchanged.

#### uToken Exchange Rate

uTokens are intended to work in the following way:

> The total supply of uTokens of a given denomination, if exchanged, are worth the total amount of the associated token denomination in the lending pool, including that which has been borrowed out and any interest accrued on it.

Thus, the uToken exchange rate for a given `denom` and associated `uDenom` is calculated as:

`exchangeRate(denom) = [ ModuleBalance(denom) - ReservedAmount(denom) + TotalBorrowed(denom) ] / TotalSupply(uDenom)`

In state, uToken exchange rates are not stored as the can be calculated on demand.

Exchange rates satisfy the invariant `exchangeRate(denom) >= 1.0`

#### Supply Utilization

Supply utilization of a token denomination is calculated as:

`supplyUtilization(denom) = TotalBorrowed(denom) / [ ModuleBalance(denom) - ReservedAmount(denom) + TotalBorrowed(denom) ]`

Supply utilization ranges between zero and one in general. In edge cases where `ReservedAmount(denom) > ModuleBalance(denom)`, utilization is taken to be `1.0`.

#### Token Price

The leverage module makes use of the oracle's spot prices and historic prices for all assets.

The spot price is the price which was voted on by validators during the most recent window (usually 30 seconds). If voting failed, this price will not exist.

The historic price is basically a median price for the asset over a given time period requested by the leverage module (`3 hours * Token.HistoricMedians`). For assets which do not use historic medians, the historic price simply returns the spot price.

Often the leverage module will select from both prices when deriving important values. For example:

- `PriceModeSpot` is used for most queries as well as liquidation transactions.
- `PriceModeHigh` takes the higher of spot and historic prices, and is used primarily to calculated borrowed value.
- `PriceModeLow` takes the lower of the two prices, and it used to calculate collateral value during borrow limit calculations.

Transactions will also have different behaviors when encountering missing spot or historic prices.
Missing collateral prices will allow a transaction to succeed if all _known_ collateral is sufficient to cover the user's resulting position.
Missing borrow prices cause any transaction which would increase borrowed value or decrease borrow limit to fail.

#### Collateral Weight and Borrow Factor

Each token in the `Token Registry` has a parameter called `CollateralWeight`, always less than 1, which determines the portion of the token's value that goes towards a user's borrow limit, when the token is used as collateral.

An additional implied parameter is defined as `BorrowFactor = maximum(0.5, CollateralWeight)`. Borrow factor limits the effectiveness of any collateral which is borrowing the token in question.

For example, an account using a single collateral token with `CollateralWeight 0.8` borrowing a single token with `CollateralWeight 0.7` will reduce the effective `CollateralWeight` of the account's collateral to `0.7` when computing borrow limit.

#### Special Asset Pairs

The leverage module can define pairs of assets which are advantaged when one is used as collateral to borrow the other.

They are defined in the form `[Asset A, Asset B, Special Collateral Weight, Special Liquidation Threshold]`. In effect, this means that

> When a user has collateral of `Asset A` and borrows `Asset B`, or vice versa, the `CollateralWeight` of both `Asset A` and `Asset B` are replaced by `Special Collateral Weight`. The `LiquidationThreshold` of the assets is also replaced by that of the special pair.

#### Special Asset Pair Examples

> Consider a scenario where assets `A,B,C,D` all have collateral weight `0.75`. There is also a special asset pair `[A,B,0.9]` which privileges borrows between those two assets.
> (Note: Liquidation threshold has been omitted from the special pair in this example.)
>
> A user with `Collateral: $10A, Borrowed: $7A` is unaffected by any special asset pairs. The maximum `A` it could borrow is `$7.50`
>
> A user with `Collateral: $10A + $10C, Borrowed: $7B, $7C` has a special pair in effect. The special pair resolves to `Collateral: $7.77A, Borrowed: $7B` and the rest of their position is treated normally as `Collateral: $2.23A + $10C, Borrowed: $7C` which has significant room to borrow more.
>
> A user with `Collateral: $10A + $10C, Borrowed: $15B` has a special pair in effect. The special pair resolves to `Collateral: $10A, Borrowed: $9B` and the rest of their position is treated normally as `Collateral: $10C, Borrowed: $6C` which can accomodate an additional `$1.50` of borrowing.

#### Borrow Limit

A user's borrow limit is the sum of the contributions from each collateral they have deposited, with some modifications due to `Borrow Factor` and `Special Asset Pairs`.

The full calculation of a user's borrow limit is as follows:

1. Calculate the USD value of the user's collateral assets, using the _lower_ of either spot price or historic price for each asset. Collateral with missing prices is treated as zero-valued when attempting to borrow new assets or withdraw collateral, but will block any liquidations until collateral price returns.
2. Calculate the USD value of the user's borrowed assets, using the _higher_ of either spot price or historic price for each asset. Borrowed assets with missing prices block any new borrowing or withdrawing of collateral, but are trated as zero valued during liquidations.
3. Sort all `Special Asset Pairs` with assets matching parts of the user's position, starting with the highest `Special Collateral Weight`.
4. For each special asser pair, match collateral tokens with borrowed tokens until one of the two runs out. The matched amounts satisfy `Collateral Value (A) * Special Collateral Weight (A,B) = Borrowed Value (B)` for each special asset pair `[A,B,CW]`. Subtract the collateral and borrowed tokens from the user's remaining position.
5. Then sum the `CollateralValue * CollateralWeight` for each unpaired collateral token, and subtract the sum of `BorrowedValue` for each unpaired borrow token. This value is the user's unused borrow limit (and is negative if they are over limit.)
6. Also sum `CollateralValue` for each collateral token, and subtract the sum of `BorrowedValue / max(0.5,CollateralWeight)` for each borrowed token. This value is the user's unused collateral according to borrow factor (and can also be negative, in which case it should be multiplied by the weighted average collateral weight of the collateral to reflect actual usage).
7. The user's current borrowed value, plus the lower of their unused borrow limit or unused collateral, is their borrow limit.

Note that the result of step 7 is the user's ideal borrow limit, their maximum borrowed value if all additional borrowed tokens had collateral weight greater than or equal to the weight of the remaining collateral, so as not to be limited by borrow factor.
When borrowing tokens with inferior `Borrow Factor`, the user's actual borrow limit will be lower.

#### Example Borrow Limit Calculation

> Collateral: $20 ATOM + $20 UMEE + $40 STATOM
> Borrowed: $50 ATOM
>
> Assume the following collateral weights: UMEE 0.35, ATOM 0.6, STATOM 0.5. Also assume a special asset pair [STATOM, ATOM, 0.75] is in effect. STATOM gets a boost when borrowing ATOM.
> Starting at step 3 above, since prices are already given, we first isolate any special asset pairs.
> $40 STATOM with a special collateral weight of 0.75 can borrow $30 ATOM. These amounts are matched with each other. The user's remaining position now looks like the following (sorted by collateral weights):
>
> Collateral: $20 ATOM (0.6) + $20 UMEE (0.35)
> Borrowed: $20 ATOM (0.6)
>
> Using collateral weights on the unpaired position, `$20 * 0.6 + $20 * 0.35 = $19`, and `$19 - $20 = $-1`, so the user's borrow limit is one dollar less than their current total borrowed value. `BorrowLimit = $49`.
>
> Using borrow factor on the unpaired position, `$20 + $20 - ($20 / 0.6) = $6.67`, which is greater than `$-1` so borrow factor does not cause any additional restrictions.

#### Max Borrow

This calculation must sometimes be done in reverse, for example when computing `MaxWithdraw` or `MaxBorrow` based on what change in the user's position would produce a `Borrow Limit` exactly equal to their borrowed value.
The result of these calculations will vary depending on the asset requested, or if it is part of any special pairs.

#### Example Max Borrow Calculation

See [EXAMPLES.md](./EXAMPLES.md) for an example of a max borrow calculation, including an edge case that would cause the module to return less than the true maximum after special pairs.

#### Liquidation Threshold

Any user whose borrow value is above their liquidation threshold is eligible to be liquidated.

Each token in the `Token Registry` has a parameter called `LiquidationThreshold`, always greater than or equal to collateral weight, but less than 1, which determines the portion of the token's value that goes towards a borrower's liquidation threshold when the token is used as collateral.

When a borrow position is limited by simple borrow limit (without special asset pairs or borrow factor), a user's liquidation threshold is the sum of the contributions from each denomination of collateral they have deposited:

```go
  collateral := GetBorrowerCollateral(borrower) // sdk.Coins
  for _, coin := range collateral {
     liquidationThreshold += GetLiquidationThreshold(coin.Denom) * TokenValue(coin) // TokenValue is in usd
  }
```

Liquidation threshold can also be reduced by borrow factor or increased by special asset pairs.
When those are taken into account, the procedure for deriving a user's liquidation threshold is identical to the procedure for borrow limit, except `LiquidationThreshold` is used instead of `CollateralWeight` for individual tokens and for special pairs.

#### Borrow APY

Umee uses a dynamic interest rate model. The borrow APY for each borrowed token denomination changes based on that token Supply Utilization.

The `Token` struct stored in state for a given denomination defines three points on the `Utilization vs Borrow APY` graph:

- At utilization = `0.0`, borrow APY = `Token.BaseBorrowRate`
- At utilization = `Token.KinkUtilization`, borrow APY = `Token.KinkBorrowRate`
- At utilization = `1.0`, borrow APY = `Token.MaxBorrowRate`

When utilization is between two of the above values, borrow APY is determined by linear interpolation between the two points. The resulting graph looks like a straight line with a "kink" in it.

#### Supplying APY

The interest accrued on borrows, after some of it is set aside for reserves, is distributed to all suppliers (i.e. uToken holders) of that denomination by virtue of the uToken exchange rate increasing.

While Supplying APY is never explicitly used in the leverage module due to its indirect nature, it is available for querying and can be calculated:

`SupplyAPY(token) = BorrowAPY(token) * SupplyUtilization(token) * [1.0 - ReserveFactor(token)]`

#### Close Factor

When a borrower is above their borrow limit, their open borrows are eligible for liquidation. In order to reduce the severity of liquidation events that can occur to borrowers that only slightly exceed their borrow limits, a dynamic `CloseFactor` applies.

A `CloseFactor` can be between 0 and 1. For example, a `CloseFactor = 0.25` means that a liquidator can at most pay back 25% of a borrower's current total borrowed value in a single transaction.

Two module parameters are required to compute a borrower's `CloseFactor` based on how far their `TotalBorrowedValue` exceeds their `BorrowLimit` (both of which are USD values determined using price oracles)

```go
portionOverLimit := (TotalBorrowedValue / BorrowLimit) - 1
// e.g. (1100/1000) - 1 = 0.1, or 10% over borrow limit

if portionOverLimit > params.CompleteLiquidationThreshold {
  CloseFactor = 1.0
} else {
  CloseFactor = Interpolate(             // linear interpolation
    0.0,                                 // minimum x
    params.CompleteLiquidationThreshold, // maximum x
    params.MinimumCloseFactor,           // minimum y
    1.0,                                 // maximum y
  )
}
```

Note that close factor is always `1.0` if borrowed value is below the module parameter `SmallLiquidationSize`.

#### Total Supplied

The `TotalSupplied` of a token denom is the sum of all tokens supplied to the asset facility, including those that have been borrowed out and any interest accrued, minus reserves.

`TotalSupplied(denom) = ModuleBalance(denom) - ReservedAmount(denom) + TotalBorrowed(denom)`

## State

The `x/leverage` module keeps the following objects in state:

- Registered Token (Token settings)\*: `0x01 | denom -> Token`
- Adjusted Borrowed Amount: `0x02 | borrowerAddress | denom -> sdk.Dec`
- Collateral Setting: `0x03 | borrowerAddress | denom -> 0x01`
- Collateral Amount: `0x04 | borrowerAddress | denom -> sdk.Int`
- Reserved Amount: `0x05 | denom -> sdk.Int`
- Last Interest Accrual (Unix Time): `0x06 -> int64`
- Bad Debt Instance: `0x07 | borrowerAddress | denom -> 0x01`
- Interest Scalar: `0x08 | denom -> sdk.Dec`
- Total Borrowed: `0x09 | denom -> sdk.Dec`
- Totak UToken Supply: `0x0A | denom -> sdk.Int`

The following serialization methods are used unless otherwise stated:

- `sdk.Dec.Marshal()` and `sdk.Int.Marshal()` for numeric types
- `[]byte(denom) | 0x00` for asset and uToken denominations (strings)
- `address.MustLengthPrefix(sdk.Address)` for account addresses
- `cdc.Marshal` and `cdc.Unmarshal` for `gogoproto/types.Int64Value` wrapper around int64

Note that collateral settings and instances of bad debt are both tracked using a value of `0x01`. In both cases, the `0x01` means `true` ("enabled" or "present") and a missing or deleted entry means `false`. No value besides `0x01` is ever stored.

### Adjusted Total Borrowed

Unlike all other quantities in state, `AdjustedTotalBorrowed` values are not present in imported and exported genesis state.

Instead, every time an individual `AdjustedBorrow` is set during `ImportGenesis`, its respective token's `AdjustedTotalBorrowed` is increased by the same amount. Thus, it is indirectly imported as the sum of individual positions.

Similarly, `AdjustedTotalBorrowed` is never set independently during regular operations. It is modified during calls to `setAdjustedBorrow`, always increasing or decreasing by the change in the individual borrow being set.

## Queries

See [leverage query proto](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/query.proto) for list of supported queries.

Additionally, the query `liquidation-targets` is only enabled if the node is started with a flag:

```bash
# Enabled
umeed start --enable-liquidator-query

# Enabled
umeed start -l

# Disabled
umeed start
```

## Messages

See [leverage tx proto](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/tx.proto#L11) for full documentation of supported messages.

Here are their basic functions:

### Supplying

- `MsgSupply`: Supplies base tokens to the module and receives uTokens in exchange. UTokens can later be used to withdraw.
- `MsgWithdraw`: Exchanges uTokens for the base tokens originally supplied, plus interest. UTokens withdrawn can be any combination of wallet uTokens (from supply) or collateral uTokens. When withdrawing collateral, borrow limit cannot be exceeded or the withdrawal will fail.
- `MsgMaxWithdraw`: Withdraws the maximum allowed amount of uTokens, respecting the user's borrow limit and the module's liquidity requirements, if any.

### Collateralizing

- `MsgCollateralize`: Sends uTokens to the module as the collateral. Collateral increases a user's borrow limit, but can be siezed in a liquidation if borrowed value exceeds a certain threshold above borrow limit due to price movements or interest owed. Collateral tokens still earn supply interest while collateralized.
- `MsgSupplyCollateral`: Combines `MsgSupply` and `MsgCollateralize`.
- `MsgDecollateralize`: Returns some collateral uTokens to wallet balance, without withdrawing base tokens. Borrow limit cannot be exceeded or the decollateralize will fail.

### Borrowing

- `MsgBorrow` Borrows base tokens from the module. Borrow limit cannot be exceeded or the transaction will fail.
- `MsgRepay` Repays borrowed tokens to the module, plus interest owed.

### Liquidation

- `MsgLiquidate` Liquidates a borrower whose borrowed value has exceeded their liquidation threshold (which is a certain amount above their borrow limit). The liquidator repays a portion of their debt using base tokens, and receives uTokens from the target's collateral, or the equivalent base tokens. The maximum liquidation amount is restricted by both the liquidator's specified amount and the borrower's liquidation eligibility, which may be partial.
- `MsgLeveragedLiquidate` Liquidates a borrower, but instead of repaying with base tokens from the liquidator wallet balance, moves the debt from the borrower to the liquidator and creates a new borrow position for the liquidator. Liquidator receives uTokens from the bororwer's collateral, and immediately collateralizes them to secure the liquidator positoin.

This transaction will succeed even if the liquidator could not afford to borrow the initial tokens (thanks to the new collateral position acquired from the borrower), as long as they are below 80% usage of their new borrow limit after the reward collateral is added.
The liquidator is left with a new borrow that they must pay off, and new collateral which can eventually be withdrawn.

## Update Registry Proposal

`Update-Registry` gov proposal will adds the new tokens to token registry or update the existing token with new settings.

Under certain conditions, tokens will be automatically deleted:

- The token has been blacklisted by a previous proposal or the current one
- The token has not been supplied to the module, so there are no uTokens, borrows, or collateral associated with it.

The conditions allow for mistakenly registered tokens which have never been used to be removed from the registry. It is not safe to remove a token with active supply or borrows, so those stay listed in the registry when blacklisted.

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
      "@type": "/umee.leverage.v1.MsgGovUpdateRegistry",
      "authority": "umee10d07y265gmmuvt4z0w9aw880jnsr700jg5w6jp",
      "title": "Update the Leverage Token Registry",
      "description": "Update the uumee token in the leverage registry.",
      "add_tokens": [
        {
          "base_denom": "uumee",
          "reserve_factor": "0.100000000000000000",
          "collateral_weight": "0.050000000000000000",
          "liquidation_threshold": "0.050000000000000000",
          "base_borrow_rate": "0.020000000000000000",
          "kink_borrow_rate": "0.200000000000000000",
          "max_borrow_rate": "1.500000000000000000",
          "kink_utilization": "0.200000000000000000",
          "liquidation_incentive": "0.100000000000000000",
          "symbol_denom": "UMEE",
          "exponent": 6,
          "enable_msg_supply": true,
          "enable_msg_borrow": true,
          "blacklist": false,
          "max_collateral_share": "0.900000000000000000",
          "max_supply_utilization": "0.900000000000000000",
          "min_collateral_liquidity": "0.900000000000000000",
          "max_supply": "123123"
        }
      ],
      "update_tokens": [
        {
          "base_denom": "uatom",
          "reserve_factor": "0.100000000000000000",
          "collateral_weight": "0.050000000000000000",
          "liquidation_threshold": "0.050000000000000000",
          "base_borrow_rate": "0.020000000000000000",
          "kink_borrow_rate": "0.200000000000000000",
          "max_borrow_rate": "1.500000000000000000",
          "kink_utilization": "0.200000000000000000",
          "liquidation_incentive": "0.100000000000000000",
          "symbol_denom": "ATOM",
          "exponent": 6,
          "enable_msg_supply": true,
          "enable_msg_borrow": true,
          "blacklist": false,
          "max_collateral_share": "0.900000000000000000",
          "max_supply_utilization": "0.900000000000000000",
          "min_collateral_liquidity": "0.900000000000000000",
          "max_supply": "123123"
        }
      ]
    }
  ],
  "metadata": "AQ==",
  "deposit": "100uumee"
}
```

## Events

See [leverage events proto](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/events.proto) for list of supported events.

## Params

See [leverage module proto](https://github.com/umee-network/umee/blob/main/proto/umee/leverage/v1/leverage.proto) for list of supported module params.

## End Block

Every block, the leverage module runs the following steps in order:

- Repay bad debts using reserves
- Accrue interest on borrows

### Sweep Bad Debt

Borrowers whose entire balance of collateral has been liquidated but still owe debt are marked by their final liquidation transaction. This periodic routine sweeps up all marked `address | denom` bad debt entries in the keeper, performing the following steps for each:

- Determine the about of [Reserves](#reserves) in the borrowed denomination available to repay the debt
- Repay the full amount owed using reserves, or the maximum amount available if reserves are insufficient
- Emit a "Bad Debt Repaid" event indicating amount repaid, if nonzero
- Emit a "Reserves Exhausted" event with the borrow amount remaining, if nonzero

### Accrue Interest

At every epoch, the module recalculates [Borrow APY](#borrow-apy) and [Supplying APY](#supplying-apy) for each accepted asset type, storing them in state for easier query.

Borrow APY is then used to accrue interest on all open borrows.

After interest accrues, a portion of the amount for each denom is added to the state's `ReservedAmount` of each borrowed denomination.

Then, an additional portion of interest accrued is transferred from the `leverage` module account to the `oracle` module to fund its reward pool.
