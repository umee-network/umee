
# UMEE Tokenomic Restructuring Proposal

## Introduction

The UMEE Tokenomic Restructuring proposal aims to optimize the tokenomics of the $UMEE token, which is issued following the standard Cosmos SDK mechanism. The goal is to progressively limit the token issuance in an algorithmic way, taking into account various factors such as network security, adoption of liquid staking, and additional income sources for stakers through interchain security and chain-based revenue streams.

## Design Overview

1. The $UMEE 2.0 Tokenomics will adopt a similar Cosmos dynamic inflation mechanism based on the bonding rate of the network. This allows the Umee chain to naturally adjust the inflation rate to drive staking activities and reach the targeted bonding rate.

2. To initiate the new tokenomics, the maximum and minimum inflation rates will decrease by 25%. Subsequent decreases will occur every two years.

   - Current min and max inflation rates: 7% - 14%
   - New max and min inflation rates at the start: 5.6% - 11.2%

3. The max and min inflation rates will decrease by 25% every two years in inflation cycles. This is in anticipation of the growth of the Umee ecosystem and wider adoption of liquid staking.

   - Each inflation cycle will have a variable yearly total amount of newly emitted tokens.
   - At the end of each inflation cycle, the `min_inflation_rate`, `max_inflation_rate` will be decreased by 25%.
   - The inflation rate change speed will be accelerated from 1 year to 6 months.

4. The total $UMEE supply will be capped at 21 billion tokens.

   - Inflation will be reduced during each inflation cycle until minting stops when the maximum supply is reached.
   - Minting will resume if the supply drops below 21 billion, and it will stop again once the 21 billion mark is reached.
   - The supply will be `StakingTokenSupply`.

5. All inflationary tokens will be distributed to stakers and delegators, similar to the current mechanism.

6. **Umee governance will have the power to vote and modify the `inflation params`` if the external environment significantly changes, necessitating a different approach.**

## Technical Specification

The Inflation Params to be stored in the `x/ugov` module:

```go
// max supply based on the bank/QueryTotalSupplyRequest 
max_supply       sdk.Coin
// length of the inflation cycle as described in the Design
inflation_cycle_duration  time.Duration
// in basis points. 12 corresponds to 0.012. New inflation is calculated as:
// old_inflation * (10'000 - inflation_reduction_rate)/10'000
inflation_reduction_rate sdk.Dec
```

1. `max_supply` : This parameter determines the maximum supply of token. $UMEE max supply is 21 billion tokens.
2. `inflation_cycle_duration` : This paramter determines the duration of inflation cycle.
3. `inflation_reduction_rate` (100bp to 10'000bp): This parameter determines the rate at which the inflation rate is reduced. A value of 100bp indicates 1%, 0.1 corresponds to a 10% reduction rate, and 0.01 represents a 1% reduction rate.

### InflationCalculationFn

The inflation calculation logic will be implemented as follows:

```go
type Calculator struct {
    UgovKeeperB ugovkeeper.Builder
    MintKeeper  MintKeeper
}

func (c Calculator) InflationRate(ctx sdk.Context, minter minttypes.Minter, mintParams minttypes.Params,
    bondedRatio sdk.Dec) sdk.Dec {

    ugovKeeper := c.UgovKeeperB.Keeper(&ctx)
    inflationParams := ugovKeeper.InflationParams()
    maxSupplyAmount := inflationParams.MaxSupply.Amount

    totalSupply := c.MintKeeper.StakingTokenSupply(ctx)
    if totalSupply.GTE(maxSupplyAmount) {
        // supply has already reached the maximum amount, so inflation should be zero
        return sdk.ZeroDec()
    }

    cycleEnd, err := ugovKeeper.GetInflationCycleEnd()
    util.Panic(err)

    if ctx.BlockTime().After(cycleEnd) {
        // new inflation cycle is starting, so we need to update the inflation max and min rate
        factor := bpmath.One - inflationParams.InflationReductionRate
        mintParams.InflationMax = factor.MulDec(mintParams.InflationMax)
        mintParams.InflationMin = factor.MulDec(mintParams.InflationMin)

        c.MintKeeper.SetParams(ctx, mintParams)

        err := ugovKeeper.SetInflationCycleEnd(ctx.BlockTime().Add(inflationParams.InflationCycle))
        util.Panic(err)
    }

    minter.Inflation = minttypes.DefaultInflationCalculationFn(ctx, minter, mintParams, bondedRatio)
    return c.AdjustInflation(totalSupply, inflationParams.MaxSupply.Amount, minter, mintParams)
}

// AdjustInflation checks if the newly minted coins will exceed the MaxSupply, and adjusts the inflation accordingly.
func (c Calculator) AdjustInflation(totalSupply, maxSupply math.Int, minter minttypes.Minter,
    params minttypes.Params) sdk.Dec {
    minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
    newSupply := minter.BlockProvision(params).Amount
    newTotalSupply := totalSupply.Add(newSupply)
    if newTotalSupply.GT(maxSupply) {
        newTotalSupply = maxSupply.Sub(totalSupply)
        newAnnualProvision := newTotalSupply.Mul(sdk.NewInt(int64(params.BlocksPerYear)))
        // AnnualProvisions = Inflation * TotalSupply
        // Mint Coins = AnnualProvisions / BlocksPerYear
        // Inflation = (New Mint Coins * BlocksPerYear) / TotalSupply
        return sdk.NewDec(newAnnualProvision.Quo(totalSupply).Int64())
    }
    return minter.Inflation
}
```

## Conclusion

The UMEE Tokenomic Restructuring proposal aims to optimize the $UMEE token economics, providing a dynamic and adaptive mechanism for inflation adjustment based on the network's bonding rate. The proposed changes will support the growth of the Umee ecosystem and enhance the overall staking experience for validators and delegators. Umee governance retains the ability to adapt the inflation schedule if required to respond to external changes in the future.
