# ADR 008: Borrow Tracking

## Changelog

- Jan 27, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

One of the more computationally expensive operations in Umee is "iterate over all borrows". It is currently necessary on several occasions:

- When calculating borrow utilization (require's a denomination's total borrowed)
- When accruing interest (dynamic interest requires borrow utilization, and all borrows must be modified)
- When deriving uToken exchange rates (require's a denomination's total borrowed)

In order to reduce the performance impact of such iteration, `InterestEpoch` was created, so that such calculations only occurred every N blocks (e.g. 100).

This design proposal seeks to modify how borrows are stored, in order to achieve the same functionality without ever iterating over all borrows. As a result, `InterestEpoch` can also be eliminated, effectively happening every block.

## Decision

Chain state will now store an `InterestScalar` for each accepted asset denom, which starts at `1.0` and increases multiplicatively every time interest would accrue on borrows of that denom.

Instead of directly storing `sdk.Int BorrowedAmount` for each of a user's borrowed denoms, state records their `sdk.Dec AdjustedBorrow`, respecting the following definition:

> `AdjustedBorrow(denom,user)` * `InterestScalar(denom)` = `BorrowedAmount(denom,user)`

State will additionally store a value `TotalAdjustedBorrows` for each denom, which tracks the sum of all users' adjusted borrows in the chosen denomination.
It must increase or decrease every time assets are borrowed or repaid (including bad debt repayment), but is not modified when interest accrues.

This change in borrow storage will be opaque to other parts of the leverage module: functions like `GetBorrow`, `GetTotalBorrows`, and `SetBorrow` will use `InterestScalar` internally, such that their inputs and outputs contain the real borrowed amount.

As a result of efficiency gains, the parameter `InterestEpoch` will be removed, with periodic functions taking place every block.

Additionally, the values `BorrowAPY`, `LendAPY`, and `uTokenExchangeRate` will be removed from state, instead being efficiently calculated when needed.

## Detailed Design

This decision mainly updates existing features, rather than adding new ones. The following changes are required to update the leverage module:

**Params:**
- Remove `InterestEpoch`

**State:**
- Add `InterestScalar` and `TotalAdjustedBorrows` to state, and add `Get/Set` functions 
- Remove `BorrowAPY`, `LendAPY`, and `uTokenExchangeRate` from state, and remove `Set` functions

**Getters:**
- Modify `Get` functions for `BorrowAPY`, `LendAPY`, and `uTokenExchangeRate` to calculate values in real time.
- Modify `GetBorrow(addr,denom)`, `GetBorrowerBorrows(addr)`, and `GetAllBorrows()` to use `InterestScalar` and `AdjustedBorrow`
- Modify `GetTotalBorrows(denom)` to use `InterestScalar` and `TotalAdjustedBorrows`

**Setters:**
- Modify `SetBorrow(addr,denom)` to use `InterestScalar` and a stored `AdjustedBorrow`, and to additionally update `TotalAdjustedBorrows` based on the difference from the previous value (calls get before set.)

**Genesis:**
- Rename the `Borrow` struct in genesis state to `AdjustedBorrow`, and have it store the new `sdk.Dec`

**Invariants:**
- Add invariant which checks `TotalAdjustedBorrows` against the total of all `AdjustedBorrow` returned by `GetAllBorrows()`

## Consequences

This design change should address our lingering tradeoff between performance and `InterestEpoch`

### Positive
- Borrow totals and borrow utilization can be calculated in O(1) time instead of O(N) as N is the total number of borrow positions across all users
- Periodic functions can now take place every block instead of every `InterestEpoch` blocks
- Quantities like uToken exchange rates and lend APYs now update instantly to new borrow and lend activity, even between multiple transactions within the same block.

### Negative
- The concepts of `AdjustedBorrow` and `InterestScalar` are introduced, increasing conceptual complexity

## References

The implementations discussed in some previous ADRs are partially superseded by this decision:
- [ADR-002: Borrow Assets](./ADR-002-borrow-assets.md)
- [ADR-004: Interest and Reserves](./ADR-004-interest-and-reserves.md)
