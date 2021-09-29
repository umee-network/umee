# ADR 003: Borrowing assets using uToken collateral

## Changelog

- September 27, 2021: Initial Draft (@toteki)
- September 29, 2021: Changed design after review suggestions (@toteki, @alexanderbez, @brentxu)

## Status

Proposed

## Context

One of the base functions of the Umee universal capital facility is to allow users to borrow allowed asset types, using their own uTokens (obtained normally, by depositing assets) as collateral.

## Alternative Approaches

- Borrower deposits uTokens into module account when using as collateral, instead of keeping in user wallet, flagging as collateral-enabled, and restricting transfer. We would not longer have to extend the `x/bank` module, but there would be additional message types.
- Allow amounts of uToken to be specifically marked as collateral, rather than toggling collateral on/off for each asset type. This can be used with or without the other change above, and would allow more fine-grained control of collateral by borrowers.

## Decision

The Cosmos `x/bank` module and the existing `umee/x/leverage` deposit features are prerequisites for these new capabilities.

The flow of events is as follows:
- Borrower already has uTokens in their account
- Borrower marks those uTokens as eligible for use as collateral
- Borrower deposits uTokens into asset facility as collateral - module remembers collateral position
- Borrower requests to borrow assets from `leverage` module - module checks request, disburses tokens if acceptable, and remembers borrow position
- While borrow position is open, modified `x/bank` module prevents transactions that would result in the ownging account's borrow position being higher than its calculated borrow limit (e.g. sending too many uTokens that are being used as collateral)
- Eventually, borrower pays repayment amount (in full or in part) in the same asset denomination that was borrowed.

Additionally, the following events occur at `EndBlock`:
- Repayment amounts for open borrow positions increases by a borrowed-token-specific interest rate
- Open borrow positions are checked to see if they fall below liquidation threshold

The `umee/x/leverage` module keeper will be responsible for remembering each open borrow position.
If the same user account opens multiple borrow positions in the same token denomination, the second position simply increases the amount of the first.

Additionally it has been discussed that, rather than an account's specific uToken denoms being tied to specific borrow positions, the sum of all borrow positions and collateral uTokens related to an account is used to check the account's overcollateralization levels and mark its positions for liquidation.

Note that the exchange rate of Assets:uAssets (e.g. `uatom:u/uatom`) is still a shifting exchange rate that grows with interest - see [ADR-001: Interest Stream](./ADR-001-interest-stream.md).

In contrast, the exchange rate of original assets:borrowed assets (e.g. `uatom:ether`) can only be determined using a price oracle.

The calculated borrow limit, which weighs collateral uTokens against borrowed assets (e.g. `u/uatom:ether`) is derived from combining the two above.

Note also that as a consequence of uToken interest, the asset value of uToken collateral increases over time, meaning a user who repays positions in full and redeems collateral uTokens will receive back more base assets than they deposited originally, reducing effective interest.

## Detailed Design

For the purposes of borrowing and repaying assets, as well as marking uTokens as collateral, the `umee/x/leverage` module does not mint or burn tokens. It uses its keeper to remember open positions and user collateral settings, and the `x/bank` module to perform all necessary balance checks and token transfers.

Additionally, the `x/bank` module must be extended to prevent users from transferring uTokens when doing so would lower their borrow limit below thir current total borrowed amount.

Because borrow limits weight the value of different token types together, the calculation will require the use of price oracles (and placeholder functions before those are available).

### Basic Message Types

To implement the borrow/repay/collateral functionality of the Asset Facility, the four common message types will be:
```go
// MsgSetCollateral - a borrower enables or disables a specific utoken type in their wallet to be used as collateral
type MsgSetCollateral struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Denom    string         `json:"denom" yaml:"denom"`
  Enable   bool           `json:"enable" yaml:"enable"`
}
// MsgBorrowAsset - a user wishes to borrow assets of one or more allowed types
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

Asset borrowing respects the user's borrowing limit, which compares collateral uTokens associated with one base asset, to borrowed assets of a second base asset type. Price oracles must be used to compare the values of base assets (e.g. Atoms:Ether.)

_Note: The `Coins` type seen in the `Amount` fields can contain multiple token types. The messages above should fail if even one of the coin types fails or is invalid, rather than partially succeeding._

It is necessary that messages be signed by the borrower's account. Thus the method `GetSigners` should return the `Borrower` address for all message types above.

### Keeper additions

Using the `sdk.Coins` built-in type, which combines multiple {Denom,Amount} pairs as a single object, the `umee/x/leverage` keeper may roughly remember open borrow and collateral positions as follows:

```go
// pseudocode
// for each borrower address, combine all open borrow positions into one sdk.Coins object:
keeper[borrowPrefix + borrowerAddress] = BorrowedCoins
// additionally user collateral settings are stored as a slice of enabled denoms
keeper[collateralPrefix + borrowerAddress] = []string{denom1,denom2...}
```

This will be accomplished by adding new prefixes and helper functions to `x/leverage/types/keys.go`, and using the proper codec to marshal `sdk.Coins` and `[]string` into bytes when storing them as values.

### APIs and Handlers
Both CLI and gRPC must be supported when sending the above message types, and all necessary handlers must be created in order to process and validate them as transactions.

### Testing

Assuming a placeholder token allow-list of at least two elements (e.g. `uumee`,`uatom`), and uTokens existing (e.g. `u/uumee`,`u/uatom`), an integration test can be created in which one user account sends a `MsgBorrowAsset` and a `MsgRepayAsset` of the appropriate token types. It is also possible to use a single asset type for both the collateral and the borrowed asset.

## Open Questions
- See ADR-002 open questions on whitelisting asset types and uniquely identifying ibc/ assets regardless of ibc path.

## Consequences

### Positive
- uTokens used as collateral increase in base asset value in the same way that lend positions do. This counteracts borrow position interest.
- UX of enabling/disabling token types as collateral is simpler than depositing specific amounts

### Negative
- `x/bank` module must extend to prohibit peer-to-peer transfers of uTokens when they would violate `x/leverage` borrowing limit.

### Neutral
- Borrow feature relies on allow-list of token types
- Borrow feature relies on price oracles for base asset types
- Borrow interest will rely on not-yet-implemented dynamic rate calculation

## References
