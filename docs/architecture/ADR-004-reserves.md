# ADR 004: Borrow interest implementation and reserves

## Changelog

- October 14, 2021: Initial Draft (@toteki)
- October 18, 2021: More Details on Interest (@toteki)

## Status

Proposed

## Context

Borrow positions on Umee accrue interest over time.
When interest accrues, the sum of all assets owed by all users increases for each borrowed token denomination. The amount of that increase serves to benefit lenders (by increasing the token:uToken exchange rate), and also to increase the amount of base assets the Umee system holds in reserve.

The mechanism by which interest is calculated, and then split between incentivizing lenders as per [ADR-001](./ADR-001-interest-stream.md) and reserves as defined in this ADR, will follow.


## Decision

Interest on all open borrow positions will be applied every X blocks during `EndBlock`, where X is a governance parameter called `BorrowInterestEpoch`.

Reserves are implemented as a per-token value attached to the `x/leverage` module account.  The module account also holds the lending pool of base assets. As an example, if the module account contains 2000 `uatom` and its `uatom ReserveAmount` is 500, then only 1500 of its `uatom` are available for borrowing or withdrawal.

A governance parameter `ReserveFactor` must be kept which specifies the portion of borrow interest that will be funnelled into reserves. This parameter will either be decided per token, or as a single value for all asset types.

> TODO: Decide on the above.

Both the reserved amount of a given token, and the token:uToken exchange rate, increase when interest is _accrued_, rather then when it is _repaid_.

## Detailed Design

As noted in [ADR-003](./ADR-003-borrow-assets.md), open borrow positions are stored in the`x/leverage` module with the keys `borrowPrefix | lengthPrefixed(borrowerAddress) | tokenDenom` and values of type `sdk.Int`.

When accruing interest, the borrowed amount (`sdk.Int`) must be increased for each open borrow position. The increase should be calculated as follows...

### Borrow Interest Epoch

> accrued interest = borrowed amount * (interest rate * time elapsed)

A governance parameter `BorrowInterestEpoch` with type `int64` will determine the number blocks to wait between interest calculations. Whenever the number of blocks created since the last interest accrual reaches the desired value, interest accrues on all borrow positions simultaneously during `EndBlock`:

```go
// EndBlock
// k is the x/leverage keeper and ctx contains information on the current block
if ctx.BlockHeight() >= k.GetLastInterestBlock() + k.GetBorrowInterestEpoch() {
    // unix times (int64 values, measured in seconds)
    secondsElapsed := ctx.BlockTime.Unix() - k.GetLastInterestTime()
    // accrue interest to all borrow positions at once
    err = k.AccrueAllInterest(secondsElapsed)
    // Store the current block's height and time in keeper
    k.SetLastInterestBlock(ctx.BlockHeight())
    k.SetLastInterestTime(ctx.BlockTime())
}
```

To support the function above, the `x/leverage` module keeper must store the last unix time and block height at which interest accrued. Both values will be marshaled to binary from `sdk.Int`. Single byte prefixes can be used:

```go
key = interestLastHeightPrefix // e.g. 0x06
key = interestLastTimePrefix // e.g. 0x07
```

### Dynamic Borrow Interest Rates

Borrow interest rates are dynamic. They are calculated using the lending pool's current borrow utilization for each asset type, as well as multiple governance parameters that are decided on a per-token basis. Even the model used for the interest rate (which determines what other governance parameters are required) must itself be governed.

Here are some example initial governance parameters which would be stored for a single asset type:
```go
InterestRateModel = 0x01 // enumeration. Which model (if multiple available) to use?
// Model 0x01, based on Compound's JumpRateModelV2, requires the following per token:
BaseAPY = sdk.NewDec("0.02")
KinkAPY = sdk.NewDec("0.2")
MaxAPY = sdk.NewDec("1.0")
KinkUtilization = sdk.NewDec("0.8")
```

Each parameter, being stored per token, would need a keeper prefix:

