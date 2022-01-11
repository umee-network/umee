# Overview

This document covers basic concepts and math that determine the `leverage` module's behavior.

## Accepted Assets

At the foundation of the `leverage` module is the [Token Registry](02_state.md#Token-Registry), which contains a list of accepts base asset types.

This list is controlled by governance, and serves to limit the asset types available for transactions like borrowing and lending, and also any query services based on denomination.

### uTokens

Every base asset has an associated _uToken_ denomination.

uTokens do not have parameters like the `Token` struct does, and they are always represented in account balances with a denom of `UTokenPrefix + token.BaseDenom`. For example, the base asset `uumee` is associated with the uToken denomination `u/uumee`.

## Lending and Borrowing

Users have the following actions available to them:

- [Lend](04_messages.md#MsgLendAsset) accepted asset types to the module, receiving _uTokens_ in exchange.

  Lenders earn interest at an effective rate of the asset's [Lending APY](01_concepts.md#Lending-APY) as the [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate) increases over time.

- [Enable or Disable](04_messages.md#MsgSetCollateral) a uToken denomination as collateral for borrowing.

  Enabling _uTokens_ as collateral stores them in the `leverage` module account so they cannot be transferred while in use. Disabling _uTokens_ as collateral returns them to the user's account. A user cannot disable a uToken denomination if it would reduce their [Borrow Limit](01_concepts.md#Borrow-Limit) below their total borrowed value.

- [Withdraw](04_messages.md#MsgWithdrawAsset lent assets by turning in uTokens of the associated denomination.

   Withdraw respects the [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate). A user can always withdraw non-collateral uTokens, but can only withdraw collateral-enabled uTokens if it would not reduce their [Borrow Limit](01_concepts.md#Borrow-Limit) below their total borrowed value.

- [Borrow](04_messages.md#MsgBorrowAsset) assets of an accepted type, up to their [Borrow Limit](01_concepts.md#Borrow-Limit). 

  Interest will accrue on borrows for as long as they are not paid off, with the amount owed increasing at a rate of the asset's [Borrow APY](01_concepts.md#Borrow-APY).

- [Repay](04_messages.md#MsgRepayAsset) assets of a borrowed type, directly reducing the amount owed.

  Repayments that exceed a borrower's amount owed in the selected denomination succeed at paying the reduced amount rather than failing outright.

- [Liquidate](04_messages.md#MsgLiquidate) undercollateralized borrows a different user whose total borrowed value is greater than their [Borrow Limit](01_concepts.md#Borrow-Limit).

  The liquidator must select a reward denomination present in the borrower's uToken collateral. Liquidation is limited by [Close Factor](01_concepts.md#Close-Factor) and available balances, and will succeed at a reduced amount rather than fail outright when possible.

  If a borrower is serverly past their borrow limit, incentivized liquidation may exhaust all of their collateral and leave some debt behind. When liquidation exhausts the last of a borrower's collateral, its remaining debt is marked as _bad debt_ in the keeper, so it can be repaid using module reserves.

## Reserves

A portion of accrued interest on all borrows (determined per-token by the parameter `ReserveFactor`) is set aside as a reserves, which are automatically used to pay down bad debt.

Rather than being stored in a separate account, the `ReserveAmount` of any given token is stored in the module's state, after which point the module respects the reserved amount by treating part of the balance of the `leverage` module account as off-limits.

For example, if the module contains `1000 uumee` and `100 uumee` are reserved, then only `900 uumee` are available for Borrow and Withdraw transactions. If `40 uumee` of reserves are then used to pay off a bad debt, the module acount will have `960 uumee` with `60 uumee` reserved, keeping the available balance at `900 uumee`.

## Derived Values

Some important quantities that govern the behavior of the `leverage` module are derived from a combination of parameters, borrow values, and oracle prices. The math and reasoning behind these values will appear below.

As a reminder, the following values are always available as a basis for calculations:
- Account token and uToken balances, available through the `bank` module.
- Total supply of any token or uToken denomination, available through the `bank` module.
- The `leverage` module account balance, available through the `bank` module.
- Collateral uToken amounts held in the `leverage` module account for individual borrowers, stored in `leverage` module [State](02_state.md).
- Borrowed denominations and amounts for individual borrowers, stored in `leverage` module [State](02_state.md).
- Total borrows summed over all borrower accounts, derived from the above.
- Leverage module [Parameters](07_params.md)
- Token parameters from the [Token Registry](02_state.md#Token-Registry)

The more complex derived values must use the values above as a basis.

### uToken Exchange Rate

uTokens are intended to work in the following way:

> The total supply of uTokens of a given denomination, if exchanged, are worth the total amount of the associated token denomination in the lending pool, including that which has been borrowed out and any interest accrued on it.

Thus, the uToken exchange rate for a given `denom` and associated `uDenom` is calculated as:

`exchangeRate(denom) = [ ModuleBalance(denom) - ReservedAmount(denom) + TotalBorrowed(denom) ] / TotalSupply(uDenom)`

For efficiency, and because the exchange rate can only be affected by interest accruing (and not by `Lend`, `Withdraw`, `Borrow`, `Repay`, and `Liquidate` transactions), uToken exchange rates are calculated every `InterestEpoch` and stored in state.

Exchange rates satisfy the invariant

`exchangeRate(denom) >= 1.0`

### Borrow Utilization

Borrow utilization of a token denomination is calculated as:

`borrowUtilization(denom) = TotalBorrowed(denom) / [ ModuleBalance(denom) - ReservedAmount(denom) + TotalBorrowed(denom) ]`

Borrow utilization ranges between zero and one in general. In edge cases where `ReservedAmount(denom) > ModuleBalance(denom)`, utilization is taken to be `1.0`.

### Borrow Limit

Each token in the `Token Registry` has a parameter called `CollateralWeight`, always less than 1, which determines the portion of the token's value that goes towards a user's borrow limit, when the token is used as collateral.

A user's borrow limit is the sum of the contributions from each denomination of collateral they have deposited.

```go
  collateral := GetBorrowerCollateral(borrower) // sdk.Coins
  for _, coin := range collateral {
    borrowLimit += TokenValue(coin) // Oracle price of denomination * Amount
  }
```

### Borrow APY

Umee uses a dynamic interest rate model. The borrow APY for each borrowed token denomination changes based on that denomination's Borrow Utilization.

The `Token` struct stored in state for a given denomination defines three points on the `Utilization vs Borrow APY` graph:

- At utilization = `0.0`, borrow APY = `Token.BaseBorrowRate`
- At utilization = `Token.KinkUtilizationRate`, borrow APY = `Token.KinkBorrowRate`
- At utilization = `1.0`, borrow APY = `Token.MaxBorrowRate`

When utilization is between two of the above values, borrow APY is determined by linear interpolation between the two points. The resulting graph looks like a straight line with a "kink" in it.

### Lending APY

The interest accrued on borrows, after some of it is set aside for reserved is distributed to all lenders (i.e. uToken holders) of that denomination by virtue of the uToken exchange rate increasing.

While Lending APY is never explicity used in the leverage module due to its indirect nature, it is available for querying and can be calculated:

`LendAPY(token) = BorrowAPY(token) * BorrowUtilization(token) * [1.0 - ReserveFactor(token)]`

### Close Factor

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

### Market Size

The `MarketSize` of a token denom is the USD value of all tokens lent to the asset facility, including those that have been borrowed out and any interest accrued, minus reserves.

`MarketSize(denom) = oracle.Price(denom) * [ ModuleBalance(denom) - ReservedAmount(denom) + TotalBorrowed(denom) ]`