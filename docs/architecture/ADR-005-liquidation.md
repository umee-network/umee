# ADR 005: Liquidation

## Changelog

- November 19, 2021: Initial Draft (@toteki)

## Status

Proposed

## Context

When borrowers on Umee exceed their borrow limit due to interest accrual or asset price fluctuations, their positions become eligible for liquidation by other users.

Such users (Liquidators) pay off part of all of the borrower's loan, in exchange for a value of collateral equal to the amount paid off plus an additional percentage (the liquidation incentive).

We must build the features necessary for liquidators to continuously look out for liquidation opportunities, then carry out chosen liquidations.

Additional parameters will be required which define the liquidation incentive and any other new variables.

## Decision

Liquidation will require one message type (`MsdLiquidate`), one per-token parameter (`LiquidationIncentive`), and potentially other parameters if a dynamic liquidation incentive is used.

There is no event type for when a borrower becomes a valid liquidation target, nor a list of valid targets stored in the module. Liquidators will have to use an off-chain tool to query their nodes periodically.

## Detailed Design

A function `IsLiquidationEligible(borrowerAddr)` can be created to determine if a borrower is currently exceeding their borrow limit. Any liquidation attempt against a borrower not over their limit should fail.

A borrower's total borrower value (expressed in USD) can be computed from their total borrowed tokens and a price oracle.

Their borrow limit is calculated similarly using borrower uToken balances, borrower collateral settings, current uToken exchange rates, and token collateral weights.

### Basic Message Type

To implement the liquidation functionality of the Asset Facility, one message type is required:

```go
// MsgLiquidate - a liquidator targets a specific borrower, asset type, and collateral type for complete or partial liquidation
type MsgLiquidate struct {
  Liquidator    sdk.AccAddress
  Borrower      sdk.AccAddress
  RepayDenom    string
  RepayAmount   sdk.Int
  RewardDenom   string
}
```

RepayDenom is the borrowed asset denom to be repaid (because the borrower may have multiple open borrows). It is always a base asset type (not a uToken).

RepayAmount is the maximum amount of asset the liquidator is willing to repay. This field enables partial liquidation.

RewardDenom is the collateral type which the liquidator will recieve in exchange for repaying the borrower's loan. It is always a uToken denomination.

It is necessary that messages be signed by the liquidator's account. Thus the method `GetSigners` should return the `Liquidator` address for the message type above.

### Partial Liquidation

It is possible for the amount repaid by a liquidator to be less than the borrower's total borrow position in the target denom. Such partial liquidations can succeed without issue, reducing the borrowed amount by `RepayAmount` and rewarding collateral proportional to the amount repaid.

Additionally, the `RepayAmount` found in `MsgLiquidate` functions as a maximum the liquidator is willing to repay. There are multiple factors that may reduce the actual repayment below this amount:

- Borrowed amount of `RepayDenom` is below `RepayAmount`
- Insufficient `RewardDenom` collateral to match repaid value plus liquidation incentive
- Parameter-based maximum, e.g. if a parameter `LiquidationCap = 0.50` caps liquidated value in a single transaction at 50% of the total borrowed value.

In the above scenarios, the `MsgLiquidate` should succeed with the maximum amount allowed by the conditions, rather than fail outright.

### Token Parameters

In order to incentivize liquidators to target certain collateral types for liquidation first, the token parameter `LiquidationIncentive` is used.

When a `MsgLiquidate` causes liquidation to occur, the liquidator receives collateral equal to (100% + `RewardDenom.LiquidationIncentive`) of the repaid value worth of collateral.

For example, if the liquidation incentive for `uatom` is `0.15`, then the liquidator receives `u/uatom` collateral worth 115% of the borrowed base assets they repaid. The denom of the base assets does not affect this calculation.

### Calculating Liquidation Amounts

When a `MsgLiquidate` is received, the `x/leverage` module must determine if the targeted borrow address is eligible for liquidation.

```go
    // from MsgLiquidate (liquidatorAddr, borrowerAddr, repayDenom, repayAmount, rewardDenom)

    borrowed := GetTotalBorrows(borrowerAddr)
    borrowValue := oracle.TotalValue(borrowed) // price oracle

    collateral := GetCollateralBalance(borrowerAddr)
    collateral = MultiplyByCollateralWeight(collateral)
    collateralValue := oracle.TotalValue(collateral) // price oracle

    if borrowValue > collateralValue {
      // borrower is over their borrow limit, and therefore eligible for liquidation
    }
```

After eligibility is confirmed, parameters governing liquidation can be fetched:

```go
    // This function allows for the possibility of dynamic liquidation parameters
    // based on collateral denomination, borrowed value, and collateral value
    liquidationIncentive, closeFactor := GetLiquidationParameters(rewardDenom, borrowValue, collateralValue)
```

The liquidation incentive is the bonus collateral received when a liquidator repays a borrow position.
(e.g. incentive=`0.2` means liquidator receives 120% the value of their repayment back in collateral)

The close factor is the portion of a borrow position eligible for liquidation in this single liquidation event.

See _Alternatives_ section at the bottom of this document for multiple possible implementations of dynamic liquidation parameters. Any such feature can be isolated inside the function above.

Once parameters are fetched, the final liquidation amounts (repayment and reward) must be calculated.

