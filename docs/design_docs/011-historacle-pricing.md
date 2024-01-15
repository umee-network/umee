# Design Doc 011: Historacle Pricing

## Changelog

- Nov 09, 2022: Initial feature description (@adamewozniak)
- Nov 14, 2022: Additional scenario description (@brentxu)
- Nov 15, 2022: Add AcceptList functionality (@adamewozniak)
- Nov 28, 2022: Remove AcceptList functionality & Add Median Stamps (@adamewozniak)
- Dec 1, 2022: Update computation (@robert-zaremba)
- Jan 2, 2023: Add Avg algorithm (@robert-zaremba)

## Status

Accepted

## Abstract

In order to support small-volume assets in a safe manner, we need to be able to calculate the weighted medians of our prices over the past X amount of time (eg: 6 days).
This will allow the leverage module to make decisions around when to allow additional borrowing activity if an asset price is considered abnormal. Moreover we also support average prices to support other measures, like IBC outflow quota.

## Context

Right now, the volume of assets such as Juno, Umee, and Osmo are so low that it would be relatively easy to use the leverage protocol on UMEE to manipulate the prices of these assets and game the system.

The attack goes:

1. Attacker deposits & collateralizes a large amount FOO, a token with low volume on exchanges.
2. Attacker spikes the price of FOO on the exchanges by buying a large amount.
3. Attacker borrows USDC using their FOO as collateral at its current (spiked) oracle price.
4. The price of FOO returns to normal, and the value of the USDC exceeds the value of the FOO collateral.
5. Attacker exits the market for a profit.

In order to avoid these attacks, and to continue with our goal of allowing users to collateralize and borrow low-volume assets, we need to have a safety net.

Currently, this is defended against by disallowing the use of `Umee` as collateral, however, we would like to re-enable that and list other assets with low volume for collateral.

Historacle pricing will provide an API for the leverage module to tell when prices have changed in an abnormal fashion and avoid such events.

## Specification

These values are stored in state in order to avoid the `x/leverage` module from having to calculate them while making decisions around allowable positions for users to take.

### Parameters

We define epoch periods, during which additional computation will be performed:

- `Historic Stamp Period`: will determine how often exchange rates are stamped & stored, until the `Maximum Price Stamps` is met.
- `Median Stamp Period`: will determine how often the Median and the Standard Deviations around the Median are calculated, which will also be stored in the state machine.

Hardcoded parameters:

- `AvgPeriod`: will determine the length of the window where we average prices.
- `AvgShift`: will determine the time difference between averages

We define two _Maximum_ values, which correspond to the most we will store of a measurement at a given time. This can be multiplied by their respective Epochs to find which length of time information is kept.

- `Maximum Price Stamps`: The maximum amount of `Price Stamps` we will store. Prices will be pruned via FIFO.
- `Maximum Median Stamps`: The maximum amount of `Median Stamps` and Standard Deviations we will store for each asset. Medians will be pruned via FIFO.

### Historic Data Table

The following values will be stored in the storage:

| Name                | Description                                                       | How often it is recorded      | When it is pruned                     |
| ------------------- | ----------------------------------------------------------------- | ----------------------------- | ------------------------------------- |
| `Price Stamp`       | A price for an asset recorded at a given block                    | Every `Historic Stamp Period` | After `Maximum Price Stamps` is met.  |
| `Median Stamp`      | Of a given asset & block, the median of the stored `Price Stamps` | Every `Median Stamp Period`   | After `Maximum Median Stamps` is met. |
| `[]AvgCounter`      | Sum and number of prices per denom                                | Every `Price Update`          | After `Avg Period` is met.            |
| `CurrentAvgCounter` | Index of the most complete AvgCounter used to server queries      | Every `AvgPeriod / AvgShift`  | Never                                 |

Averages are implemented using a list of records:

```go
type AvgCounter struct {
    Sum sdkmath.LegacyDec   // sum of USD value of default denom (eg umee)
    Num uint32    // number of aggregated prices
    Starts uint64 // timestamp
}
```

Each record will store a sum of prices over a moving window. Each such counter will be a rolling sum over the `Avg Period` window. The `AvgCounter.Starts` will determine if we should reset and roll over for a new period.
We will have `AvgPeriod / AvgShift` counters per denom. This is how this can work:

```text
     AvgCounter_1        AvgCounter_1        AvgCounter_1
|-------------------|-------------------|-------------------|---
         AvgCounter_2        AvgCounter_2        AvgCounter_2
----|-------------------|-------------------|-------------------|

\--\
  | AvgShift length
```

### Proposed API

Modules will have access to the following `keeper` functions from the `x/oracle` module:

