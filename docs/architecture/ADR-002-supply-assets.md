# ADR 002: Supplying Assets

## Changelog

- September 13, 2021: Initial draft (@toteki)
- September 29, 2021: Update message names to reflect implementation (@toteki)
- July 2, 2022: Cleanup and simplify (@toteki)

## Status

Implemented

## Context

The `x/leverage` module allows users to supply and withdraw assets.

The flow of events is as follows:
- Lender supplies a token to the module
- Module mints and sends uTokens to the lender
- Lender redeems uTokens for original tokens plus interest after some time

## Decision

The following message types will be created

```proto
// MsgLendAsset represents a lender's request to supply a base asset type to the
// module. Coin denomination must be a registered Token.
message MsgLendAsset {
  string                   lender = 1;
  cosmos.base.v1beta1.Coin coin   = 2 [(gogoproto.nullable) = false];
}

When a user supplies tokens, the tokens are moved to the `x/leverage` module account. Simultaneously, uTokens are minted and sent to the user's wallet. The module uses its current token:uToken exchange rate.

// MsgWithdrawAsset represents a lender's request to withdraw supplied assets.
// Coin denomination must be a uToken.
message MsgWithdrawAsset {
  string                   lender = 1;
  cosmos.base.v1beta1.Coin coin   = 2 [(gogoproto.nullable) = false];
}
```

When a user withdraws tokens, the tokens are moved from the `x/leverage` module account back to their wallet. Simultaneously, uTokens are removed from the user's wallet and burned by the module. The module uses its current token:uToken exchange rate.

If the full requested amount of tokens is not available for withdrawal, the transaction fails.

## Consequences

### Positive
- [x/bank module]([Cosmos Bank Keeper](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/bank/spec/02_keepers.md)) provides the underlying functionality.

### Negative
- If tokens were to be sent unexpectedly to the `x/leverage` module account (see the warning on [this page](https://docs.cosmos.network/master/modules/bank/)), it would have the effect of donating said tokens to existing suppliers if the tokens were an accepted type. Otherwise, the tokens would remain inert.

### Neutral
- Requires an allow-list (Token Registry), to be implemented later

## References

- [Cosmos Bank Keeper](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/x/bank/spec/02_keepers.md)
