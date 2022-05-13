# ADR 005: Liquidation

## Changelog

- November 19, 2021: Initial Draft (@toteki)
- May 2022: documentation updates (@robert-zaremba)

## Status

Accepted

## Context

When borrowers on Umee exceed their borrow limit due to interest accrual or asset price fluctuations, their positions become eligible for liquidation.

Third party liquidators pay off part of all of the borrower's loan, in exchange for a value of collateral equal to the amount paid off plus an additional bonus (the liquidation incentive).

We must build the features necessary for liquidators to continuously look out for liquidation opportunities, then carry out chosen liquidations.

Additional parameters will be required which define the liquidation incentive and any other new variables.

## Decision

Liquidation will require one message type `MsgLiquidate`, one per-token parameter `LiquidationIncentive`, and two global parameters `MinimumCloseFactor` and `CompleteLiquidationThreshold`.

The blockchain doesn't issue any event to signal that a borrow position can be liquidated, nor provide a list of valid targets. Liquidators will have to use an off-chain tools to query their nodes periodically.

## Detailed Design

We don't provide a function that checks if a given borrower can be liquidated to avoid spamming an app with periodical queries. Any liquidation attempt against a borrower not eligible for liquidation will fail.

A borrow position is represented by a pair `(borrower_address, coin)`, where borrower address is an entity requesting a loan. A borrower's total borrowed value (expressed in USD) can be computed from their total borrowed tokens and the `x/oracle` price oracle module.

Their borrow limit is calculated similarly using the borrower's uToken balance, their collateral settings, current uToken exchange rates, and token collateral weights. Liquidation happens when a sum of borrower loans is bigger than the `CalculateMaxBorrow(borrower_collateral)`.

During liquidation any of the borrower's collateral token can be liquidated to pay off any borrower's loan denom.

### Message Types

To implement the liquidation functionality of the Asset Facility, one message type is required:

```go
// MsgLiquidate - a liquidator targets a specific borrower, asset type, and
// collateral type for complete or partial liquidation
type MsgLiquidate struct {
  Liquidator    sdk.AccAddress
  Borrower      sdk.AccAddress
  Repayment     sdk.Coin // borrow denom + amount
  RewardDenom   string   // collateral denom
}
```

Repayment's denom is the borrowed asset denom to be repaid (because the borrower may have multiple open borrows). It is always a base asset type (not a uToken).
Its amount is the maximum amount of asset the liquidator is willing to repay. This field enables partial liquidation.

RewardDenom is the collateral type which the liquidator will receive in exchange for repaying the borrower's loan. It is always a uToken denomination.

`Liquidator` is the signer of the message and the account which will do repayment and receive reward.

### Partial Liquidation

It is possible for the amount repaid by a liquidator to be less than the borrower's total borrow position in the target denom. Such partial liquidations can succeed without issue, reducing the borrowed amount by `Repayment.Amount` and rewarding collateral proportional to the amount repaid.

Additionally, the `Repayment.Amount` found in `MsgLiquidate` functions as a maximum the liquidator is willing to repay. There are multiple factors that may reduce the actual repayment below this amount:

- Borrowed amount of `Repayment.Denom` is below `Repayment.Amount`
- Insufficient `RewardDenom` collateral to match repaid value plus liquidation incentive
- Parameter-based maximum, e.g. if a parameter `LiquidationCap = 0.50` caps liquidated value in a single transaction at 50% of the total borrowed value.

In the above scenarios, the `MsgLiquidate` should succeed with the maximum amount allowed by the conditions, rather than fail outright.

### Token Parameters

In order to incentivize liquidators to target certain collateral types for liquidation first, we introduce a `LiquidationIncentive` parameter - defined for each supported borrowed denom.

When a `MsgLiquidate` causes liquidation to occur, the liquidator receives collateral equal to (100% + `RewardDenom.LiquidationIncentive`) of the repaid value worth of collateral.

For example, if the liquidation incentive for `atom` is `0.15`, then the liquidator receives `u/atom` collateral worth 115% of the borrowed base assets they repaid. The denom of the base assets does not affect this calculation.

### Calculating Liquidation Amounts

When a `MsgLiquidate` is received, the `x/leverage` module must determine if the targeted borrow position is eligible for liquidation.

