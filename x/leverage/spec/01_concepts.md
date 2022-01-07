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

- [Enable or Disable](04_messages.md#MsgSetCollateral) a _uToken_ denomination as collateral for borrowing.

  Enabling _uTokens_ as collateral stores them in the `leverage` module account so they cannot be transferred while in use. Disabling _uTokens_ as collateral returns them to the user's account. A user cannot disable a _uToken_ denomination if it would reduce their [Borrow Limit](01_concepts.md#Borrow-Limit) below their total borrowed value.

- [Withdraw](04_messages.md#MsgWithdrawAsset lent assets by turning in _uTokens_ of the associated denomination.

   Withdraw respects the [uToken Exchange Rate](01_concepts.md#uToken-Exchange-Rate). A user can always withdraw non-collateral uTokens, but can only withdraw collateral-enabled uTokens if it would not reduce their [Borrow Limit](01_concepts.md#Borrow-Limit) below their total borrowed value.

- [Borrow](04_messages.md#MsgBorrowAsset) assets of an accepted type, up to their [Borrow Limit](01_concepts.md#Borrow-Limit). 

  Interest will accrue on borrows for as long as they are not paid off, with the amount owed increasing at a rate of the asset's [Borrow APY](01_concepts.md#Borrow-APY).

- [Repay](04_messages.md#MsgRepayAsset) assets of a borrowed type, directly reducing the amount owed.

  Repayments that exceed a borrower's amount owed in the selected denomination succeed at paying the reduced amount rather than failing outright.

- [Liquidate](04_messages.md#MsgLiquidate) undercollateralized borrows a different user whose total borrowed value is greater than their [Borrow Limit](01_concepts.md#Borrow-Limit).

  The liquidator must select a reward denomination present in the borrower's _uToken_ collateral. Liquidation is limited by [Close Factor](01_concepts.md#Close-Factor) and available balances, and will succeed at a reduced amount rather than fail outright when possible.

  If a borrower is serverly past their borrow limit, incentivized liquidation may exhaust all of their collateral and leave some debt behind. When liquidation exhausts the last of a borrower's collateral, its remaining debt is marked as _bad debt_ in the keeper, so it can be repaid using module reserves.

## Reserves

A portion of accrued interest on all borrows (determined per-token by the parameter `ReserveFactor`) is set aside as a reserves, which are automatically used to pay down bad debt.

Rather than being stored in a separate account, the `ReserveAmount` of any given token is stored in the module's state, after which point the module respects the reserved amount by treating part of the balance of the `leverage` module account as off-limits.

For example, if the module contains `1000 uumee` and `100 uumee` are reserved, then only `900 uumee` are available for Borrow and Withdraw transactions. If `40 uumee` of reserves are then used to pay off a bad debt, the module acount will have `960 uumee` with `60 uumee` reserved, keeping the available balance at `900 uumee`.

## Derived Values

Some important quantities that govern the behavior of the `leverage` module are derived from a combination of parameters, borrow values, and oracle prices. The math and reasoning behind these values will appear below.

As a reminder, the following values are always available as a basis for calculations:
- User token balances, available through the `bank` module. This works for uTokens too.
- The `leverage` module account balance, available through the `bank` module.
- Collateral _uToken_ amounts held in the `leverage` module account for individual borrowers, stored in `leverage` module [State](02_state.md).
- Borrowed denominations and amounts for individual borrowers, stored in `leverage` module [State](02_state.md).
- Leverage module [Parameters](07_params.md)
- Token parameters from the [Token Registry](02_state.md#Token-Registry)

The more complex derived values must use the values above as a basis.

### uToken Exchange Rate

TODO

### Borrow Utilization

TODO

### Dynamic Interest Rate

TODO

### Borrow Limit

TODO

### Borrow APY

TODO

### Lending APY

TODO

### Close Factor

TODO

### Market Size

TODO