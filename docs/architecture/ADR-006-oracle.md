# ADR 006: Oracle

## Changelog

- November 29th, 2021: Initial Draft (@adamewozniak)

## Status

Proposed

## Context

Umee needs an oracle to determine the price of assets.

From section 5.1 of the [Umee Whitepaper](https://umee.cc/umee-whitepaper/):

> Oracle reads asset price and updates the state to the Asset Facility Coordination
Protocol

We've already decided to import a large chunk of this from [Terra's oracle](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html), although a few parts of this are specific to Terra's protocol and do not need to be implemented.

## Alternative Approaches

- Cloning the x/oracle module completely. This would leave our code dirty, and we'd later have issues interfacing with Terra's [Cross Exchange Rate](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html#compute-cross-exchange-rate-using-reference-terra), since it's designed for getting the exchange rate of only Terra.
- Using something like Band or Chainlink. This would be additional overhead, and we'd [have less control](https://github.com/umee-network/umee/issues/97#issuecomment-923914840) over how our oracle works.

## Decision

We'd like to use the concepts introduced in [Terra's Oracle](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html#concepts), but with a few modifications :

- Combine the `price-feeder` and `feeder` into a single binary [Ref](https://github.com/umee-network/umee/issues/97#issuecomment-939610302)
- Only support MsgAggregateExchangeRate(Pre)Vote, i.e. not allow individual price updates [Ref](https://github.com/umee-network/umee/issues/97#issuecomment-939610302)
- Skip the logic for the [Cross Exchange Rate](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html#compute-cross-exchange-rate-using-reference-terra), and record simplified exchange rates like `ATOMUSDT`
- Remove support for the [Tobin Tax](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-market.html#tobintax), a Terra-specific fee for when users spot trade.

## Detailed Design

- Terra's design for the voting procedure as documented [here](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html#voting-procedure)

## Consequences

### Positive

- More control over how our oracle works, so that we can integrate it with the `x/leverage` module how we like
- Less overhead
- Cleaner code

### Negative

- More time spent on `x/oracle` development

### Neutral
- Requires us to [evaluate the whitelist](https://github.com/umee-network/umee/issues/225) for the oracle as we're not sure whether or not we want to accept all of the operators' exchange rates

## References

- [Terra Oracle Spec](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-oracle.html)
- [Terra's Tobin Tax](https://docs.terra.money/Reference/Terra-core/Module-specifications/spec-market.html#tobintax)
- [Discussion on Oracle Decision](https://github.com/umee-network/umee/issues/97#issuecomment-923914840)