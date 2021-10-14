# ADR 004: Accumulating reserves from borrow interest

## Changelog

- October 14, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

Borrow positions on Umee accrue interest over time.
When interest accrues, the sum of all assets owed by all users increases for each borrowed token denomination, and the amount of that increase serves to benefit lenders (by increasing the total amount of assets owed by the Umee system to lenders), and also to increase the amount of base assets the Umee system holds in reserve.

The mechanism by which interest is calculated, and then split between incentivizing lenders as per [ADR-001](./ADR-001-interest-stream.md) and reserves as defined in this ADR, will follow.

## Alternative Approaches

- Setting aside a separate account (or even separate module account, for example by making `x/reserve`) would separate reserve balances from the lending pool.

## Decision

The `x/leverage` module already has a module account capable of storing base asset types, and it currently does so for assets in the lending pool.

Adding reserve functionality should be as simple as using the existing module account to store reserve assets, and tracking (for each stored denomination) the amount which cannot be moved from the module acount (by means of a borrower's Borrow or a lender's Withdraw) because it is reserved.

The aforementioned restrictions on borrows and withdrawals must then be implemented.

Additionally, a governance parameter must be kept which specifies the portion of borrow interest that will be funnelled into reserves. This parameter will either be decided per token, or as a single parameter for all asset types.

> TODO: Decide on the above.

Regarding the actual increasing of reserves when interest accrues on borrow positions, the following approach can be used (which interacts with the lender interest stream defined in [ADR-001](./ADR-001-interest-stream.md).)

For each token type:

- The `x/leverage` has a non-negative balance of the token in the module account
- There is a non-negative sum of all open borrow positions using that token type, which can be calculated by the `x/leverage` module.
- The `x/leverage` module records a non-negative amount of the token which is declared to be reserved.
- There is also an associated uToken type, of which the total amount in circulation can be measured by the `x/bank` module.

Then:

- For every open borrow of every token type, interest accrues during EndBlock using a dynamically calculated interest rate.
- At the same moment, a portion of that increase is applied to the amount of the token that is declared to be reserved.
- The token:uToken exchange rate, which would have previously calculated by `(total tokens borrowed + module account balance) / uTokens in circulation` for a given token type and associated uToken denomination, must now be computed as `(total tokens borrowed + module account balance - reserved amount) / uTokens in circulation`.

## Detailed Design

The `x/leverage` module account can be used without modification to store reserves.

The `x/leverage` module keeper must store, for each token denom, store an `sdk.Int` representing the amount of of that token which is reserved (i.e. may not be borrowed or withdrawn from the module account.) The keeper will work as follows:

```go
// pseudocode
reservePrefix | tokenDenom = sdk.Int
```

The logic for increasing the reserved amount of each token will occur in the EndBlock function, simultaneous with interest accrual:

```go
// pseudocode
for each borrow {
    denom = borrow.Denom
    interest := borrow.Amount * DeriveInterestRate(denom) * time // the short time between blocks
    borrow.Amount += interest
    keeper.ReserveAmount[denom] += interest * params.ReserveFactor // e.g. 5% of interest goes to reserves
}
```

No new message types (outside of governance of the reserve parameter) are required to implement reserve functionality.

Existing functionality will be modified:

- Asset withdrawal (by lenders) will not permit withdrawals that would reduce the module account's balance of a token to below the reserved amount of that token.
- Asset borrowing (by borrowers) will not permit borrows that would do the same.

Example scenario:

> The global `ReserveFactor` paramater is 0.05
>
> The derived interest rate for `atom` is assumed to be 0.0001% per block time (10^-6).
>
> Alice has borrowed 2000 `atom`, which is 2 * 10^9 `uatom` internally. (Note: These are not utokens, just the smallest units of the token. 1 atom = 10^6 uatom.)
>
> On the next `EndBlock`, interest is accrued on the borrow. Using the exampke interest rate in this example, Alice's amount owed goes from 2 * 10^9 `uatom` to 2.000002 * 10^9. In other words, it increases by 2000 `uatom`.
>
> Because of the `ReserveFactor` of 0.05, the `x/leverage` module's reserved amount of `uatom` increases by 100 (which is 2000 * 0.05) due to the interest accruing on Alice's loan.
>
> The same math is applied to all open borrows, which may be from different users and in different asset types, during EndBlock.

Note that it is the "amount reserved" by the module that is increasing, not the actual balance of  the `x/leverage` account. The amount reserved is simply the amount below which the module account cannot be allowed to drop as the result of borrows or withdrawals.

Here is an additional example scenario, to illustrate that the module account balance of a given token _can_ become less than the reserved amount, when a token type is at or near 100% borrow utilization:

> Lending pool and reserve amount of `atom` both start at zero.
>
> Bob, a lender, lends 1000 `atom` to the lending pool.
>
> Alice immediately borrows all 1000 `atom`
>
> On EndBlock, interest accrues. Alice now owes 1000.001 `atom` (the amount owed increased by 1000 `uatom`). The amount of `uatom` the module is required to reserve increases from 0 to 50 (assuming the `ReserveFactor` parameter is 0.05 like in the previous example.)
>
> The module account (lending pool + reserves) still contains 0 `uatom` due to the first two steps. Its `uatom` balance is therefore less than the required 50 `uatom` reserves.

The scenario above is not fatal to our model - lender Bob continues to gain value as the token:uToken exchange rate increases, and we are not storing any negative numbers in place of balances - but the next 50 `uatom` lent by a lender or repaid by a borrower would serve to meet the reserve requirements rather than being immediately available for borrowing or withdrawal.

The edge case above can only occur when the available lending pool (i.e. module account balance minus reserve requirements) for a specific token denomination, is less than `ReserveFactor` times the interest accrued on all open loans for that token in a single block. In practical terms, that means ~100% borrow utilization.

## Consequences

### Positive

- Existing module account can be used for reserves
- Borrow and withdraw restrictions are straightforward to calculate (module balance - reserve amount)
- No tranactions or token transfers are required when accrued interest contributes to reserves

### Negative

- Reserved amounts can temporarily exceed module account balance in some cases (when interest accrues on a token type at or very near 100% borrowing utilization)

### Neutral

- Requires governance parameter defining the portion of interest that must go to reserves
- Asset reserve amounts are recorded directly by the `x/leverage` module
- token:uToken exchange rate remains a derived value (calculated from module account balances, all outstanding loans, total utokens in circulation, and reserve amounts, for any given token).

## References
