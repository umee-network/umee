package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/util/bpmath"
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

	cycleEnd := ugovKeeper.GetInflationCycleEnd()
	if ctx.BlockTime().After(cycleEnd) {
		// new inflation cycle is starting , so we need to update the inflation max and min rate
		factor := bpmath.One - inflationParams.InflationReductionRate
		mintParams.InflationMax = factor.MulDec(mintParams.InflationMax)
		mintParams.InflationMin = factor.MulDec(mintParams.InflationMin)

		c.MintKeeper.SetParams(ctx, mintParams)

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
		// Inflation = (New Mint Coins  * BlocksPerYear ) / TotalSupply
		return sdk.NewDec(newAnnualProvision.Quo(totalSupply).Int64())
	}
	return minter.Inflation
}
