# End Block

At the end of every block, the leverage module checks whether the current block height is a multiple of `InterestEpoch`.

Every `InterestEpoch`, it runs the following steps in order:
- Repay bad debts using reserves
- Update borrow and lend APY and accrue interest on borrows
- Update uToken exchange rates

## Sweep Bad Debt

Borrowers whose entire balance of collateral has been liquidated but still owe debt are marked by their final liquidation transaction. This periodic routine sweeps up all marked `address | denom` bad debt entries in the keeper, performing the following steps for each:

- Determine the about of [Reserves](01_concepts.md#Reserves) in the borrowed denomination available to repay the debt
- Repay the full amount owed using reserves, or the maxmimum amount available if reserves are insufficient
- Emit a "Bad Debt Repaid" event indicating amount repaid, if nonzero
- Emit a "Reserves Exhausted" event with the borrow amount remaining, if nonzero

## Accrue Interest

At every epoch, the module recalculates [Borrow APY](01_concepts.md#Borrow-APY) and [Lending APY](01_concepts.md#Lending-APY) for each accepted asset type, storing them in state for easier query.

Borrow APY is then used to accrue interest on all open borrows.

## Update Exchange Rates

Because [uToken Exchange Rates](01_concepts.md#uToken-Exchange-Rate) only change with interest accrual, they are reculculated and stored every epoch for each accepted asset type.