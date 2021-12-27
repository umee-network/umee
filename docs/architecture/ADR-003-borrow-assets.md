# ADR 003: Borrowing assets using uToken collateral

## Changelog

- September 27, 2021: Initial Draft (@toteki)
- September 29, 2021: Changed design after review suggestions (@toteki, @alexanderbez, @brentxu)
- October 5, 2021: MsgSetCollateral and borrower-address-prefixed store keys (@toteki)
- December 16th, 2021: Collateral storage updated to use module account

## Status

Proposed

## Context

One of the base functions of the Umee universal capital facility is to allow users to borrow allowed asset types, using their own uTokens (obtained normally, by depositing assets) as collateral.

## Alternative Approaches

- Allow amounts of uToken to be specifically marked as collateral, rather than toggling collateral on/off for each asset type. This would allow more fine-grained control of collateral by borrowers.

## Decision

The Cosmos `x/bank` module and the existing `umee/x/leverage` deposit features are prerequisites for these new capabilities.

The flow of events is as follows:
- Borrower already has uTokens in their account
- Borrower marks those uTokens as eligible for use as collateral, which stores them in the `x/leverage` module account
- Borrower requests to borrow assets from `leverage` module - module checks request, disburses tokens if acceptable, and remembers borrow position
- While the borrow position is open, transactions that would result in the ongoing account's borrow position being higher than its calculated borrow limit are prevented (i.e. borrowing too much, withdrawing too many uTokens that are being used as collateral, disabling essential collateral)
- Eventually, the borrower pays the repayment amount (in full or in part) in the same asset denomination that was borrowed.

Additionally, the following events occur at `EndBlock`:
- Repayment amounts for open borrow positions increases by a borrowed-token-specific interest rate

The `umee/x/leverage` module will be responsible for remembering each open borrow position.
If the same user account opens multiple borrow positions in the same token denomination, the second position simply increases the amount of the first.

Additionally, rather than an account's specific uToken denoms being tied to specific borrow positions, the sum of all borrow positions and collateral uTokens related to an account is to calculate the account's borrow limit.

Note that the exchange rate of Assets:uAssets (e.g. `uatom:u/uatom`) is still a shifting exchange rate that grows with interest - see [ADR-001: Interest Stream](./ADR-001-interest-stream.md).

In contrast, the exchange rate of original assets:borrowed assets (e.g. `uatom:ether`) can only be determined using a price oracle.

The calculated borrow limit, which weighs collateral uTokens against borrowed assets (e.g. `u/uatom:ether`) is derived from combining the two above.

Note also that as a consequence of uToken interest, the asset value of uToken collateral increases over time, meaning a user who repays positions in full and redeems collateral uTokens will receive back more base assets than they deposited originally, reducing the effective interest.

## Detailed Design

For the purposes of borrowing and repaying assets, as well as marking uTokens as collateral, the `umee/x/leverage` module does not mint or burn tokens. It uses its keeper to remember open positions and user collateral settings, and the `x/bank` module to perform all necessary balance checks and token transfers.

To prevent users from transferring uTokens when doing so would lower their borrow limit below their current total borrowed amount, uTokens that are enabled as collateral are stored in the `x/leverage` module account, and moved back to the user's wallet when they are disabled as collateral.
As long as it would not reduce a user's borrow limit below their borrowed value, `MsgWithdrawAsset` can withdraw base assets in exchange uTokens used as collateral, even when they are stored in the module account.

Because borrow limits weight the value of different token denominations together, the calculation will require the use of `x/oracle`.

### Basic Message Types

To implement the borrow/repay functionality of the Asset Facility, the three common message types will be:

```go
// MsgSetCollateral - a borrower enables or disables a specific utoken type in their wallet to be used as collateral
type MsgSetCollateral struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Denom    string         `json:"denom" yaml:"denom"`
  Enable   bool           `json:"enable" yaml:"enable"`
}
// MsgBorrowAsset - a user wishes to borrow assets of an allowed type
type MsgBorrowAsset struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coin       `json:"amount" yaml:"amount"`
}
// MsgRepayAsset - a user wishes to repay assets of a borrowed type
type MsgRepayAsset struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coin       `json:"amount" yaml:"amount"`
}
```
Messages must use denominations only in the allow-list. Collateral is always a uToken denomination, and assets are never uTokens.

Asset borrowing respects the user's borrowing limit, which compares the total value of a borrower's collateral (scaled by each token's `CollateralWeight`) with the borrower's total borrowed value. Price `x/oracle` must be used to calculate borrow and collateral values.

It is necessary that messages be signed by the borrower's account. Thus the method `GetSigners` should return the `Borrower` address for all message types above.

### Module State (Keeper) Additions

Using the `sdk.Coins` built-in type, which combines multiple {Denom,Amount} pairs as a single object, the `umee/x/leverage` module may remember open borrow and collateral positions as follows:

```go
// pseudocode
// for each borrower address store open borrows
borrowPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = sdk.Int
// additionally borrower collateral settings are stored for enabled denoms
collateralSettingPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = true/false
// and the amount of collateral stored for each denom is kept
collateralAmountPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = sdk.Int
```

This will be accomplished by adding new prefixes and helper functions to `x/leverage/types/keys.go`.

The use of borrowerAddress before tokenDenom in the store keys allows the KVstore's "iterate by prefix" functionality to perform operations on all keys of a type relating to a specific address, e.g. "all open borrow positions belonging to an individual user". The same applies to collateral settings and amounts.

In contrast, if we had put tokenDenom before borrower address, it would favor operations on the set of all keys associated with a given token.

### APIs and Handlers
Both CLI and gRPC must be supported when sending the above message types, and all necessary handlers must be created in order to process and validate them as transactions.

## Consequences

### Positive

- uTokens used as a collateral increase in base asset value in the same way that lends positions do. This counteracts borrow position interest.
- UX of enabling/disabling token types as collateral is simpler than depositing specific amounts
- `lengthPrefixed(borrowerAddress) | tokenDenom` key pattern facilitates getting open borrow and collateral positions by account address.

### Negative

- `lengthPrefixed(borrowerAddress) | tokenDenom` key pattern makes it more difficult to get all open borrow positions by token denomination.

### Neutral
- Borrow feature relies on an allow-list of token types
- Borrow feature relies on price oracles for base asset types
- Borrow interest will rely on dynamic rate calculation

## References
