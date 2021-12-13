# ADR 005: Liquidation

## Changelog

- December 10, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

Debt positions on Umee always start out over-collateralized, but if sufficient interest accrues or asset prices fluctuate too quickly, some borrowers may reach a state where the total value of their collateral is less than the value of their borrowed assets.

Such debt (now under-collateralized) may become over-collateralized on its own if asset prices rebound in the right direction, but it is also eligible for liquidation in its current state.
When fully liquidated, accounts whose `BorrowedValue` is greater than `CollateralValue` (adjusted for liquidation incentive) will result in an account with nonzero total borrows but zero collateral, thus no incentive for repayment.

It is in the interests of the overall system to repay such _bad debt_ using reserves, in order to prevent it accruing interest forever and damaging the health of the lending pool.

## Alternatives

### Repayment Eligibility Condition

One consideration to make is whether borrowers should be eligible for bad debt repayment whenever `Collateral Value < Borrow Value`, or if eligibility should be delayed until `Collateral Value == 0`.

In the former case, the system would also confiscate any remaining collateral, as though it were a liquidator, and would have to process the collateral in a way of our choosing (e.g. exchange for base assets, and add them to reserves).

In the latter case, liquidators would first have liquidate the portion of the debt that can be exchanged for collateral, reducing the total burden on reserves when the borrow is being repaid.

Also in the latter case, if liquidators do not (for whatever reason) liquidate until the under-collateralized account's collateral is zero, its borrows would continue to accrue interest and remain ineligible for bad debt repayment.
This should not be a common case, as long as liquidation incentives exist, but might happen if liquidators are sparse, or if the remaining collateral is so small that it is not worth the gas fee to retrieve.

### Checking for Repayment Eligibility

A second consideration is when to check for borrowers' eligibility for bad debt repayment. The two likely options are immediately during `LiquidateBorrow`, or during `EndBlock` every `InterestEpoch`. A third option might be to check periodically, but on a separately controlled interval.

The advantage of an immediate check during `LiquidateBorrow` is that only the borrow being liquidated needs to be checked for eligibility, instead of periodically iterating over all borrows. Additionally, `CollateralValue` has already been calculated in that function.

However, there is an edge case (reserve exhaustion) where borrows eligible for reserve-driven-repayment during their final liquidation, cannot be fully repaid at that moment.
If using an immediate check during `LiquidateBorrow`, there would be no future liquidation against that address to trigger bad debt repayment after reserves recovered - the debt would accrue interest indefinitely.

Also note that the first decision on the condition for bad debt repayment eligibility changes the effectiveness of these options:

- A condition of `Collateral Value == 0` favors checks during `LiquidateBorrow`, because liquidation is the only action on under-collateralized accounts that is allowed to reduce collateral to zero.
- A condition of `Collateral Value < Borrow Value` mandates periodic checks, as such an inequality can result not only from liquidation, but also from interest accrual and/or asset price fluctuations.

## Decision

The simplest way to implement bad debt repayment is as follows:

At the end of `LiquidateBorrow`, which is triggered by `MsgLiquidate`, the `borrowerAddress` in question is checked for `Collateral Value == 0`. If true, a separate function `RepayBadDebt(borrowerAddress)` is called.

The function `RepayBadDebt(borrowerAddress)` immediately repays any remaining borrows on the address in question, or repays the maximum available amount for any borrowed asset denomination where the borrowed amount exceeds available reserves.

### Recovery from Exhausted Reserves

When `RepayBadDebt(borrowerAddress)` fails to repay a borrow in full due to insufficient reserves, it stores information about the borrower and denomination in question so repayment can be completed later.

```go
// pseudocode
// store a list of denominations with nonzero bad debt
badDebtDenomPrefix | tokenDenom = true
// store a list of borrow addresses with bad debt, sorted by denom
badDebtAddressPrefix | lengthPrefixed(tokenDenom) | borrowerAddress = true
```

Then, every `InterestEpoch`, bad debt positions can be iterated through and repaid as reserves become available. Any `denom|address` that is fully repaid is cleared from the list, as well any `denom` which no longer has any associated bad debt addresses.

Keeping a list of denominations in addition to addresses enables the iteration to efficiently skip denoms whose reserves are still zero, thus only reading addresses with bad debt that can actually be repaid at the time.

### Messages, Events, and Logs

No new message types are required for bad debt repayment, as it is automatic.

The function `RepayBadDebt` must create event `EventRepayBadDebt` and logs for each borrowed asset it repays, recording address, denom, and amount.

An additional event `ReservesExhausted` could be created to monitor for that specific edge case. It should record the `denom`, `borrowerAddress`, and remaining `BorrowedAmount` for any borrow that `RepayBadDebt` fails to repay.

### Edge Cases

There are two edge cases that would allow bad debt to slip past detection or automatic repayment:

- Nonzero collateral: If liquidators do not reduce borrower collateral to zero, even leaving a miniscule amount, bad debt will not be detected at the end of `LiquidateBorrow`.

This could be solved by any Liquidator - but if undercollateralized borrowers are being left for long periods without being completely liquidated, then a bot could be set up to finish them, even if that bot is controlled by Umee.

- Reserve exhaustion: If reserves are insufficient to fully repay bad debt at the moment `RepayBadDebt` is called, the remaining debt will not be targeted again (as with zero collateral, no `LiquidateBorrow` will occur in the future).

This edge case is handled by storing bad un-handled bad debt addresses using `BadDebtAddressPrefix`.

## Consequences

### Positive

- Bad debt is eligibe for reserve-funded repayment only when borrower collateral is zero, allowing liquidators to soften the blow to reserves
- Computation-efficient checks for `Collateral Value == 0` during `LiquidateBorrow` do not require iterating over all borrowers.
- System automatically recovers from reserve exhaustion

### Negative

- Bad debt with nonzero collateral goes undetected until the collateral is liquidated

### Neutral

- Events and logs track when bad debt is repaid by reserves
- Events and logs detect exhaustion of reserves and associated details