```go
   // Repayment cannot exceed liquidator's available balance
   liquidatorBalance := GetBalance(liquidatorAddr).Amount(repayDenom)
   repayAmount = Min(repayAmount, liquidatorBalance)

   // Repayment cannot exceed borrowed value * close factor
   repayValue := oracle.Value(sdk.NewCoin(repayDenom, repayAmount)) // price oracle
   if repayValue > borrowValue * closeFactor {
     partial := (borrowValue * closeFactor) / repayValue
     repayAmount = repayAmount * partial
   }

   // Collateral reward assuming current repayment amount
   rewardAmount := oracle.EquivalentValue(repayDenom, repayAmount, rewardDenom.ToBaseAssetDenom) // price oracle
   rewardAmount = rewardAmount / (GetExchangeRate(rewardDenom)) // apply uToken exchange rate
   rewardAmount = rewardAmount * (1 + liquidationIncentive) // apply liquidation incentive, e.g. +10%

   // Reward amount cannot exceed available collateral
   if rewardAmount > GetBalance(borrowerAddr,rewardDenom) {
     // only repay what can be correctly compensated
     partial := GetBalance(borrowerAddr,rewardDenom) / rewardAmount
     repayAmount = repayAmount * partial 
     // use all collateral of rewardDenom
     rewardAmount = GetBalance(borrowerAddr,rewardDenom) 
   }
```

Then the borrow can be repaid and the collateral rewarded using the liquidator's account.

## Consequences

### Positive
- Dynamic liquidation incentives can be supported

### Negative
- Offchain tool required to effectively scan for liquidation opportunities

### Neutral
- New message type `MsgLiquidate` is created
- New per-token parameters will be created to determine liquidation incentives
- New global parameters may be created if we choose to use dynamic close factors

## Alternatives: Dynamic Liquidation Parameters

[Study](https://arxiv.org/pdf/2106.06389.pdf) of existing DeFi protocols with fixed incentive liquidation has concluded the following:

> the existing liquidation designs well incentivize liquidators but sell excessive amounts of discounted collateral at the borrowers’ expenses.

Examining one exising liquidation scheme ([Compound](https://zengo.com/understanding-compounds-liquidation/)), two main parameters define maximum borrower losses due to liquidation:
- Liquidation Incentive (10%)
- Close Factor (50%)
When a borrower is even 0.0001% over their borrow limit, they stand to lose value equal to 5% their borrowed value in a single liquidation event.
That is, the liquidator pays off 50% of their borrow and receives collateral worth 55% of its value.

It should be possible to improve upon this aspect of the system by scaling the two parameters shown above, based on how far a borrower is over their borrow limit. There are two basic possibilities:

> Option 1: Dynamic liquidation Incentive
>
> Liquidation Incentive ranges from `0% = MinLiquidationIncentive` to `10% = MaxLiquidationIncentive` when the borrower is between 0% and `10% = FullLiquidationIncentiveThreshold` over their borrow limit, then stays at `MaxLiquidationIncentive`.
> 
> | Borrow Limit (BL) |Borrowed Value (BV) | BV / BL | Liquidation Incentive |
> | - | - | - | - |
> | 100 | 100.1 | 1.001 | 0.1% |
> | 100 | 102 | 1.02 | 2% |
> | 100 | 105 | 1.05 | 5% |
> | 100 | 110 | 1.1 | 10% |
> | 100 | 140 | 1.4 | 10% |

The Dynamic liquidation Incentive option would take advantage of market forces to reduce the liquidation incentive to the lowest sustainable value.
Liquidators would have a chance to liquidate for a lower incentive (e.g. 2%) if deemed profitable, otherwise the incentive continues to increase as the borrow accrues interest.

The `CloseFactor` parameter can be any value without affecting this method.

In this method, token-specific liquidation incentives (which incentivize liquidation of some asset types before others) would likely modify the parameter `MaxLiquidationIncentive`.

> Option 2: Dynamic close factor
>
> Close factor ranges from `0.0 = MinimumCloseFactor` to 1.0 when the borrower is between 0% and `20% = CompleteLiquidationThreshold` over borrow limit, then stays at 1.0
> 
> | Borrow Limit (BL) |Borrowed Value (BV) | BV / BL | Close Factor |
> | - | - | - | - |
> | 100 | 100.1 | 1.001 | 0.005 |
> | 100 | 102 | 1.02 | 0.1 |
> | 100 | 110 | 1.1 | 0.5 |
> | 100 | 130 | 1.2 | 1.0 |
> | 100 | 140 | 1.4 | 1.0 |

The Dynamic Close Factor option would again take advantage of market forces to reduce excessive collateral selloffs, by reducing the portion of collateral initially eligible for liquidation.
Liquidators would have the change to liquidate smaller portions of the borrow if profitable and bring the position back into health.
Otherwise, close factor would continue to increase as the borrow accrues interest.

This also allows borrows to be liquidated completely in one transaction, once they arr serverely over their borrow limit.

The `LiquidationIncentive` parameter can be any value, including varying from token to token, without affecting this method.

> **++++++++++++++++++**
> **DECISIONS REQUIRED**
> **++++++++++++++++++**
> 
> Which strategy should we employ for dynamic liquidation parameters
> 
> a) No dynamic liquidation. Has parameters:
>    - Per-token `LiquidationIncentive`
>    - Global parameter `CloseFactor`
>
> b) Dynamic incentive (option 1), with parameters:
>    - Per-token `MinLiquidationIncentive`
>    - Per-token `MaxLiquidationIncentive`
>    - Global parameter `FullLiquidationIncentiveThreshold`
>    - Global parameter `CloseFactor`
>
> c) Dynamic close factor (option 2), with parameters:
>    - Per-token `LiquidationIncentive`
>    - Global parameter `MinimumCloseFactor`
>    - Global parameter `CompleteLiquidationThreshold`
>
> Once this has a final decision, I can update the document appropriately.

## References

- [An Empirical Study of DeFi Liquidations:Incentives, Risks, and Instabilities](https://arxiv.org/pdf/2106.06389.pdf)
- [Understanding Compound's Liquidation](https://zengo.com/understanding-compounds-liquidation/)