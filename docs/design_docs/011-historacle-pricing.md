# Design Doc 011: Historacle Pricing

## Status

Draft

## Abstract

In order to support small-volume assets in a safe manner, we need to be able to calculate the weighted medians of our prices over the past X amount of time (30 days or so). This will allow the leverage module to make decisions around when to allow liquidations if an asset price is considered abnormal.

## Context

Right now, the volume of assets such as Juno, Umee, and Osmo are so low that it would be relatively easy to use the leverage protocol on UMEE to manipulate the prices of these assets and game the system.

The attack goes:

1. Account A deposits & collateralizes a large amount of USDC.
2. Account A uses collateral to borrow a lot of the low-volume token (here, FOO).
3. Account A sells the tokens to Account B, inflating the price of FOO.
4. Account B collateralizes inflated FOO, and borrows USDC.
5. Account B exits the market for a profit.

In order to avoid this, and to continue with our goal of allowing users to collateralize low-volume assets, we need to have a safety net.

Currently, this is defended against by disallowing the use of `Umee` as collateral, however, we would like to re-enable that and list other assets with low volume for collateral.

Historacle pricing will provide an API for the leverage module to tell when prices have changed in an abnormal fashion and avoid such events.

## Specification

Each `Stamp Period`, the `x/oracle` module will now "stamp" the set of exchange rates in memory, and store it until a `Pruning Period` has passed (30 days). Another epoch, `Median Period`, will determine how often the Median and the `Standard Deviation around the Median` are calculated, which will also be stored in memory.

These values are stored in state in order to avoid the `x/leverage` module from having to calculate them while making decisions around allowable positions for users to take.

### Proposed API

The leverage protocol will have access to a `Keeper` function from the `x/oracle` module, which will take a `denom`, and will return whether or not the current price of that asset is within the `Standard Deviation around the Median`.

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

### Negative

- State bloat.
- Additional processing to do during each epoch.

## Further Discussions

1. Can we lessen state bloat without impacting the `x/leverage` module's efficiency?

## Comments

Originally a 24-hour median was thought to be effective enough to defend against this type of attack, but our modeling shows that we would need 30 days for adequate protection.

## References

- [Twitter Thread on Mango Markets Hack](https://twitter.com/joshua_j_lim/status/1579987648546246658?ref_src=twsrc%5Etfw%7Ctwcamp%5Etweetembed%7Ctwterm%5E1579987648546246658%7Ctwgr%5E4d9aa2fbcc251280df2e9f47258135bac802b986%7Ctwcon%5Es1_&ref_url=https%3A%2F%2Fdecrypt.co%2F111727%2Fsolana-defi-trading-platform-mango-markets-loses-100m-in-hack)
