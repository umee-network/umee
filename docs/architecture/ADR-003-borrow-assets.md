# ADR 003: Borrowing assets using uToken collateral

## Changelog

- September 27, 2021: Initial Draft (@toteki)
- September 29, 2021: Changed design after review suggestions (@toteki, @alexanderbez, @brentxu)
- October 5, 2021: MsgSetCollateral and borrower-address-prefixed store keys (@toteki)
- December 16, 2021: Collateral storage updated to use module account
- May 20, 2022: Introduce collateral max utilization (@robert-zaremba)

## Status

Accepted

## Context

One of the base functions of the Umee universal capital facility is to allow users to borrow allowed asset types, using their own uTokens (obtained normally, by depositing assets) as a collateral.

## Decision

The Cosmos `x/bank` module and the existing `umee/x/leverage` deposit features are prerequisites for these new capabilities.

The flow of events is as follows:

- Borrower already has uTokens in their account.
- Borrower marks his uTokens as eligible for collateral using the `x/leverage` module account.
- Borrower requests to borrow assets from `x/leverage` module. The module validates the request, disburses tokens if acceptable, and stores borrow position.
- While the borrow position is open, transactions that would result the borrow position being higher than its calculated borrow limit are prevented (i.e. borrowing too much, withdrawing too many uTokens that are being used as collateral, disabling collateral).
- Eventually, the borrower repays the borrowed position (in full or in part).

Additionally, the following events occur at `EndBlock`:

- Fess are added to the open borrow positions based on token-specific interest rate.

The `umee/x/leverage` module stores each open borrow position.
If the same user account opens multiple borrow positions in the same token, the second position simply increases the amount of the first.

Additionally, rather than segregating each borrow position with a specific collateral deposit (uToken coins) we aggregate them. The sum of all account's collateral uTokens related is used to calculate the account's borrow limit.
We define a **borrow limit** invariant:
\__sum of account's borrow positions must be smaller than the account borrow limit_.

Note that the exchange rate of Asset:u-Asset has a dynamic exchange rate that grows with accruing interest - see [ADR-001: Interest Stream](./ADR-001-interest-stream.md).

In contrast, the exchange rate of collateral:borrowed assets (e.g. `atom:ether`) can only be determined using a price oracle.

The calculated borrow limit, which weighs collateral uTokens against borrowed assets (e.g. `u/atom:ether`) is derived from combining the two above. The weight of each uToken is defined as `CollateralWeight`.

Note also that as a consequence of uToken interest, the asset value of uToken collateral increases over time, meaning a user who repays positions in full and redeems collateral uTokens will receive back more base assets than they deposited originally, reducing the effective interest.

## Detailed Design

For the purposes of borrowing and repaying assets, as well as marking uTokens as collateral, the `umee/x/leverage` module does not mint or burn tokens. It stores borrower open positions and collateral settings, and the `x/bank` module to perform all necessary balance checks and token transfers. User collateral (uTokens) are deposited in `x/leverage` module and withdrawn back to the user `x/bank` account balance when the user disables uTokens as a collateral.

During every operation which involves borrow position or collateral we check that the _borrow limit_ invariant holds.

`x/oracle` module is used to provide exchange rates of tokens to calculate borrow limit.

### API

To implement the borrow/repay functionality of the Asset Facility, the three message types are defined:

```go
// MsgSetCollateral - a borrower enables or disables a specific uToken type in their wallet to be used as collateral
type MsgSetCollateral struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  // not a uToken
  Denom    string         `json:"denom" yaml:"denom"`
  Enable   bool           `json:"enable" yaml:"enable"`
}

// MsgBorrowAsset - a user wishes to borrow assets of an allowed type
type MsgBorrowAsset struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  // not a uToken
  Amount   sdk.Coin       `json:"amount" yaml:"amount"`
}

// MsgRepayAsset - a user wishes to repay assets of a borrowed type
type MsgRepayAsset struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  // uToken
  Amount   sdk.Coin       `json:"amount" yaml:"amount"`
}
```

Tokens used in above messages must belong to the allow-list. Collateral must be a uToken.

Messages must be signed by the borrower's account.

Both CLI and gRPC must be supported for the above messages.

### Storage layout

Using the `sdk.Coins` built-in type, which combines multiple {Denom,Amount} pairs as a single object, the `umee/x/leverage` module stores open borrow and collateral positions as follows:

```go
// open borrows:
borrowPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = sdk.Int

// borrower collateral settings for enabled denoms:
collateralSettingPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = true/false

// and the amount of collateral deposited for each uToken:
collateralAmountPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = sdk.Int
```

This will be accomplished by adding new prefixes and helper functions to `x/leverage/types/keys.go`.

The use of borrowerAddress before tokenDenom in the store keys allows to "iterate by borrower" functionality, e.g. "all open borrow positions belonging to an individual user". The same applies to collateral settings and amounts.

In contrast, if we had put tokenDenom before borrower address, it would favor operations on the set of all keys associated with a given token.

## Alternative Approaches

- Allow amounts of uToken to be specifically marked as collateral, rather than toggling collateral on/off for each asset type. This would allow more fine-grained control of collateral by borrowers.

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
