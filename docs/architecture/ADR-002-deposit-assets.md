# ADR 002: Depositing assets for uTokens

## Changelog

- September 13, 2021: Initial Draft (@toteki)
- September 29, 2021: (Housekeeping) update message names to reflect implementation (@toteki)

## Status

Accepted

## Context

One of the base functions of the Umee universal capital facility is to allow liquidity providers to deposit assets, and earn interest on their deposits.

The associated feature “Lender deposits asset for uToken & redeems uToken for a single Cosmos asset type" was initially discussed as follows:

- Lender deposits (locks) a Cosmos asset (like Atoms or Umee) into asset facilities.
- Facility mints and sends u-Assets in response (u-Atom, u-umee).
- Lender redeems u-Assets for the original assets.
- Asset facility knows its current balances of all asset types.
- Asset facility knows the amount of all u-Asset types in circulation.

The Cosmos `x/bank` module can be used as the basis for these capabilities.

## Alternative Approaches

While the proposed implementation will use the Cosmos banking module to simultaneously transfer assets and mint uTokens, it might also be possible to use a [Liquidity Pool](https://tutorials.cosmos.network/liquidity-module/) for the non-minting portion.
The capital facility would be just another account which offers trades on Asset:u-Asset pools (though it would still require a separate way of minting the uTokens it offers). This option should be considered if direct use of the banking module is rejected.

## Decision

The Cosmos `x/bank` module can be used as the basis for the required capabilities.

## Detailed Design

The Asset Facility will have the capability to mint and burn uTokens (but not their corresponding original asset types). It will have access to an allow-list of said asset and uToken types.

The Asset Facility will possess a `Module Account` to store original assets, and that module account should be forbidden from being the recipient of transactions except those intended by the module (see the warning on [this page](https://docs.cosmos.network/master/modules/bank/)).

The Asset Facility should harness the Cosmos `x/bank` module's `BaseKeeper` for the following capabilities

- Read the supply (chain-wide) of a given coin (both base assets and uTokens)
- Send coins (both base assets and uTokens) from module account to user account (and vice versa)
- Mint uTokens
- Burn uTokens
- `SendKeeper`: Use the _BlockedAddr_ feature to guard against unexpected transfers to module account(s)
- `ViewKeeper`: Read individual account balances

The [BaseKeeper](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/bank/spec/02_keepers.md) of the Cosmos `x/bank` module provides the required capabilities:

> The base keeper provides full-permission access: the ability to arbitrarily modify any account's balance and mint or burn coins.

Note that `BaseKeeper` also has functions to query the total amount of coins of each asset type registerd in the x/bank module, and can also query individual account balances using its embedded `ViewKeeper`.

### Basic Message Types

For reference, here is the `Bank` module's built-in coin transfer message as seen [here](https://docs.cosmos.network/v0.39/basics/app-anatomy.html), which is used when regular users send tokens to one another:

```go
// MsgSend - high level transaction of the coin module
type MsgSend struct {
  FromAddress sdk.AccAddress `json:"from_address" yaml:"from_address"`
  ToAddress   sdk.AccAddress `json:"to_address" yaml:"to_address"`
  Amount      sdk.Coins      `json:"amount" yaml:"amount"`
}
```

The `sdk.Coins` type is a slice (ordered list) of `sdk.Coin` which contains a denom (token type) and amount.

Asset Facility deposit functionality is provided by the two following message types:

```go
// MsgLendAsset - a lender wishes to deposit assets and receive uAssets
type MsgLendAsset struct {
  Lender sdk.AccAddress `json:"lender" yaml:"lender"`
  Amount sdk.Coin       `json:"amount" yaml:"amount"`
}

// MsgWithdrawAsset - redeems uAsset for original assets
type MsgWithdrawAsset struct {
  // Lender is the owner of the uAsset
  Lender sdk.AccAddress `json:"lender" yaml:"lender"`
  Amount sdk.Coin       `json:"amount" yaml:"amount"`
}
```

This resembles the built-in MsgSend, but either ToAddress or FromAddress is removed because the module's address should be used automatically. The remaining address is that of the lender.

MsgDepositAsset must use only allow-listed, non-uToken denominations. MsgWithdrawAsset must use only uToken denominations.

These messages should trigger the appropriate reaction (disbursement of uTokens after deposit, return of assets on withdrawal). The exchange rate defined in ADR-001 must be used.

_Note: The `Coin` type seen in the `Amount` fields contains a single token denomination and amount._

It is necessary that `MsgLendAsset` and `MsgWithdrawAsset` be signed by the lender's account. According to the [Transactions Page](https://docs.cosmos.network/master/core/transactions.html)

> Every message in a transaction must be signed by the addresses specified by its GetSigners.

Thus `MsgLendAsset.GetSigners` and `MsgWithdrawAsset.GetSigners` should return the `Lender` address.

### API

Both CLI and gRPC must be supported when sending the above message types, and all necessary handlers must be created in order to process and validate them as transactions. As part of this initial feature, an exact list of such steps required when adding message types will be created and added to future issues.

### Testing

Assuming a placeholder token allow-list of one element (e.g. `umee`), and a uToken existing (e.g. `u-umee`), an end-to-end test can be created in which one user account sends a `MsgLendAsset` and a `MsgWithdrawAsset` of the appropriate token types.

## Considerations

- IBC tokens coming from different channels will have different Denom. Hence we have a fragmentation problem. Example: `atom` coming directly from the Cosmos Hub will have different IBC denom than an `atom` coming directly from Osmosis.

## Consequences

### Positive

- Banking module already provides underlying functionality
- Asset deposit does not require us to modify state beyond what is already done by the `x/bank` module (e.g. account balances): deposit of an assets into a module account is atomic with the minting of respective uTokens to the sender's account, and uToken ownership is the sole requirement for asset withdrawal. We do not need to track individual user deposit history.

### Negative

### Neutral

- Asset facility will store base assets in a Module Account
- Asset facility relies on an allow-list of token types, to be implemented later

## References

- [Cosmos Bank Keeper](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/bank/spec/02_keepers.md)
