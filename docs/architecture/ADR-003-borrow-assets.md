# ADR 003: Borrowing assets using uToken collateral

## Changelog

- September 27, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

One of the base functions of the Umee universal capital facility is to allow users to borrow allowed asset types, using their own uTokens (obtained normally, by depositing assets) as collateral.

## Alternative Approaches

It might be worth considering adding message types which do two or more of the steps below at once (e.g. deposit base asset, convert to uToken collateral, and receive loan all at once). The individual steps to need to be possible in isolation, though.

## Decision

The Cosmos `x/bank` module and the existing `umee/x/leverage` deposit features are prerequisites for these new capabilities.

The flow of events is as follows:
- Borrower has already used `umee/x/leverage` module to deposit base assets and receive uTokens
- Borrower deposits uTokens into asset facility as collateral - facility remembers collateral position
- Borrower requests to borrow assets from capital facility - facility checks request, disburses tokens if acceptable, and remembers borrow position
- Eventually, borrower pays repayment amount (in full or in part) using borrowed asset denomination
- Borrower may withdraw collateral uTokens at any time as long as minimum overcollateralization requirement remains satisfied
- On asset repayment amount = 0 for a given borrow position, facility forgets loan.
- On uToken collateral amount = 0 for a given collateral position, facility forgets collateral.

Additionally, the following events occur periodically (at the end of every block):
- Repayment amounts for open borrow positions increases by a borrowed-token-specific interest rate
- Open borrow positions are checked to see if they fall below liquidation threshold

The `umee/x/leverage` module keeper will be responsible for remembering each open borrow or collateral position.
If the same user account opens multiple of the same type of position (e.g. borrow) in the same token denomination, the second position simply increases the amount of the first.

Additionally it has been discussed that, rather than an account's specific collateral positions being tied to specific borrow positions, the sum of all borrow and collateral positions related to an account is used to check the account's overcollateralization levels and mark its positions for liquidation.

Note that the exchange rate of Assets:uAssets (e.g. `uatom:u/uatom`) will be a shifting exchange rate that grows with interest - see ADR-001.

In contrast, the exchange rate of original assets:borrowed assets (e.g. `uatom:ether`) can only be determined using a price oracle.

The true borrow exchange rate used in the transaction (e.g. `u/uatom:ether`) is derived from the two above.

Note also that as a consequence of uToken interest, the asset value of uToken collateral increases over time, meaning a user who repays positions in full and redeems collateral uTokens will receive back more base assets than they deposited originally, reducing effective interest.

## Detailed Design

For the purposes of borrowing and repaying assets, as well as depositing and withdrawing uToken collateral, the `umee/x/leverage` module does not mint or burn tokens. It uses its module account to store collateral uTokens, its keeper to remember open positions, and the `x/bank` module to perform all necessary balance checks and token transfers.

### Basic Message Types

To implement the deposit functionality of the Asset Facility, the two common message types will be:
```go
// MsgDepositCollateral - a user wishes to deposit uToken collateral of one or more types
type MsgDepositCollateral struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coins      `json:"amount" yaml:"amount"`
}
// MsgWithdrawCollateral - a user wishes to withdraw uToken collateral of one or more types
type MsgWithdrawCollateral struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coins      `json:"amount" yaml:"amount"`
}
// MsgBorrowAsset - a user wishes to borrow assets of one or more allowed types, assuming collateral already deposited
type MsgBorrowAsset struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coins      `json:"amount" yaml:"amount"`
}
// MsgRepayAsset - a user wishes to repay assets of one or more borrowed types
type MsgRepayAsset struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coins      `json:"amount" yaml:"amount"`
}
```
Messages must use only whitelisted denominations. Collateral is always a uToken denomination, and assets are never uTokens.

Asset borrowing compares collateral uTokens associated with one base asset, to borrowed assets of a second base asset type. Price oracles must be used to compare the values of base assets (e.g. Atoms:Ether.)

_Note: The `Coins` type seen in the `Amount` fields can contain multiple token types. The messages above should fail if even one of the coin types fails or is invalid, rather than partially succeeding._

It is necessary that messages be signed by the borrower's account. Thus the method `GetSigners` should return the `Borrower` address for all message types above.

### Keeper additions

Using the `sdk.Coins` built-in type, which combines multiple {Denom,Amount} pairs as a single object, the `umee/x/leverage` keeper may roughly remember open borrow and collateral positions as follows:

```go
// pseudocode
// for each borrower address, combine all open positions of each type (borrower/collateral) into one sdk.Coins object:
keeper[collateralPrefix + borrowerAddress] = CollateralCoins
keeper[borrowPrefix + borrowerAddress] = BorrowedCoins
```

This will be accomplished by adding new prefixes and helper functions to `x/leverage/types/keys.go`, and using the proper codec to marshal `sdk.Coins` into bytes when storing them as values.

### APIs and Handlers
Both CLI and gRPC must be supported when sending the above message types, and all necessary handlers must be created in order to process and validate them as transactions.

### Testing

Assuming a placeholder token allow-list of at least two elements (e.g. `uumee`,`uatom`), and uTokens existing (e.g. `u/uumee`,`u/uatom`), an integration test can be created in which one user account sends a `MsgBorrowAsset` and a `MsgRepayAsset` of the appropriate token types.

## Open Questions
- See ADR-002 open questions on whitelisting asset types and uniquely identifying ibc/ assets regardless of ibc path.

## Consequences

### Positive
- uTokens used as collateral increase in base asset value in the same way that lend positions do. This counteracts borrow position interest.

### Negative

### Neutral
- Borrow feature relies on allow-list of token types
- Borrow feature relies on price oracles for base asset types

## References
