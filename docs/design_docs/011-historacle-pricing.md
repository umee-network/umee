# Design Doc 011: Historacle Pricing

## Status

Draft

## Abstract

In order to support small-volume assets in a safe manner, we need to be able to calculate the weighted medians of our prices over the past X amount of time (30 days or so). This will allow the leverage module to make decisions around when to allow additional borrowing activity if an asset price is considered abnormal.

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

We define three epoch periods, during which additional computation will be performed:
- `Stamp Period`: duration during which the `x/oracle` module will now "stamp" the set of exchange rates in the state machine until a `Pruning Period` has passed (30 days).
- `Pruning Period`: duration after which the `x/oracle` module will begin to prune expired historic prices.
- `Median Period`: will determine how often the Median and the `Standard Deviation around the Median` are calculated, which will also be stored in the state machine.

These values are stored in state in order to avoid the `x/leverage` module from having to calculate them while making decisions around allowable positions for users to take.

### Proposed API

The `x/leverage` module will have access to the following `keeper` functions from the `x/oracle` module:
- `HistoricMedian(denom) (sdk.Dec, error)` returns the median price of an asset in the last `Pruning Period`
- `WithinHistoricDeviation(denom) (bool, error)` returns whether or not the current price of an asset is within the `Standard Deviation around the Median`.

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

### Negative

- State bloat.
- Additional processing to do during each epoch.

## Further Discussions

1. Can we lessen state bloat without impacting the `x/leverage` module's efficiency?

## Comments

Originally a 24-hour median was thought to be effective enough to defend against this type of attack, but our modeling shows that we would need 30 days for adequate protection.

### Additional Attack Scenario

Another version of this attack goes:

1. Attacker collateralizes USDC and borrows FOO
2. Attacker sells borrowed FOO on exchanges, dumping the price
3. Attacker borrows an even larger amount of FOO using the same collateral, due to the lower price
4. The price of FOO returns to normal, and the value of the FOO exceeds the value of the USDC collateral.
5. Attacker exits the market for a profit.

### Computational Cost

Where we define:

> `Amount of Active Whitelisted Exchange Rates` = *ER*
> `Amount of Historic Prices` = *H*
> `Sorting algorithm` = *Sort*


Each `Stamp Period`, we will :

1. Iterate over `Historic Prices`, and prune any which are past `Pruning Period`.
2. Iterate over the current set of exchange rates, and copy them into the state with a key of `{Denom}{Block}` and value of `ExchangeRate`

> *H* + 2*ER* + 1

Each `Median Period`, we will :

- For each `Active Exchange Rate`, iterate over the historic prices and sort by ExchangeRate
- Find the `Median`, and store it in state
- Find the `Standard Deviation around the Median`, and store it in state

 Given a standard deviation where we have the median of each denom, find the square of each historic price's distance from the median, sum those values up, and average them:

> *STD* = *ER*(2*H* + 2)

The cost of the `Median Period` is:

> (*ER* x *Sort*(*HP*) + 4) + (*ER* x *STD*)

## References

- [Twitter Thread on Mango Markets Hack](https://twitter.com/joshua_j_lim/status/1579987648546246658?ref_src=twsrc%5Etfw%7Ctwcamp%5Etweetembed%7Ctwterm%5E1579987648546246658%7Ctwgr%5E4d9aa2fbcc251280df2e9f47258135bac802b986%7Ctwcon%5Es1_&ref_url=https%3A%2F%2Fdecrypt.co%2F111727%2Fsolana-defi-trading-platform-mango-markets-loses-100m-in-hack)
