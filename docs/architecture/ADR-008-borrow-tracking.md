# ADR 008: Borrow Tracking

## Changelog

- Jan 27, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

One of the more computationally expensive operations in Umee is "iterate over all borrows". It is currently necessary on several occasions:

- When calculating supply utilization (requires total borrowed tokens)
- When accruing interest (dynamic interest requires supply utilization, and all borrows must be modified)
- When deriving uToken exchange rates (requires total borrowed tokens)

In order to reduce the performance impact of such iteration, `InterestEpoch` was created, so that such calculations only occurred every N blocks (e.g. 100).

This design proposal seeks to modify how borrows are stored, in order to achieve the same functionality without ever iterating over all borrows. As a result, `InterestEpoch` can also be eliminated, effectively happening every block.

## Decision

Chain state will now store an `InterestScalar` for each accepted asset denom, which starts at `1.0` and increases multiplicatively every time interest would accrue on borrows of that denom.

Instead of directly storing `sdk.Int BorrowedAmount` for each of a user's borrowed denoms, state records their `sdk.Dec AdjustedBorrow`, respecting the following definition:

> `AdjustedBorrow(denom, address)` \* `InterestScalar(denom)` = `BorrowedAmount(denom, address)`

State will additionally store a value `TotalAdjustedBorrows` for each denom, which tracks the sum of all users' adjusted borrows in the chosen denomination.
It must increase or decrease every time assets are borrowed or repaid (including bad debt repayment), but is not modified when interest accrues.

This change in borrow storage will be opaque to other parts of the leverage module: functions like `GetBorrow`, `GetTotalBorrows`, and `SetBorrow` will use `InterestScalar` internally, such that their inputs and outputs contain the real borrowed amount.

As a result of efficiency gains, the parameter `InterestEpoch` will be removed, with periodic functions taking place every block.

Additionally, the values `BorrowAPY`, `SupplyAPY`, and `uTokenExchangeRate` will be removed from state, instead being efficiently calculated when needed.

## Detailed Design

This decision mainly updates existing features, rather than adding new ones. The following changes are required to update the leverage module:

**Params:**

- Remove `InterestEpoch`

**State:**

- Add `InterestScalar` and `TotalAdjustedBorrows` to state, and add `Get/Set` functions
- Remove `BorrowAPY`, `SupplyAPY`, and `uTokenExchangeRate` from state, and remove `Set` functions

**Getters:**

- Modify `Get` functions for `BorrowAPY`, `SupplyAPY`, and `uTokenExchangeRate` to calculate values in real time.
- Modify `GetBorrow(addr,denom)`, `GetBorrowerBorrows(addr)`, and `GetAllBorrows()` to use `InterestScalar` and `AdjustedBorrow`
- Modify `GetTotalBorrows(denom)` to use `InterestScalar` and `TotalAdjustedBorrows`

**Setters:**

- Modify `SetBorrow(addr,denom)` to use `InterestScalar` and a stored `AdjustedBorrow`, and to additionally update `TotalAdjustedBorrows` based on the difference from the previous value (calls get before set.)

**Genesis:**

- Rename the `Borrow` struct in genesis state to `AdjustedBorrow`, with `Amount` field changing to `sdk.Dec` from `sdk.Int`

**Invariants:**

- Add invariant which checks `TotalAdjustedBorrows` against the sum of all `AdjustedBorrow` returned by `GetAllBorrows()`

### Example Scenarios

The following example scenario should help clarify the meaning of the `AdjustedBorrow` and `InterestScalar`:

> Assume a fresh system, containing a token denom `uumee` which has never been borrowed or accrued interest before.
>
> By definition, `TotalAdjustedBorrowed("uumee") = 0.0` and `InterestScalar("uumee") = 1.0`
>
> User `Alice` borrows `1000 uumee`. This information is stored in state as `AdjustedBorrow(alice,"uumee") = 1000.000`, because adjusted borrow amount is real borrow amount divided by interest scalar. As a result, `TotalAdjustedBorrow("uumee") = 1000.000`. Interest scalar is unchanged by borrowing.
>
> User `Bob` also borrows `2000 uumee`, which is stored as `AdjustedBorrow(bob,"uumee") = 2000.000`. As a result, `TotalAdjustedBorrow("uumee") = 3000.000`. Interest scalar is unchanged.
>
> Suppose that the interest rate for `uumee` borrows works out to be `0.0003% per block`. On the next EndBlock, `InterestScalar("uumee") *= 1.000003`. Both `AdjustedBorrow` values and `TotalAdjustedBorrow` are unchanged.
>
> Fast forward without any additional borrow or repay activity, to where `InterestScalar("uumee") = 1.5`. Due to interest accrued, Alice's real borrowed amount is `1500 uumee`, and Bob's is `3000 uumee`. Total borrowed across the system is thus `4500 uumee`. This has been accomplished by changing InterestScalar, so `AdjustedBorrow` and `TotalAdjustedBorrow` values are unchanged.
>
> For example, `AdjustedBorrow(alice,"uumee") == 1000.000`, so to get alice's real borrowed amount, `GetBorrow(alice,"uumee") = 1000.000 * 1.5 = 1500uumee`.
>
> Now Alice wished to borrow an additional `500 uumee`, so whe will owe a total of 2000. Her adjusted borrow is increased by the newly borrowed amount divided by InterestScalar: `AdjustedBorrow(alice,"uumee") = 1000.000 + (500 / 1.5) = 1333.333`.
>
> In addition, `TotalAdustedBorrow("uumee") = 3000.000 + (500 / 1.5) = 3333.333` reflects the increase. Note that the total `uumee` borrowed across the system is now `3333.333 * 1.5 = 4500 + 500 = 5000`.
>
> Finally, Bob will attempt to repay 1000 of his `3000 uumee` owed. Note that `AdjustedBorrow(bob,"uumee") = 2000.000` before the transaction.
>
> After Bob makes his repayment, `AdjustedBorrow(bob,"uumee") = 2000.000 - (1000 / 1.5) = 1333.333` and `TotalAdjustedBorrow("uumee") = 3333.333 - (1000 / 1.5) = 2666.666`. The same amount is subtracted from both quantities.

The scenerio above illustrates how borrowing, repayment, and interest accrual work with `InterestScalar`. Accruing interest and calculating total borrowed amounts of a token denom do not require iterating over all borrows, as long as `AdjustedBorrow` is used and `TotalAdjustedBorrow` is kept up to date on borrow and repay.

## Consequences

This design change should address our lingering tradeoff between performance and `InterestEpoch`

### Positive

- Total borrows and supply utilization can be calculated in O(1) time instead of O(N) as N is the total number of borrow positions across all users
- Periodic functions can now take place every block instead of every `InterestEpoch` blocks
- Quantities like uToken exchange rates and supply APYs now update instantly to new borrow and supply activity, even between multiple transactions within the same block.

### Negative

- The concepts of `AdjustedBorrow` and `InterestScalar` are introduced, increasing conceptual complexity

## References

The implementations discussed in some previous ADRs are partially superseded by this decision:

- [ADR-002: Borrow Assets](./ADR-002-borrow-assets.md)
- [ADR-004: Interest and Reserves](./ADR-004-interest-and-reserves.md)