```go
// For the interest rate model enumeration, which always exists for each token
interestModelPrefix | denom | 0x00
// For other parameters, which only exist when the interest model is a certain value
interestParamPrefix | lengthPrefixed(denom) | paramNumber
// example paramNumbers
//  BaseInterest: 0x00
//  KinkInterest: 0x01
//  MaxInterest: 0x02
//  KinkUtilization: 0x03
```

The `0x01` interest model shown above, based on [Compound's JumpRateModelV2 model](https://compound.finance/governance/proposals/20) defined in [this contract](https://etherscan.io/address/0xfb564da37b41b2f6b6edcc3e56fbf523bd9f2012#code), can be summarized as follows:

> The (x,y) = (utilization, interest rate) graph is a line with a kink in it, defined by three points
> At 0% utilization, there is a base interest rate `BaseInterest`
> The kink at `KinkUtilization` utilization has interest rate `KinkInterest`
> At 100% utilization, the interest rate is `MaxInterest`

Because parameters can cease to be necessary when the interest model changes (e.g. if an interest model 0x02 only required two parameters instead of four), the governance process should delete unused parameters from the keeper when changing models if deemed necessary.

The `x/leverage` module keeper will contain a function which derives the current interest rate of an asset type:

```go
func (k Keeper) DeriveInterestRate(ctx sdk.Context, denom string) (sdk.Dec, error) {
    // Implementation must detect which InterestRateModel has been assigned to the denom,
    // then calculate the denom's borrowing utilization, and return the resulting
    // annual interest rate as a decimal.
}
```

### Accruing Interest

The `x/leverage` module keeper will contain a function which accrues interest on all open borrow positions at once, which is called when `EndBlock` detects that `BorrowInterestEpoch` has elapsed.

```go
func (k Keeper) AccrueAllInterest(ctx sdk.Context, secondsElapsed int64) error {
    // for each borrow, expressed as an sdk.Coin(Denom, Amount) associated with an sdk.Address
    {
        derivedInterest, err := k.DeriveInterestRate(denom)
        // derived interest is annual, so we must conver time to years for the math to work
        yearsElapsed := sdk.OneDec.MulInt64(secondsElapsed).QuoInt64(31536000) // seconds per year
        // accruedInterest = interest rate * borrow amount * time elapsed
        accruedInterest = derivedInterest.Mul(borrow.Amount).Mul(yearsElapsed)
        // then accrue calculated interest to the individual loan
        borrow.Amount = borrow.Amount.Add(accruedInterest.Ceil())
        err = k.UpdateLoan(ctx, borrowedAddress, borrow)
        // and then increase reserves (omitted)
    }
}
```

In the codebase, the function above will be written more efficiently by reusing yearsElapsed and remembering the derived APY for all borrows of the same asset type. Error handling and iterator details are also omitted for clarity here.

### Storing Reserves

The code snippet above also interacts with reserves. The general design of reserves follows:

The `x/leverage` module keeper must store, for each token denom, an `sdk.Int` representing the amount of that token which is reserved. The key to store the reserve amount per unique denomination is formed by:

```go
    // reserveAmountPrefix | denom | 0
    var key []byte
    key = append(key, KeyPrefixReserveAmount...)
    key = append(key, []byte(tokenDenom)...)
    return append(key, 0) // append 0 for null-termination
```

The portion of accrued interest set aside as reserves (an `sdk.Dec`) is determined per-token by a governance parameter called `ReserveFactor`. Its keeper prefix is constructed as follows:

```go
    // reserveFactorPrefix | denom | 0
    var key []byte
    key = append(key, KeyPrefixReserveFactor...)
    key = append(key, []byte(tokenDenom)...)
    return append(key, 0) // append 0 for null-termination
```

Reserves are part of the module account's balance, but may not leave module account as the result of `MsgBorrowAsset` or `MsgWithdrawAsset`. Only governance actions outside the scope of this ADR may release or transfer reserves.

The function to increase reserves of a denomination, after accrued interest on a borrow has already been

To increase reserves whenever interest is accrued, the following must be added inside function `AccrueAllInterest`:

```go
// Initially
newReserves := sdk.NewCoins()
// ...
// for each borrow, after UpdateLoan
increase := accruedInterest.Mul(k.GetReserveFactor(denom)).Ciel()
newReserves, err = newReserves.Add(sdk.NewCoin(denom,increase))
// ...
// and after the for loop
for _, coin := range newReserves {
    amount, err := k.GetReserveAmount(ctx, denom)
    amount = amount.Add(coin.Amount)
    err = k.SetReserveAmount(ctx, denom, amount)
}
```

### uToken Exchange Rate Impact

In financial terms, accrued interest is split between reserves and lender profits. Because reserve amounts are set explicitly in code but the uToken exchange rate is implicit, the uToken rate calculation depends on reserved amounts in practice:

The token:uToken exchange rate, which would have previously calculated by `(total tokens borrowed + module account balance) / uTokens in circulation` for a given token type and associated uToken denomination, must now be computed as `(total tokens borrowed + module account balance - reserved amount) / uTokens in circulation`.

### Message Types

No new message types (outside of governance of the many reserve parameters) are required to implement dynamic interest rate and reserve functionality.

### Modifications to Withdraw and Borrow

Existing functionality will be modified:

- Asset withdrawal (by lenders) will not permit withdrawals that would reduce the module account's balance of a token to below the reserved amount of that token.
- Asset borrowing (by borrowers) will not permit borrows that would do the same.

### Example Scenarios

Example scenario:

> The global `ReserveFactor` paramater is 0.05
>
> The derived interest rate for `atom` is assumed to be 0.0001% (10^-6) per `BorrowInterestEpoch` blocks.
>
> Alice has borrowed 2000 `atom`, which is 2 * 10^9 `uatom` internally. (Note: These are not utokens, just the smallest units of the token. 1 atom = 10^6 uatom.)
>
> On the next `EndBlock` where `BorrowInterestEpoch` has passed, interest is accrued on the borrow. Using the example interest rate in this example, Alice's amount owed goes from 2 * 10^9 `uatom` to 2.000002. In other words, it increases by 2000 `uatom`.
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
> On the next `EndBlock` after `BorrowInterestEpoch` has passed, interest accrues. Alice now owes 1000.001 `atom` (the amount owed increased by 1000 `uatom`). The amount of `uatom` the module is required to reserve increases from 0 to 50 (assuming the `ReserveFactor` parameter is 0.05 like in the previous example.)
>
> The module account (lending pool + reserves) still contains 0 `uatom` due to the first two steps. Its `uatom` balance is therefore less than the required 50 `uatom` reserves.

The scenario above is not fatal to our model - lender Bob continues to gain value as the token:uToken exchange rate increases, and we are not storing any negative numbers in place of balances - but the next 50 `uatom` lent by a lender or repaid by a borrower would serve to meet the reserve requirements rather than being immediately available for borrowing or withdrawal.

The edge case above can only occur when the available lending pool (i.e. module account balance minus reserve requirements) for a specific token denomination, is less than `ReserveFactor` times the interest accrued on all open loans for that token in a single block. In practical terms, that means ~100% borrow utilization.

This is not a threatening scenario, as it resolves as soon as either a sufficent `RepayAsset1` or a `LendAsset` is made in the relevant asset type, both of which are **highly** incentivized by the extreme dynamic interest rates found near 100% utilization.

## Consequences

### Positive

- Existing module account can be used for reserves
- Borrow and withdraw restrictions are straightforward to calculate (module balance - reserve amount)
- No tranactions or token transfers are required when accrued interest contributes to reserves

### Negative

- Reserved amounts can temporarily exceed module account balance in some cases (when interest accrues on a token type at or very near 100% borrowing utilization)

### Neutral

- Requires governance parameter `ReserveFactor` defining the portion of interest that must go to reserves
- Requires governance parameter `BorrowInterestEpoch` defining how many blocks to wait between interest calculations
- Requires governance paramater `InterestRateModel` and multiple model-specific parameters to calculate dynamic interest rates.
- Asset reserve amounts are recorded directly by the `x/leverage` module

## References
