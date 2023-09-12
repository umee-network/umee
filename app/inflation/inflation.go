package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/util/bpmath"
	"github.com/umee-network/umee/v6/x/ugov"
)

type Calculator struct {
	UgovKeeperB ugov.ParamsKeeperBuilder
	MintKeeper  MintKeeper
}

func (c Calculator) InflationRate(ctx sdk.Context, minter minttypes.Minter, mintParams minttypes.Params,
	bondedRatio sdk.Dec,
) sdk.Dec {
	ugovKeeper := c.UgovKeeperB(&ctx)
	inflationParams := ugovKeeper.InflationParams()
	maxSupplyAmount := inflationParams.MaxSupply.Amount

	stakingTokenSupply := c.MintKeeper.StakingTokenSupply(ctx)
	if stakingTokenSupply.GTE(maxSupplyAmount) {
		// staking token supply is already reached the maximum amount, so inflation should be zero
		return sdk.ZeroDec()
	}

	cycleEnd := ugovKeeper.InflationCycleEnd()
	if ctx.BlockTime().After(cycleEnd) {
		// new inflation cycle is starting, so we need to update the inflation max and min rate
		factor := bpmath.One - inflationParams.InflationReductionRate
		mintParams.InflationMax = factor.MulDec(mintParams.InflationMax)
		mintParams.InflationMin = factor.MulDec(mintParams.InflationMin)
		// inflation rate change = (max rate - min rate) / 6months
		mintParams.InflationRateChange = sdk.NewDec(2).Mul(mintParams.InflationMax.Sub(mintParams.InflationMin))
		c.MintKeeper.SetParams(ctx, mintParams)

		err := ugovKeeper.SetInflationCycleEnd(ctx.BlockTime().Add(inflationParams.InflationCycle))
		util.Panic(err)
		ctx.Logger().Info("inflation min and max rates are updated",
			"inflation_max", mintParams.InflationMax, "inflation_min", mintParams.InflationMin,
			"inflation_cycle_end", ctx.BlockTime().Add(inflationParams.InflationCycle).String(),
		)
	}

	minter.Inflation = minter.NextInflationRate(mintParams, bondedRatio)
	return c.AdjustInflation(stakingTokenSupply, inflationParams.MaxSupply.Amount, minter, mintParams)
}

// AdjustInflation checks if newly minting coins will execeed the MaxSupply then it will adjust the inflation with
// respect to MaxSupply
func (c Calculator) AdjustInflation(stakingTokenSupply, maxSupply math.Int, minter minttypes.Minter,
	params minttypes.Params,
) sdk.Dec {
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, stakingTokenSupply)
	newMintingAmount := minter.BlockProvision(params).Amount
	newTotalSupply := stakingTokenSupply.Add(newMintingAmount)

	if newTotalSupply.GT(maxSupply) {
		newTotalSupply = maxSupply.Sub(stakingTokenSupply)
		annualProvision := newTotalSupply.Mul(sdk.NewInt(int64(params.BlocksPerYear)))
		newAnnualProvision := sdk.NewDec(annualProvision.Int64())
		// AnnualProvisions = Inflation * TotalSupply
		// Mint Coins  = AnnualProvisions / BlocksPerYear
		// Inflation = (New Mint Coins  * BlocksPerYear ) / TotalSupply
		return newAnnualProvision.Quo(sdk.NewDec(stakingTokenSupply.Int64()))
	}
	return minter.Inflation
}
