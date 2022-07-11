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

- Borrower already has uTokens in their balance.
- Borrower marks his uTokens as collateral, which transfers them to the `x/leverage` module account.
- Borrower requests to borrow assets from `x/leverage` module. The module validates the request, disburses tokens if acceptable, and stores borrow position.
- While the borrow position is open, transactions that would result the borrow position being higher than its calculated borrow limit are prevented (i.e. borrowing too much, withdrawing too many uTokens that are being used as collateral, disabling collateral).
- Eventually, the borrower repays the borrowed position (in full or in part).

Additionally, the following events occur at `EndBlock`:

- Fees are added to the open borrow positions based on token-specific interest rate.

The `umee/x/leverage` module stores each open borrow position.
If the same user account opens multiple borrow positions in the same token, the second position simply increases the amount of the first.

Additionally, rather than segregating each borrow position with a specific collateral deposit (uToken coins) we aggregate them. The sum of all account's collateral uTokens related is used to calculate the account's borrow limit.
We define a **borrow limit** rule:
\__sum of account's borrow positions must be smaller than the account borrow limit_.

Note that the exchange rate of Asset:u-Asset has a dynamic exchange rate that grows with accruing interest - see [ADR-001: Interest Stream](./ADR-001-interest-stream.md).

In contrast, the exchange rate of collateral:borrowed assets (e.g. `atom:ether`) can only be determined using a price oracle.

The calculated borrow limit, which weighs collateral uTokens against borrowed assets (e.g. `u/atom:ether`) is derived from combining the two above. The weight of each uToken is defined as `CollateralWeight`.

Note also that as a consequence of uToken interest, the asset value of uToken collateral increases over time, meaning a user who repays positions in full and redeems collateral uTokens will receive back more base assets than they deposited originally, reducing the effective interest.

### Collateral utilization

Definitions:

- `total_supply(tokenA)`: total amount of tokenA provided to the leverage protocol (including coins marked as a collateral).
- `available_supply(tokenA) = total_supply(tokenA) - reserve - total_borrow(tokenA)`: amount of tokenA available in the system for borrowing.
- `supply_utilization(tokenA) = total_borrow(tokenA) / (total_supply(tokenA) - reserve)`. It equals 0 if there is no borrow for tokenA. It equals 1 if all ledable tokenA are borrowed.
- `total_collateral(tokenA)`: total amount of tokenA used as a collateral.

We define a _token collateral utilization_:

```go
collateral_utilization(tokenA) = total_collateral(tokenA) / available_supply(tokenA)
```

Note: system must not allow to have available_supply to equal zero.

Intuition: we want collateral utilization to grow when there is less liquid tokenA available in the system to cover the liquidation.
Collateral utilization of tokenA is growing when suppliers withdraw their tokenA collateral or when borrowers take a new loan of tokenA.
If a `tokenA` is not used a collateral then it's _collateral utilization_ is zero.
It is bigger than 1 when available supply is lower than the amount of `tokenA` used as a collateral.
When it is `N`, it means that only `1/N` of the collateral is available for redemption (u/tokenA -> tokenA).

#### Examples

Let's say we have 1000A (token A) supplied to the system (for lending or collateral). Below let's consider a state with total amount of A borrowed (B) and total amount of B used as a collateral (C) and computed collateral utilization (CU):

1. B=0, C=0 → CU=0
1. B=0, C=500 → CU=0.5
1. B=0, C=1000 → CU=1
1. B=500, C=0 → CU=0
1. B=500, C=500 → CU=1
1. B=500, C=1000 → CU=2
1. B=999, C=0 → CU=0
1. B=999, C=500 → CU=500
1. B=999, C=100 → CU=1000

#### Motivation

High collateral utilization is dangerous for the system:

- When collateral utilization is above 1, liquidators may not be able to withdraw their the liquidated collateral.
- Liquidators, when liquidating a borrower, they get into position their _uToken_.
  In case of bad market conditions and magnified liquidations, liquidators will like to redeem the _uToken_ for the principle (the underlying token).
  However, when there are many `uToken` redeem operation, the collateral utilization is approaching to 1 and liquidators won't be able to get the principle and sell it to monetize their profits.
  This will dramatically increase the risk of getting a profit by liquidators and could cause the system being insolvent.

Let's draw the following scenario to picture the liquidators risk:

