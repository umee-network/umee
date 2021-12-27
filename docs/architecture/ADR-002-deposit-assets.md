# ADR 002: Depositing assets for uTokens

## Changelog

- September 13, 2021: Initial Draft (@toteki)
- September 29, 2021: (Housekeeping) update message names to reflect implementation (@toteki)

## Status

Proposed

## Context

One of the base functions of the Umee universal capital facility is to allow liquidity providers to deposit assets, and earn interest on their deposits.

The associated feature “Lender deposits asset for uToken & redeems uToken for a single cosmos asset type" was initially discussed as follows:
- Lender deposits (locks) a cosmos asset type (likely Atoms or uumee) into asset facilities
- Facility mints and sends u-Assets in response (u-Atom, u-uumee)
- Lender redeems u-Assets for original assets
- Asset facility knows its current balances of all asset types
- Asset facility knows the amount of all uToken types in circulation

The Cosmos `x/bank` module can be used as the basis for these capabilities.

Note that the exchange rate of Assets:uAssets will be a shifting exchange rate that grows with interest - see [ADR-001: Interest Stream](./ADR-001-interest-stream.md).

## Alternative Approaches

While the proposed implementation will use the Cosmos banking module to simultaneously transfer assets and mint uTokens, it might also be possible to use a [Liquidity Pool](https://tutorials.cosmos.network/liquidity-module/) for the non-minting portion.
The capital facility would be just another account which offers trades on Asset:uAsset pools (though it would still require a separate way of minting the uTokens it offers). This option should be considered if direct use of the banking module is rejected.

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
- `SendKeeper`: Use the BlockedAddr feature to guard against unexpected transfers to module account(s)
- `ViewKeeper`: Read individual account balances

The [BaseKeeper](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/bank/spec/02_keepers.md) of the Cosmos `x/bank` module comes with the following capabilities:
> The base keeper provides full-permission access: the ability to arbitrarily modify any account's balance and mint or burn coins.

Note that `BaseKeeper` also has functions which read the total coins of each asset type in circulation on the chain, and can also read individual account balances using its embedded `ViewKeeper`.

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
The `sdk.Coins` type is a slice (ordered list) of `sdk.Coin` which contains a token type and amount.

To implement the deposit functionality of the Asset Facility, the two common message types will be:
```go
// MsgLendAsset - a lender wishes to deposit assets and receive uAssets
type MsgLendAsset struct {
  Lender sdk.AccAddress `json:"lender" yaml:"lender"`
  Amount sdk.Coin       `json:"amount" yaml:"amount"`
}
// MsgWithdrawAsset - a lender wishes to redeem uAssets for original assets
type MsgWithdrawAsset struct {
  Lender sdk.AccAddress `json:"lender" yaml:"lender"`
  Amount sdk.Coin       `json:"amount" yaml:"amount"`
}
```
This resembles the built-in MsgSend, but either ToAddress or FromAddress is removed because the module's address should be used automatically. The remaining address is that of the lender.

MsgDepositAsset must use only allow-listed, non-uToken denominations. MsgWithdrawAsset must use only uToken denominations.

These messages should trigger the appropriate reaction (disbursement of uTokens after deposit, return of assets on withdrawal). The exchange rate defined in ADR-001 must be used.

_Note: The `Coin` type seen in the `Amount` fields contains a single token denomination and amount._

It is necessary that `MsgLendAsset` and `MsgWithdrawAsset` be signed by the lender's account. According to the [Transactions Page](https://docs.cosmos.network/master/core/transactions.html)
>Every message in a transaction must be signed by the addresses specified by its GetSigners.

Thus `MsgLendAsset.GetSigners` and `MsgWithdrawAsset.GetSigners` should return the `Lender` address.

### APIs and Handlers
Both CLI and gRPC must be supported when sending the above message types, and all necessary handlers must be created in order to process and validate them as transactions. As part of this initial feature, an exact list of such steps required when adding message types will be created and added to future issues.

### Testing

Assuming a placeholder token allow-list of one element (e.g. `uumee`), and a uToken existing (e.g. `u-uumee`), an end-to-end test can be created in which one user account sends a `MsgLendAsset` and a `MsgWithdrawAsset` of the appropriate token types.

## Open Questions
- How will we allow-list asset types for deposit into the asset facilities?
- How can IBC tokens be identified in such a way that they are unique, and immune to counterfeit? (e.g. someone makes a new Cosmos blockchain and Token with identical ChainID and token name to an existing one)

## Consequences

### Positive
- Banking module already provides underlying functionality
Asset deposit does not require us to modify state beyond what is already done by the `x/bank` module (e.g. account balances): Deposit of assets into the module account is simultaneous with the minting of uTokens to the sender's account, and uToken ownership is the sole requirement for asset withdrawal. We do not need to track individual user deposit history.

### Negative

### Neutral
- Asset facility will store base assets in a Module Account
- Asset facility relies on an allow-list of token types, to be implemented later

## References

- [Cosmos Bank Keeper](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/bank/spec/02_keepers.md)
