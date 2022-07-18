# ADR 010: Safety Parameters

## Changelog

- Jul 4, 2022: Max collateral utilization added to ADR-003 (@robert-zaremba)
- Jul 18, 2022: Initial draft moved discussion to new ADR (@toteki)

## Status

Proposed

## Context

Umee manages the health of individual borrowers using borrow limits, collateral weights, and liquidation thresholds, but it currently lacks parameters that manages the health of token markets as a whole.

Several new parameters and alternatives are being proposed.

## Decision

Parameters we decide to use will be listed below in subsections:
- Max Collateral Share
- Max Borrow Utilization
- Min Collateral Liquidity

Rejected or alternative implementations will appear in the alternatives section:
- Max Collateral Utilization

### Useful Definitions

The following terms may appear in multiple discussions below:

- `total_supply(token)` total amount of a base token which has been supplied to the system, including that which has been borrowed out, plus outstanding interest, minus reserves.
- `total_supply(utoken)` total amount of uTokens of a given type in existence. When exchanged for base tokens, this amount is worth exactly the total supply.
- `total_collateral(utoken)` total amount of a uToken marked as collateral.
- `total_collateral(token)` total amount of a uToken marked as collateral, multiplied by its uToken exchange rate to get the amount of base token collateral it represents.
- `total_borrowed(token)` the sum of all existing debt in a base token, including interest.
- `reserved(token)` the amount of tokens in the module balance which are reserved for paying off bad debt. These tokens are excluded from total supply.
- `available(token)` the amount of supplied tokens which have not been borrowed out or reserved.

These equations follow:

```go
available(token) = module_balance(token) - reserved(token)
total_supply(token) = total_borrowed(token) + available(token)
uToken_exchange_rate(token) = total_supply(token) / total_supply(utoken)
```

### Maximum Borrow Utilization

One proposed parameter is `MaxBorrowUtilization`, to be defined per token.

```go
borrow_utilization(token) = total_borrowed(token) / total_supply(token) // ranges 0 - 1
```

Implementing `MaxBorrowUtilization` would restrict `MsgBorrow` from increasing `total_borrowed` above a desired level.

It may or may not restrict `MsgWithdraw` from decreasing `total_supply(token)` below a desired level - adding this restriction might trap suppliers in a liquidity crisis in exchange for keeping more supply available for `MsgLiquidate`.

| Message Type | Current Decision |
| - | - |
| `MsgBorrow` | Restrict |
| `MsgWithdraw` | Allow |
| `MsgLiquidate` | Allow |

`MaxBorrowUtilization` could still be indirectly exceeded by borrow interest accruing.

Additionally, dynamic interest rates, which functioned as borrow utilization ranged from `0` to `1`, would instead be adjusted expect values from `0` to `MaxBorrowUtilization`.

The motivation for restricting borrow utilization is to reduce the likelihood of situations where suppliers cannot use `MsgWithdraw` due to borrowers borrowing all available supply.

This parameter overlaps in function with `MinCollateralLiquidity`. By lowering `MaxBorrowUtilization`, we create a buffer zone where `MsgBorrow` cannot reduce liquidity any further, but `MsgWithdraw` is still available. This helps protect suppliers and makes it more difficult to reach `MinCollateralLiquidity` overall.

### Maximum Collateral Share

Another proposed parameter is `MaxCollateralShare`, to be defined per token.

```go
total_collateral_value(token) = total_collateral(token) * oracle_price_usd(token)
collateral_share(token) = total_collateral_value(token) / Sum_All_Tokens(total_collateral_value) // ranges 0 - 1
```

Implementing `MaxCollateralShare` would restrict `MsgAddCollateral` from increasing `total_collateral` above a desired level.

`MaxCollateralShare` could still be indirectly exceeded by fluctuating oracle prices, `MsgDecollateralize` or `MsgLiquidate` of other tokens, or interest accruing in one denom faster than another.

### Minimum Collateral Liquidity

Another proposed parameter is `MinCollateralLiquidity`, to be defined per token.

It is the reciprocal of `MaxCollateralUtilization`, functionally equivalent in which messages it would restrict.

```go
collateral_liquidity(token) = available(token) / total_collateral(token) // ranges 0 - ∞
```

Stable assets can have low `MinCollateralLiquidity` (as low as `0.02`, but more likely around `0.15`). Volatile assets should have significantly safer values, for example `0.5` or `1`.

Implementing `MinCollateralLiquidity` would restrict `MsgBorrow` and `MsgWithdraw` from decreasing `available_supply`, or `MsgCollateralize` from increasing `total_collateral` in certain market conditions.

It may or may not allow `MsgLiquidate` to decrease `available_supply` under the same conditions, to prevent crises.

| Message Type | Current Decision |
| - | - |
| `MsgBorrow` | Restrict |
| `MsgWithdraw` | Restrict |
| `MsgLiquidate` | Allow |

`MinCollateralLiquidity` could still be indirectly exceeded by supply interest accruing on a collateral denom's uToken exchange rate.

It has also been proposed separately to factor `collateral_liquidity` (or `collateral_utilization`) into dynamic interest rates, to enhance the current model which uses `borrow_utilization` (see ADR-004).

## Alternatives

### Collateral Utilization

One proposed parameter is `MaxCollateralUtilization`, to be defined per token.

It is the reciprocal of `MinCollateralLiquidity`, so the motivation and tradeoffs will be the same.

```go
collateral_utilization(token) = total_collateral(token) / available(token) // ranges 0 - ∞
```

This quantity has the property of increasing rapidly (as 1/N -> 0) when available supply is under stress.

Stable assets will have high max collateral utilization (can go even to 50). Volatile assets should have significantly smaller collateral utilization, and assets with high risk should have max collateral utilization close to 1.

#### Motivation

High collateral utilization is dangerous for the system: When collateral utilization is above 1, liquidators may not be able to withdraw their the liquidated collateral.

Let's draw the following scenario to picture the liquidators risk:

> | - | Supply | Collateral | Borrowed |
> | Alice| $1.2M BTC | - | - |
> | Bob | - | $1.5M LUNA | $1M BTC |
> | Charlie | - | $2M BTC | $1.4M LUNA |
>
> 1. Charlie predicts Luna collapse and sells the Luna.
> 2. Luna is sinking and Bob's position has to be liquidated. However:
>   - Liquidators can liquidate Bob, but they can only redeem up to 6.6% of `u/Luna` because the rest is not available (Charlie borrowed it).
>   - Charlie will not pay off her borrow position - she will wait for the final collapse and buy Luna cheaply.
>   - Liquidators will not take the risk of obtaining and holding `u/Luna` when there is a risk of Luna sinking deep.
> 3. In case of the big crash, knowledgeable liquidators won't liquidate Bob, Bob will run away with $1M of BTC, and the system will end up with a bad debt and obligation to pay Alice.

## Consequences

### Positive
- Multiple ways of preserving market health
- Per-token parameters allow fine grained control

### Negative
- Users may find restrictions unfair, e.g. "why can't I borrow just because other people's collateral is too high?"