1. Alice is providing \$1.2M USD supply.
2. Bob is providing \$1.5M in Luna as a collateral and borrows 1M USD from Alice.
3. Charlie provides \$2M in BTC as a collateral and borrows $1.4M in Luna from Bob.
4. Charlie predicts Luna collapse and sells the Luna.
5. Luna is sinking and Bob position has to be liquidated. However:
   - Suppliers can liquidate Bob, but they can only redeem up to 6.6% of `u/Luna` because the rest is not available (Charlie borrowed it).
   - Charlie will not pay off her borrow position - she will wait for the final collapse and buy Luna cheaply.
   - Liquidators will not take the risk of obtaining and holding `u/Luna` when there is a risk of Luna sinking deep.
6. In case of the big crash, liquidators won't perform a liquidation, Bob will run away with 1M USD, system will end up with a bad debt and obligation to pay Alice.

We propose to set a per token parameter: **max collateral utilization** stored in the `Token` registry (see ADR-004). The system will forbid to make any collateral related operation if the operation would move token _collateral utilization_ above _max collateral utilization_.

Stable assets will have high max collateral utilization (can go even to 50). Volatile assets should have significant smaller collateral utilization, and assets with high risk should have max collateral utilization close to 1.

## Detailed Design

For the purposes of borrowing and repaying assets, as well as marking uTokens as collateral, the `umee/x/leverage` module does not mint or burn tokens.
It stores borrower open positions and collateral settings, and uses the `x/bank` module to perform all necessary balance checks and token transfers.
User collateral (uTokens) are deposited in `x/leverage` module and withdrawn back to the user `x/bank` account balance when the user disables uTokens as a collateral.

During every operation which involves borrow position or collateral we check that the _borrow limit_ rule holds.

`x/oracle` module is used to provide exchange rates of tokens to calculate borrow limit.

### API

To implement the borrow/repay functionality of the Asset Facility, the three message types are defined:

```go
// MsgSetCollateral - a borrower enables or disables a specific uToken type in their wallet to be used as collateral
type MsgSetCollateral struct {
  Borrower sdk.AccAddress
  Denom    string
  Enable   bool
}

// MsgBorrowAsset - a user wishes to borrow assets of an allowed type
type MsgBorrowAsset struct {
  Borrower sdk.AccAddress
  // not a uToken
  Asset   sdk.Coin
}

// MsgRepayAsset - a user wishes to repay assets of a borrowed type
type MsgRepayAsset struct {
  Borrower sdk.AccAddress
  Asset    sdk.Coin
}
```

Tokens used in above messages must belong to the allow-list. Collateral must be a uToken.

Messages must be signed by the borrower's account.

Both CLI and gRPC must be supported for the above messages.

### Storage layout

Borrow positions are stored using a mechanism discussed in ADR-008

Using the `sdk.Coins` built-in type, which combines multiple {Denom,Amount} pairs as a single object, the `umee/x/leverage` module stores collateral settings and positions as follows:

```go

// borrower collateral settings for enabled denoms:
collateralSettingPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = 0x01

// and the amount of collateral deposited for each uToken:
collateralAmountPrefix | lengthPrefixed(borrowerAddress) | tokenDenom = sdk.Int

// max token collateral utilization setings
maxCollateralUtilizationPrefix | token = bigEndian(uint32)
```

This will be accomplished by adding new prefixes and helper functions to `x/leverage/types/keys.go`.

The use of borrowerAddress before tokenDenom in the store keys allows to "iterate by borrower" functionality, e.g. "all open borrow positions belonging to an individual user". The same applies to collateral settings and amounts.

In contrast, if we had put tokenDenom before borrower address, it would favor operations on the set of all keys associated with a given token.

## Alternative Approaches

- Allow amounts of uToken to be specifically marked as collateral, rather than toggling collateral on/off for each asset type. This would allow more fine-grained control of collateral by borrowers.

## Consequences

### Positive

- uTokens used as a collateral increase in base asset value in the same way that supply positions do. This counteracts borrow position interest.
- UX of enabling/disabling token types as collateral is simpler than depositing specific amounts
- `lengthPrefixed(borrowerAddress) | tokenDenom` key pattern facilitates getting open borrow and collateral positions by account address.

### Negative

- `lengthPrefixed(borrowerAddress) | tokenDenom` key pattern makes it more difficult to get all open borrow positions by token denomination.

### Neutral

- Borrow feature relies on an allow-list of token types
- Borrow feature relies on price oracles for base asset types
- Borrow interest will rely on dynamic rate calculation

## References
