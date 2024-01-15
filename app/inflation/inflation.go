package inflation

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/util/bpmath"
	"github.com/umee-network/umee/v6/x/ugov"
)

type Calculator struct {
	UgovKeeperB ugov.ParamsKeeperBuilder
	MintKeeper  mintkeeper.Keeper
}

func (c Calculator) InflationRate(ctx context.Context, minter minttypes.Minter, mintParams minttypes.Params,
	bondedRatio sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ugovKeeper := c.UgovKeeperB(&sdkCtx)
	inflationParams := ugovKeeper.InflationParams()
	maxSupplyAmount := inflationParams.MaxSupply.Amount

	stakingTokenSupply, err := c.MintKeeper.StakingTokenSupply(ctx)
	util.Panic(err)
	if stakingTokenSupply.GTE(maxSupplyAmount) {
		// staking token supply is already reached the maximum amount, so inflation should be zero
		return sdkmath.LegacyZeroDec()
	}

	cycleEnd := ugovKeeper.InflationCycleEnd()
	if sdkCtx.BlockTime().After(cycleEnd) {
		// new inflation cycle is starting, so we need to update the inflation max and min rate
		factor := bpmath.One - inflationParams.InflationReductionRate
		mintParams.InflationMax = factor.MulDec(mintParams.InflationMax)
		mintParams.InflationMin = factor.MulDec(mintParams.InflationMin)
		mintParams.InflationRateChange = fastInflationRateChange(mintParams)
		err := c.MintKeeper.Params.Set(ctx, mintParams)
		util.Panic(err)

		err = ugovKeeper.SetInflationCycleEnd(sdkCtx.BlockTime().Add(inflationParams.InflationCycle))
		util.Panic(err)
		sdkCtx.Logger().Info("inflation min and max rates are updated",
			"inflation_max", mintParams.InflationMax, "inflation_min", mintParams.InflationMin,
			"inflation_cycle_end", sdkCtx.BlockTime().Add(inflationParams.InflationCycle).String(),
		)
	}

	minter.Inflation = minter.NextInflationRate(mintParams, bondedRatio)
	return c.AdjustInflation(stakingTokenSupply, inflationParams.MaxSupply.Amount, minter, mintParams)
}

var two = sdkmath.LegacyNewDec(2)

// inflation rate change = (max_rate - min_rate) * 2
func fastInflationRateChange(p minttypes.Params) sdkmath.LegacyDec {
	return two.Mul(p.InflationMax.Sub(p.InflationMin))
}

// AdjustInflation checks if newly minting coins will execeed the MaxSupply then it will adjust the inflation with
// respect to MaxSupply
func (c Calculator) AdjustInflation(stakingTokenSupply, maxSupply sdkmath.Int, minter minttypes.Minter,
	params minttypes.Params,
) sdkmath.LegacyDec {
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, stakingTokenSupply)
	newMintingAmount := minter.BlockProvision(params).Amount
	newTotalSupply := stakingTokenSupply.Add(newMintingAmount)

	if newTotalSupply.GT(maxSupply) {
		newTotalSupply = maxSupply.Sub(stakingTokenSupply)
		annualProvision := newTotalSupply.Mul(sdkmath.NewInt(int64(params.BlocksPerYear)))
		newAnnualProvision := sdkmath.LegacyNewDec(annualProvision.Int64())
		// AnnualProvisions = Inflation * TotalSupply
		// Mint Coins  = AnnualProvisions / BlocksPerYear
		// Inflation = (New Mint Coins  * BlocksPerYear ) / TotalSupply
		return newAnnualProvision.Quo(sdkmath.LegacyNewDec(stakingTokenSupply.Int64()))
	}
	return minter.Inflation
}
