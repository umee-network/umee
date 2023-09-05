# Design Doc 12: UMEE Inflation v2

## Changelog

- August, 2023: Initial Design (Ming Duan, @robert-zaremba)

## Status

Accepted

## Abstract

Umee v2 Inflation model introduces market cap for the UMEE supply and inflation cycles when the inflation is updated.

## Context

Today, the token issuance follows standard Cosmos SDK mechanism. We want to progressively limit the issuance in an algorithmic way. The UMEE tokenomics needs to be optimized with the following considerations:

- The current network parameters - min inflation 7% max inflation 14% - results in high spending on chain security.
- The adoption of liquid staking will drive more staking with a lower staking yield required.
- Interchain security and chain based revenue streams can become additional source of income for stakers.
- Set the hard cap to influence more the market expectations of the Umee supply.

## Specification

We set the max inflation to 12 billions UMEE (the supply today is around 12.2 billions):

- Block rewards are always adjusted (zeroed or reduced if necessary) so total supply never exceeds 12 billion UMEE.
  - Supply is measured based on `bank/QueryTotalSupplyRequest`.
- Minting resumes if supply drops below 21B (and stops again once 21B is reached).
- Burn events will reduce the UMEE supply, enabling more staking rewards, according to the inflation rules. Burning events were included in the whitepaper, but the exact specification will be the subject of another design document.

The $UMEE v2 Inflation follows a similar Cosmos dynamic inflation mechanism based on the bonding rate in the `x/staking` module, to allow the Umee chain to naturally adjust inflation rate to drive staking activities to reach the target bonding rate.

On top of that we introduce _inflation cycle_, which is initially set to 2 years and controlled by the UMEE governance. In each inflation cycle, the yearly total amount of newly emitted tokens is a variable as per the `x/staking` Cosmos SDK module.
At the end of each inflation cycle, the `x/mint` module `inflation_min` and `inflation_max` parameters are automatically decreased by the `infaltion_reduction_rate` parameter, which is initially set to 25%.

Finally, we accelerate the `x/mint` [inflation rate change](https://github.com/cosmos/cosmos-sdk/blob/v0.47.2/x/mint/README.md#nextinflationrate) speed from 1 year to **6 months**.

Umee governance at any time can vote and modify the inflation schedule if the external environment changes materially and requires a different plan.

### Technical Implementation

The new inflation mechanism is build on top of the existing `x/mint` module by providing [`InflationCalculationFn`](https://github.com/cosmos/cosmos-sdk/blob/v0.46.14/x/mint/types/genesis.go#L12) to the `x/mint` module constructor. That function will use Umee `x/ugov` module which will control and store the following parameters:

```protobuf
message InflationParams {
  // max_supply is the maximum supply for liquidation.
  cosmos.base.v1beta1.Coin max_supply = 1;

  // inflation_cycle duration after which inflation rates are changed.
  google.protobuf.Duration inflation_cycle = 2;

  // inflation_reduction_rate for every inflation cycle.
  uint32 inflation_reduction_rate = 3;
}
```

The Umee `x/ugov` will be extended by the following Msg Server Message:

```protobuf
message MsgGovUpdateInflationParams {
  option (cosmos.msg.v1.signer) = "authority";

  // authority must be the address of the governance account.
  string          authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  InflationParams params    = 2;
}
```

The inflation cycle change is handled by the `InflationCalculationFn` function, implemented by the `inflation.Calculator` object (in the `app/inflation` package). The `Calculator` will handle the x/ugov module dependency.

## Alternative Approaches

We were discussing more advanced mechanism, which would define the staking rewards per user based on his usage of the `x/leverage` and `x/metoken` modules (notably amount of supplying assets).

Another possible modification is to set a constant inflation rate during each Inflation Cycle. It would make the calculation simpler, and make the yearly new supply more predictable. However, it will remove the dynamic nature of the traditional `x/staking` which adapts the block rewards based on the UMEE staking target.

## Test Case

- Unit tests to verify the Inflation Cycle changes
- Unit tests to verify the Umee

## Consequences

- Umee supply is predictable.
- Umee sell pressure is reduced.
- Umee inflation is more dynamic.

### Backwards Compatibility

- The solution is backwards compatible with the existing Cosmos SDK stack. We don't need to fork any Cosmos SDK module implementation, nor overwrite it.
- The implementation adds new method to the UMEE `x/ugov` module and new parameter set, which is additive.

## Further Discussions

We are exploring the burn mechanisms and ways how to better react on the market conditions.

## References

- Cosmos SDK [x/mint documentation](https://github.com/cosmos/cosmos-sdk/blob/v0.47.2/x/mint/README.md)
- Osmosis [Thirdening](https://medium.com/@ne_fertiti/what-is-osmosis-thirdening-how-it-affects-your-lp-staking-returns-c750f89efb14)
