
# UMEE Tokenomic Restructuring Proposal

## Introduction

The UMEE Tokenomic Restructuring proposal aims to optimize the tokenomics of the $UMEE token, which is issued following the standard Cosmos SDK mechanism. The goal is to progressively limit the token issuance in an algorithmic way, taking into account various factors such as network security, adoption of liquid staking, and additional income sources for stakers through interchain security and chain-based revenue streams.

## Design Overview

1. The $UMEE 2.0 Tokenomics will adopt dynamic inflation mechanism based on the bonding rate of the network. This allows the Umee chain to naturally adjust the inflation rate to drive staking activities and reach the targeted bonding rate.

2. To initiate the new tokenomics, the maximum and minimum inflation rates will decrease by 20%.

   - Current min and max inflation rates: 7% - 14%
   - New max and min inflation rates at the start: 5.6% - 11.2%

3. The max and min inflation rates will decrease by 25% every two years in inflation cycles. This is in anticipation of the growth of the Umee ecosystem and wider adoption of liquid staking.

   - Each inflation cycle will have a variable yearly total amount of newly emitted tokens.
   - At the end of each inflation cycle, the `min_inflation_rate`, `max_inflation_rate` will be decreased by 25%.
    See [Mint Proto](https://github.com/cosmos/cosmos-sdk/blob/v0.46.13/proto/cosmos/mint/v1beta1/mint.proto) for max and mint inflation rates
   - The inflation rate change speed will be accelerated from 1 year to 6 months.

4. The total $UMEE supply will be capped at 21 billion tokens.

   - Inflation will be reduced during each inflation cycle until minting stops when the maximum supply is reached.
   - Minting will resume if the supply drops below 21 billion, and it will stop again once the 21 billion mark is reached.
   - The supply will measured by total supply of $UMEE on network.

5. All inflationary tokens will be distributed to stakers and delegators, similar to the current mechanism.

6. **Umee governance will have the power to vote and modify these inflation parameters if the external environment significantly changes, necessitating a different approach.**

## Technical Specification

The Inflation Params will be stored in the `x/ugov` module:

```go
// maximum supply in base denom (uumee) 
max_supply       sdk.Coin
// length of the inflation cycle
inflation_cycle_duration  time.Duration
// in basis points. 12 corresponds to 0.012. New inflation is calculated as:
// old_inflation * (10'000 - inflation_reduction_rate)/10'000
inflation_reduction_rate bpmath.FixedBP
```

See [Ugov Proto](https://github.com/umee-network/umee/blob/main/proto/umee/ugov/v1/ugov.proto) for inflation params

### InflationCalculationFn

The inflation calculation logic will be implemented as follows:

```go pseudocode
Input: sdk.Context, minter, mint module params , bondedRatio
Output: Inflation 

Function inflationRate(ctx , minter, mintParams, bondedRatio):

    INFLATION_PARAMS = // GET INFLATION_PARAMS FROM UGOV MODULE 
    MAX_SUPPLY = // MAX SUPPLY OF MINTING DENOM FROM INFLATION_PARAMS

    TOTAL_TOKENS_SUPPLY =  // TOTAL SUPPLY OF MINTING DENOM 
    If TOTAL_TOKENS_SUPPLY.GTE(MAX_SUPPLY) :
        // supply has already reached the maximum amount, so inflation should be zero
        Return sdk.ZeroDec()

    INFLATION_CYCLE_END_TIME = // CURRENT INFLATION CYCLE END TIME FROM UGOV STORE 

    If ctx.BlockTime > INFLATION_CYCLE_END_TIME:
        // new inflation cycle is starting, so we need to update the inflation max and min rate
        factor = 1 - INFLATION_PARAMS.InflationReductionRate // 1 - 0.25 = 0.75
        mintParams.InflationMax = factor.MulDec(mintParams.InflationMax)
        mintParams.InflationMin = factor.MulDec(mintParams.InflationMin)

        MintKeeper.SetParams(ctx, mintParams)

        // SET CURRENT NEW INFLATION CYCLE END TIME IN UGOV STORE 
        CYCLE_END_TIME = ctx.BlockTime().Add(INFLATION_PARAMS.InflationCycle)

    Inflation = minttypes.DefaultInflationCalculationFn(ctx, minter, mintParams, bondedRatio)
    
    // if the newly minted coins will exceed the MaxSupply, and adjusts the inflation accordingly.
    MINTING_COINS =( Inflation * TOTAL_TOKENS_SUPPLY )/ mintParams.BlocksPerYear
    IF (MINTING_COINS + TOTAL_TOKENS_SUPPLY) > MAX_SUPPLY:
        MINTING_COINS = MAX_SUPPLY - TOTAL_TOKENS_SUPPLY
        Return Inflation = (MINTING_COINS * mintParams.BlocksPerYear) / TOTAL_TOKENS_SUPPLY
    
    Return Inflation
```

See implementation [here](https://github.com/umee-network/umee/blob/main/app/inflation/inflation.go).
