package inflation

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/util/bpmath"
)

type Calculator struct {
	UgovKeeperB UGovBKeeperI
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

	icst, err := ugovKeeper.GetInflationCycleStart()
	util.Panic(err)

	// Initially inflation_cycle start time is zero
	// Once chain start inflation cycle start time will be inflation rate change executed block time
	if ctx.BlockTime().After(icst.Add(inflationParams.InflationCycle)) {
		// inflation cycle is completed , so we need to update the inflation max and min rate
		// inflationReductionRate = 25 / 100 = 0.25
		inflationReductionRate := inflationParams.InflationReductionRate.ToDec().Quo(bpmath.FixedBP(100).ToDec())
		// InflationMax = PrevInflationMax * ( 1 - 0.25)
		mintParams.InflationMax = mintParams.InflationMax.Mul(sdk.OneDec().Sub(inflationReductionRate))
		// InflationMin = PrevInflationMin * ( 1 - 0.25)
		mintParams.InflationMin = mintParams.InflationMin.Mul(sdk.OneDec().Sub(inflationReductionRate))

		// update the changed inflation min and max rates
		c.MintKeeper.SetParams(ctx, mintParams)

		// update the executed time of inflation cycle
		err := ugovKeeper.SetInflationCycleStart(ctx.BlockTime())
		util.Panic(err)
		ctx.Logger().Info("inflation min and max rates are updated",
			"inflation_max", mintParams.InflationMax, "inflation_min", mintParams.InflationMin,
			"inflation_cycle_start", ctx.BlockTime().String(),
		)
	}

	minter.Inflation = minter.NextInflationRate(mintParams, bondedRatio)
	return c.AdjustInflation(totalSupply, inflationParams.MaxSupply.Amount, minter, mintParams)
}

// AdjustInflation check if newly minting coins will execeed the MaxSupply then it will adjust the inflation with
// respect to MaxSupply
func (c Calculator) AdjustInflation(totalSupply, maxSupply math.Int, minter minttypes.Minter,
	params minttypes.Params) sdk.Dec {
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	newSupply := minter.BlockProvision(params).Amount
	newTotalSupply := totalSupply.Add(newSupply)

	if newTotalSupply.GT(maxSupply) {
		newTotalSupply = maxSupply.Sub(totalSupply)
		annualProvision := newTotalSupply.Mul(sdk.NewInt(int64(params.BlocksPerYear)))
		newAnnualProvision := sdk.NewDec(annualProvision.Int64())
		// AnnualProvisions = Inflation * TotalSupply
		// Mint Coins  = AnnualProvisions / BlocksPerYear
		// so get the new Inflation
		// Inflation = (New Mint Coins  * BlocksPerYear ) / TotalSupply
		return newAnnualProvision.Quo(sdk.NewDec(totalSupply.Int64()))
	}
	return minter.Inflation
}
