package inflation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	ugovkeeper "github.com/umee-network/umee/v5/x/ugov/keeper"
)

type Calculator struct {
	UgovKeeperB ugovkeeper.Builder
	MintKeeper  MintKeeper
}

func (c Calculator) InflationRate(ctx sdk.Context, minter minttypes.Minter, params minttypes.Params, bondedRatio sdk.Dec) sdk.Dec {
	maxSupply, _ := sdk.NewIntFromString("21_000_000_000_000_000000")

	totalSupply := c.MintKeeper.StakingTokenSupply(ctx)
	if totalSupply.GTE(maxSupply) {
		// supply is already reached the maximum amount, so inflation should be zero
		return sdk.ZeroDec()
	}

	// TODO: here we need to use a new inflation function and check if we need to go to the
	// next inflation cycle
	// minter.Inflation = c.calculateInflation()
	minter.Inflation = minttypes.DefaultInflationCalculationFn(ctx, minter, params, bondedRatio)
	return readjustSupply(totalSupply, maxSupply, minter, params)
}

func (c Calculator) calculateInflation() sdk.Dec {
	panic("todo")
}

// TODO: add unit tests to this function
func readjustSupply(totalSupply, maxSupply sdk.Int, minter minttypes.Minter, params minttypes.Params) sdk.Dec {
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	newSupply := minter.BlockProvision(params).Amount
	newTotalSupply := totalSupply.Add(newSupply)
	if newTotalSupply.GT(maxSupply) {
		overdraft := newTotalSupply.Sub(maxSupply)
		maxNewSupply := newSupply.Sub(overdraft)
		factor := sdk.NewDecFromInt(maxNewSupply).QuoInt(newSupply)
		minter.Inflation = minter.Inflation.Mul(factor)
	}

	return minter.Inflation
}