```go
    // from MsgLiquidate (liquidatorAddr, borrowerAddr, repayDenom, repayAmount, rewardDenom)

    borrowed := GetTotalBorrows(borrowerAddr)
    borrowValue := TotalValue(borrowed) // price oracle

    collateral := GetCollateralBalance(borrowerAddr)
    maxBorrowValue := CalculateMaxBorrow(collateral)

    if borrowValue > maxBorrowValue {
      // borrower is over borrow limit, and therefore eligible for liquidation
    }
```

After eligibility is confirmed, parameters governing liquidation can be fetched:

```go
    // This function allows for dynamic liquidation parameters based on
    // collateral denomination, borrowed value, and collateral value
    liquidationIncentive, closeFactor := GetLiquidationParameters(rewardDenom, borrowValue, collateralValue)
```

The liquidation incentive is a collateral bonus received when a liquidator repays a borrowed position
(e.g. incentive=`0.2` means liquidator receives 20% extra of the liquidated collateral).

The close factor is the maximum portion of a borrow position eligible for liquidation in a single liquidation event.

See _Dynamic Liquidation Parameters_ section at the bottom of this document.

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

   // Given repay denom and amount, calculate the amount of rewardDenom
   // that would have equal value
   rewardAmount := EquivalentValue(repayDenom, repayAmount, rewardDenom.ToBaseAssetDenom) // price oracle
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

## Dynamic Liquidation Parameters

[Study](https://arxiv.org/pdf/2106.06389.pdf) of existing DeFi protocols with fixed incentive liquidation has concluded the following:

> the existing liquidation designs well incentivize liquidators but sell excessive amounts of discounted collateral at the borrowers’ expenses.

Examining one existing liquidation scheme ([Compound](https://zengo.com/understanding-compounds-liquidation/)), two main parameters define maximum borrower losses due to liquidation:

- Liquidation Incentive (10%)
- Close Factor (50%)
  When a borrower is even 0.0001% over their borrow limit, they stand to lose value equal to 5% of their borrowed value in a single liquidation event.
  That is, the liquidator liquidates 50% of the borrowed value and receives 5% extra in collateral.

It should be possible to improve upon this aspect of the system by scaling one of the two parameters shown above, based on how far a borrower is over their borrow limit.

> Dynamic close factor
>
> Close factor ranges from `0.0 = MinimumCloseFactor` to 1.0 when the borrower is between 0% and `20% = CompleteLiquidationThreshold` over borrow limit, then stays at 1.0
>
> | Borrow Limit (BL) | Borrowed Value (BV) | BV / BL | Close Factor |
> | ----------------- | ------------------- | ------- | ------------ |
> | 100               | 100.1               | 1.001   | 0.005        |
> | 100               | 102                 | 1.02    | 0.1          |
> | 100               | 110                 | 1.1     | 0.5          |
> | 100               | 130                 | 1.2     | 1.0          |
> | 100               | 140                 | 1.4     | 1.0          |

The Dynamic Close Factor takes advantage of market forces to reduce excessive collateral selloffs, by reducing the portion of collateral initially eligible for liquidation.
Liquidators would have the chance to liquidate smaller portions of the borrow if profitable and bring the position back into health.
Otherwise, the close factor would continue to increase as the borrow accrues interest.

This also allows borrows to be liquidated completely in one transaction, once they are severely over their borrow limit.

The `LiquidationIncentive` parameter can be any value, varying from token to token, without affecting the close factor.

## Querying for Eligible Liquidation Targets

The offchain liquidation tool employed by liquidators will need several queries to be supported in order to periodically scan for and act on liquidation opportunities:

- GetLiquidationTargets: Return a list of all borrower addresses that are over their borrow limits
- GetTotalBorrows(borrower): Returns an `sdk.Coins` containing all assets borrowed by a single borrower
- GetTotalCollateral(borrower): Returns an `sdk.Coins` containing all uTokens in borrower's balance that are enabled as collateral

In addition to the above, the liquidation tool should be able to read any global or per-token parameters in order to finish calculating borrow limits and potential liquidation rewards.

## Consequences

### Positive

- Dynamic close factors reduce excessive risk to collateral

### Negative

- Offchain tool required to effectively scan for liquidation opportunities

### Neutral

- New message type `MsgLiquidate` is created
- New per-token parameter `LiquidationIncentive` will be created to determine liquidation incentives
- New global parameters `MinimumCloseFactor` and `CompleteLiquidationThreshold` will be created for close factors

## References

- [An Empirical Study of DeFi Liquidations:Incentives, Risks, and Instabilities](https://arxiv.org/pdf/2106.06389.pdf)
- [Understanding Compound's Liquidation](https://zengo.com/understanding-compounds-liquidation/)
