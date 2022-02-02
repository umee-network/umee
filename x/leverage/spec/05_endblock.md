# End Block

Every block, the leverage module runs the following steps in order:
- Repay bad debts using reserves
- Accrue interest on borrows

## Sweep Bad Debt

Borrowers whose entire balance of collateral has been liquidated but still owe debt are marked by their final liquidation transaction. This periodic routine sweeps up all marked `address | denom` bad debt entries in the keeper, performing the following steps for each:

- Determine the about of [Reserves](01_concepts.md#Reserves) in the borrowed denomination available to repay the debt
- Repay the full amount owed using reserves, or the maxmimum amount available if reserves are insufficient
- Emit a "Bad Debt Repaid" event indicating amount repaid, if nonzero
- Emit a "Reserves Exhausted" event with the borrow amount remaining, if nonzero

## Accrue Interest

At every epoch, the module recalculates [Borrow APY](01_concepts.md#Borrow-APY) and [Lending APY](01_concepts.md#Lending-APY) for each accepted asset type, storing them in state for easier query.

Borrow APY is then used to accrue interest on all open borrows.

After interest accrues, a portion of the amount for each denom is added to the state's `ReservedAmount` of each borrowed denomination.

Then, an additional portion of interest accrued is transferred from the `leverage` module account to the `oracle` module to fund its reward pool.