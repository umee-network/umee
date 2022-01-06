# Overview

This document covers basic concepts and math that determine the `leverage` module's behavior.

## Accepted Assets

At the foundation of the `leverage` module is the [Token Registry](02_state.md#Token-Registry), which contains a list of accepts base asset types.

This list is controlled by governance, and serves to limit the asset types available for transactions like borrowing and lending, and also any query services based on denomination.

### Token Parameters

Each `Token` has parameters that govern its behavior.

```protobuf
// Token defines a token, along with its capital metadata, in the Umee capital
// facility that can be loaned and borrowed.
message Token {
  option (gogoproto.equal) = true;

  // The base_denom defines the denomination of the underlying base token.
  string base_denom = 1 [(gogoproto.moretags) = "yaml:\"base_denom\""];

  // The reserve factor defines what portion of accrued interest of the asset type
  // goes to reserves.
  string reserve_factor = 2 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"reserve_factor\""
  ];

  // The collateral_weight defines what amount of the total value of the asset
  // can contribute to a users borrowing power. If the collateral_weight is zero,
  // using this asset as collateral against borrowing will be disabled.
  string collateral_weight = 3 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"collateral_weight\""
  ];

  // The base_borrow_rate defines the base interest rate for borrowing this
  // asset.
  string base_borrow_rate = 4 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"base_borrow_rate\""
  ];

  // The kink_borrow_rate defines the interest rate for borrowing this
  // asset when utilization is at the 'kink' utilization value as defined
  // on the utilization:interest graph.
  string kink_borrow_rate = 5 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"kink_borrow_rate\""
  ];

  // The max_borrow_rate defines the interest rate for borrowing this
  // asset (seen when utilization is 100%).
  string max_borrow_rate = 6 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"max_borrow_rate\""
  ];

  // The kink_utilization_rate defines the borrow utilization rate for this
  // asset where the 'kink' on the utilization:interest graph occurs.
  string kink_utilization_rate = 7 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"kink_utilization_rate\""
  ];

  // The liquidation_incentive determines the portion of bonus collateral of
  // a token type liquidators receive as a liquidation reward.
  string liquidation_incentive = 8 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec",
    (gogoproto.nullable)   = false,
    (gogoproto.moretags)   = "yaml:\"liquidation_incentive\""
  ];

  // The symbol_denom and exponent are solely used to update the oracle's accept
  // list of allowed tokens.
  string symbol_denom = 9 [(gogoproto.moretags) = "yaml:\"symbol_denom\""];
  uint32 exponent     = 10 [(gogoproto.moretags) = "yaml:\"exponent\""];
}
```

### uTokens

Every base asset has an associated _uToken_, which is received when lending the asset and can be exchanged back for an equal or greater amount of the same base asset as governed by the uToken's [uToken Exchange Rate](01_overview.md#uToken-Exchange-Rate).

uTokens do not have parameters like the `Token` struct does, and they are always represented in account balances with a denom of `UTokenPrefix+token.BaseDenom`. For example, the base asset `uumee` is associated with the uToken denomination `u/uumee`.

## Lending and Borrowing

TODO