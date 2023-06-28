# Leverage Module

## Abstract

This document specifies the `x/leverage` module of the Umee chain.

The leverage module allows users to supply and borrow assets, and implements various features to support this, such as a token accept-list, a dynamic interest rate module, incentivized liquidation of undercollateralized debt, and automatic reserve-based repayment of bad debt.

The leverage module depends directly on `x/oracle` for asset prices, and interacts indirectly with `x/uibc`, `x/gravity`, and the cosmos `x/bank` module as these all affect account balances.

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
     - [Borrow Limit](#borrow-limit)
     - [Borrow Factor](#borrow-factor)
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

- Supply accepted asset types to the module, receiving _uTokens_ in exchange.

  Suppliers earn interest at an effective rate of the asset's [Supplying APY](#supplying-apy) as the [uToken Exchange Rate](#utoken-exchange-rate) increases over time.

  Additionally, for assets denominations already enabled as collateral, the supplied assets immediately become collateral as well, causing their borrow limit to increase.

  If a user is undercollateralized (borrowed value > borrow limit), collateral is eligible for liquidation and cannot be withdrawn until the user's borrows are healthy again.

  Care should be taken by undercollateralized users when supplying token amounts too small to restore the health of their borrows, as the newly supplied assets will be eligible for liquidation immediately.

- Enable or Disable (`MsgSetCollateral`) a uToken denomination as collateral for borrowing.

  Enabling _uTokens_ as collateral stores them in the `leverage` module account, so they cannot be transferred while in use. Disabling _uTokens_ as collateral returns them to the user's account. A user cannot disable a uToken denomination if it would reduce their [Borrow Limit](#borrow-limit) below their total borrowed value.

  If the user is undercollateralized (borrowed value > borrow limit), enabled collateral is eligible for liquidation and cannot be disabled until the user's borrows are healthy again.

- `MsgWithdraw` supplied assets by turning in uTokens of the associated denomination.
  Withdraw respects the [uToken Exchange Rate](#utoken-exchange-rate). A user can always withdraw non-collateral uTokens, but can only withdraw collateral-enabled uTokens if it would not reduce their [Borrow Limit](#borrow-limit) below their total borrowed value.

- `MsgMaxWithdraw` supplied assets by automatically calculating the maximum amount that can be withdrawn.
  This amount is calculated taking into account the available uTokens and collateral the user has, their borrow limit, and the available liquidity and collateral that can be withdrawn from the module respecting the `min_collateral_liquidity` of the `Token`.

- `MsgBorrow` assets of an accepted type, up to their [Borrow Limit](#borrow-limit).

  Interest will accrue on borrows for as long as they are not paid off, with the amount owed increasing at a rate of the asset's [Borrow APY](#borrow-apy).

- `MsgMaxBorrow` borrows assets by automatically calculating the maximum amount that can be borrowed. This amount is calculated taking into account the user's borrow limit and the module's available liquidity respecting the `min_collateral_liquidity` and `max_supply_utilization` of the `Token`.

- `MsgRepay` assets of a borrowed type, directly reducing the amount owed.

  Repayments that exceed a borrower's amount owed in the selected denomination succeed at paying the reduced amount rather than failing outright.

- `MsgLiquidate` undercollateralized borrows a different user whose total borrowed value is greater than their [Liquidation Threshold](#liquidation-threshold).

  The liquidator must select a reward denomination present in the borrower's uToken collateral. Liquidation is limited by [Close Factor](#close-factor) and available balances, and will succeed at a reduced amount rather than fail outright when possible.

  If a borrower is way past their borrow limit, incentivized liquidation may exhaust all of their collateral and leave some debt behind. When liquidation exhausts the last of a borrower's collateral, its remaining debt is marked as _bad debt_ in the keeper, so it can be repaid using module reserves.

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

#### Borrow Limit

Each token in the `Token Registry` has a parameter called `CollateralWeight`, always less than 1, which determines the portion of the token's value that goes towards a user's borrow limit, when the token is used as collateral.

A user's borrow limit is the sum of the contributions from each denomination of collateral they have deposited.

```go
  collateral := GetBorrowerCollateral(borrower) // sdk.Coins
  for _, coin := range collateral {
    borrowLimit += GetCollateralWeight(coin.Denom) * TokenValue(coin) // TokenValue is in usd
  }
```

For tokens with hith historic prices enabled (indicated by a `HistoricMedians` parameter greater than zero), each collateral `TokenValue` is computed with `PriceModeLow`, i.e. the lower of either spot price or historic price is used.

#### Borrow Factor

Each token in the `Token Registry` has a parameter called `CollateralWeight`, always less than 1, which determines the portion of the token's value that goes towards a user's borrow limit, when the token is used as collateral.

An implied parameter `BorrowFactor` is derived from `CollateralWeight` - specifically, it is the minimum of `2.0` and `1/CollateralWeight`.
The maximum borrow factor of `2.0` allows risky or non-collateral assets (`0 <= CollateralWeight < 0.5`) to be borrowed to a certain minimum degree.

When a user is borrowing, their borrow limit is whichever is more restrictive of the following two rules:

- Borrowed value must be less than collateral value times `CollateralWeight` (sum over each collateral asset)
- Borrowed value times `BorrowFactor` (sum over each borrowed asset) must be less than collateral value.

This means that when the original borrow limit based on collateral weight would allow a higher quality collateral to borrow a risky asset with a small margin of safety, the user's effective collateral weight is reduced to that of the riskier asset.
(Or `0.5` at the minimum.)

#### Historic Borrow Limit, Value

The leverage module also makes use of the oracle's historic prices to enforce an additional restriction on borrowing.

The logic is:

- For any `MsgBorrow`, `MsgMaxBorrow`, `MsgDecollateralize`, `MsgWithdraw`, or `MsgMaxWithdraw`
- The borrower’s borrowed value must be less than their borrow limit, with borrowed value being computed using `PriceModeHigh`, i.e. the higher of either spot price or historic price is used.
- Where historic prices are defined as the Median of the last `N` historic medians from the `oracle` module with `N = Token.HistoricMedians` in the leverage registry
- Else the transaction fails

#### Liquidation Threshold

Each token in the `Token Registry` has a parameter called `LiquidationThreshold`, always greater than or equal to collateral weight, but less than 1, which determines the portion of the token's value that goes towards a _borrower's_ liquidation threshold, when the token is used as collateral.

A user's liquidation threshold is the sum of the contributions from each denomination of collateral they have deposited. Any user whose borrow value is above their liquidation threshold is eligible to be liquidated.

```go
  collateral := GetBorrowerCollateral(borrower) // sdk.Coins
  for _, coin := range collateral {
     liquidationThreshold += GetLiquidationThreshold(coin.Denom) * TokenValue(coin) // TokenValue is in usd
  }
```

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