- `HistoricMedians(denom string, numStamps uint64) []sdkmath.LegacyDec` returns list of last `numStamps` amount of median prices of an asset
- `WithinHistoricDeviation(denom string) (bool, error)` returns whether or not the current price of an asset is within the Standard Deviation around the Median.
- `MedianOfHistoricMedians(denom string, numStamps uint64) (sdkmath.LegacyDec, error)` returns the Median of the all the Medians recorded within the past `numStamps` of medians.
- `AverageOfHistoricMedians(denom string, numStamps uint64) (sdkmath.LegacyDec, error)` returns the Average of all the Medians recorded within the past `numStamps` of medians.
- `MaxOfHistoricMedians(denom string, numStamps uint64) (sdkmath.LegacyDec, error)` returns the Maximum of all the Medians recorded within the past `numStamps` of medians.
- `MinOfHistoricMedians(denom string, numStamps uint64) (sdkmath.LegacyDec, error)` returns the Minimum of all the Medians recorded within the past `numStamps` of medians.
- `HistoricAvgs(denom string) []sdkmath.LegacyDec` returns the most complete of last avg prices for given asset.

### Outcomes

There are a few outcomes to consider:

> This will require a chain upgrade & migration of the `x/oracle` module.
> In order to balance performance and safety, we will have to debate the initial governance parameters.

## Alternative Approaches

### Feeder implementation

This implementation takes the on-chain process of storing historic prices and calculating the median, and takes it off-chain. Validators would then vote on the prices of these assets in a similar fashion to the `x/oracle` voting process.

There are a couple reasons for not doing this:

1. Technical overhead - the x/oracle module would either have to be heavily augmented to deal with this, or we would have to build a new module to deal with the voting process, which would lead to code replication.
2. Latency - validators take time to come to consensus in this model. Currently prices are 5 blocks behind the current APIs, this would make the median another 5 blocks behind that.

## Consequences

The major negative consequence of this is state bloat. Keeping the prices, medians, and standard deviations of each asset supported by the oracle module leads to a lot of bloat.

The major positive is less technical overhead by utilizing an existing module, and allowing for more safety features for `x/leverage`.

### Backwards Compatibility

This will not introduce a new module, and it is relatively backwards compatible. We should allow for governance to disable this feature by voting to set `Stamp Period`, `Prune Period`, and `Median Period` to 0.

### Positive

- Efficient API for the `x/leverage` module to use for safety.
- We can continue listing low-volume assets for collateral.
- Use of an existing module rather than creating a new one.
- Protects against "borrowing + price dump" attacks which are not prevented by disabling low-volume asset collateral
- Storing multiple amount of medians allow clients to do their own calculation: average of medians, median of medians ...

### Negative

- State bloat.
- Additional processing to do during each epoch.

## Further Discussions

1. Can we lessen state bloat without impacting the `x/leverage` module's efficiency?

## Comments

Currently we are planning on keeping the last 6 hours or so of medians, and a longer historic price period. This implementation is meant to be agnostic, so that the `Maximum Medians` and the `Maximum Historic Prices` have no relative constraints.

### Additional Attack Scenario

Another version of this attack goes:

1. Attacker collateralizes USDC and borrows FOO
2. Attacker sells borrowed FOO on exchanges, dumping the price
3. Attacker borrows an even larger amount of FOO using the same collateral, due to the lower price
4. The price of FOO returns to normal, and the value of the FOO exceeds the value of the USDC collateral.
5. Attacker exits the market for a profit.

Alternatively, the Attacker can also dump the price of FOO and withdraw their USDC while reaping the benefit of having the profit from selling FOO and withdrawing their original USDC collateral

### Computational Cost

Where we define:

- `ER` = amount of Active Exchange Rates
- `PS` = amount of Price Stamps
- `D` = amount of denoms

At the end of each `Stamp Period`, we will :

1. Prune `Historic Prices` (Median, Avg, ...) any which are past `Pruning Period`.
1. Collect current set of exchange rates, and copy them into the state with a key of `{Denom}{Block}` and value of TVWAP `ExchangeRate` (of that denom).
1. Prune exchange rates.

Complexity: `D*(PS*(pruning_period/stamp_period) + 2*ER + 1)`

At the end of each `Median Period`, we will :

- For each Denom, collect and sort price stamps.
- Find Median, and store it in state.
- Compute Standard Deviation around the Median, and store it in state.

Given a standard deviation where we have the median of each denom, find the square of each price stamp's distance from the median, sum those values up, and average them:

> STD = ER\*(2H + 2)

The cost of the `Median Period` is:

> (ER x Sort(HP) + 4) + (ER x STD)

## References

- [Twitter Thread on Mango Markets Hack](https://twitter.com/joshua_j_lim/status/1579987648546246658?ref_src=twsrc%5Etfw%7Ctwcamp%5Etweetembed%7Ctwterm%5E1579987648546246658%7Ctwgr%5E4d9aa2fbcc251280df2e9f47258135bac802b986%7Ctwcon%5Es1_&ref_url=https%3A%2F%2Fdecrypt.co%2F111727%2Fsolana-defi-trading-platform-mango-markets-loses-100m-in-hack)
