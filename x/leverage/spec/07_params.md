# Parameters

The leverage module contains the following parameters:

| Key                          | Type    | Example |
| -----------------------------| ------- | ------- |
| CompleteLiquidationThreshold | sdk.Dec | 0.1     |
| MinimumCloseFactor           | sdk.Dec | 0.01    |
| OracleRewardFactor           | sdk.Dec | 0.01    |
| SmallLiquidationSize         | sdk.Dec | 100.00  |

## CompleteLiquidationThreshold

CompleteLiquidationThreshold governs how far above their liquidation limit a borrower
must be to have a [Close Factor](01_concepts.md#Close-Factor) of 1.0 - that is,
to be eligible for full liquidation in a single liquidation event.

## MinimumCloseFactor

MinimumCloseFactor is the [Close Factor](01_concepts.md#Close-Factor) for
borrows that are just above their liquidation limit.

## OracleRewardFactor

OracleRewardFactor is the portion of borrow interest accrued that goes to fund
the `x/oracle` reward pool.

## SmallLiquidationSize

SmallLiquidationSize is the borrow value in USD below which [Close Factor](01_concepts.md#Close-Factor)
is always 1.