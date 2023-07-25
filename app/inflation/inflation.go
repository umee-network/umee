package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/umee-network/umee/v5/util"
	ugovkeeper "github.com/umee-network/umee/v5/x/ugov/keeper"
)

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
		// supply is already reached the maximum amount, so inflation should be zero
		return sdk.ZeroDec()
	}

	cycleEnd, err := ugovKeeper.GetInflationCycleEnd()
	util.Panic(err)

	// Initially inflation_cycle end time is zero
	// Once chain start inflation cycle end time will be  executed block time + cycle duration
	if ctx.BlockTime().After(cycleEnd) {
		// inflation cycle is completed , so we need to update the inflation max and min rate
		// inflationReductionRate = 25 / 100 = 0.25
		inflationReductionRate := inflationParams.InflationReductionRate.ToDec().Quo(sdk.NewDec(100))
		// InflationMax = PrevInflationMax * ( 1 - 0.25)
		mintParams.InflationMax = mintParams.InflationMax.Mul(sdk.OneDec().Sub(inflationReductionRate))
		// InflationMin = PrevInflationMin * ( 1 - 0.25)
		mintParams.InflationMin = mintParams.InflationMin.Mul(sdk.OneDec().Sub(inflationReductionRate))

		// update the changed inflation min and max rates
		c.MintKeeper.SetParams(ctx, mintParams)

		// update the end time of current inflation cycle
		err := ugovKeeper.SetInflationCycleEnd(ctx.BlockTime().Add(inflationParams.InflationCycle))
		util.Panic(err)
		ctx.Logger().Info("inflation min and max rates are updated",
			"inflation_max", mintParams.InflationMax, "inflation_min", mintParams.InflationMin,
			"inflation_cycle_end", ctx.BlockTime().Add(inflationParams.InflationCycle).String(),
		)
	}

	minter.Inflation = minttypes.DefaultInflationCalculationFn(ctx, minter, mintParams, bondedRatio)
	return c.adjustInflation(totalSupply, inflationParams.MaxSupply.Amount, minter, mintParams)
}

// adjustInflation check if newly minting coins will execeed the MaxSupply then it will adjust the inflation with
// respect to MaxSupply
func (c Calculator) adjustInflation(totalSupply, maxSupply math.Int, minter minttypes.Minter,
	params minttypes.Params) sdk.Dec {
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	newSupply := minter.BlockProvision(params).Amount
	newTotalSupply := totalSupply.Add(newSupply)
	if newTotalSupply.GT(maxSupply) {
		newTotalSupply = maxSupply.Sub(totalSupply)
		newAnnualProvision := newTotalSupply.Mul(sdk.NewInt(int64(params.BlocksPerYear)))
		// AnnualProvisions = Inflation * TotalSupply
		// Mint Coins  = AnnualProvisions / BlocksPerYear
		// so get the new Inflation
		// Inflation = (New Mint Coins  * BlocksPerYear ) / TotalSupply
		return sdk.NewDec(newAnnualProvision.Quo(totalSupply).Int64())
	}
	return minter.Inflation
}
