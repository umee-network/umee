package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/umee-network/umee/v5/util"
	ugov "github.com/umee-network/umee/v5/x/ugov"
)

// BeginBlock overrides the mint module BeginBlock.
func BeginBlock(ctx sdk.Context, ugovKeeper UGovKeeper, mintKeeper Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	// inflation rate change params
	lp := ugovKeeper.InflationParams()
	// mint module params
	mintParams := mintKeeper.GetParams(ctx)

	totalStakingSupply := mintKeeper.StakingTokenSupply(ctx)
	if totalStakingSupply.GTE(lp.MaxSupply.Amount) {
		// supply is already reached the maximum amount, so no more minting the staking token
		return
	}

	// fetch stored minter & params
	minter := mintKeeper.GetMinter(ctx)
	bondedRatio := mintKeeper.BondedRatio(ctx)
	// recalculate inflation rate
	minter.Inflation = InflationCalculationFn(ctx, ugovKeeper, mintKeeper, lp, mintParams, bondedRatio, minter.Inflation)
	minter.AnnualProvisions = minter.NextAnnualProvisions(mintParams, totalStakingSupply)
	mintKeeper.SetMinter(ctx, minter)

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(mintParams)
	mintedCoins := sdk.NewCoins(mintedCoin)

	err := mintKeeper.MintCoins(ctx, mintedCoins)
	util.Panic(err)

	// send the minted coins to the fee collector account
	err = mintKeeper.AddCollectedFees(ctx, mintedCoins)
	util.Panic(err)

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}

func InflationCalculationFn(ctx sdk.Context, ugovKeeper UGovKeeper, mintKeeper Keeper,
	lp ugov.InflationParams, params types.Params, bondedRatio sdk.Dec, currentInflation sdk.Dec) sdk.Dec {

	icst, err := ugovKeeper.GetInflationCycleStart()
	util.Panic(err)

	// Initially inflation_cycle start time is zero
	// Once chain start inflation cycle start time will be inflation rate change executed block time
	if ctx.BlockTime().After(icst.Add(lp.InflationCycle)) {
		// inflation cycle is completed , so we need to update the inflation max and min rate
		params.InflationMax = params.InflationMax.Mul(sdk.OneDec().Sub(lp.InflationReductionRate))
		params.InflationMin = params.InflationMin.Mul(sdk.OneDec().Sub(lp.InflationReductionRate))

		// update the changed inflation min and max rates
		mintKeeper.SetParams(ctx, params)

		// update the executed time of inflation cycle
		err := ugovKeeper.SetInflationCycleStart(ctx.BlockTime())
		util.Panic(err)
		ctx.Logger().Info("inflation min and max rates are updated",
			"inflation_max", params.InflationMax, "inflation_min", params.InflationMin,
			"inflation_cycle_start", ctx.BlockTime().String(),
		)
	}

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneDec().
		Sub(bondedRatio.Quo(params.GoalBonded)).
		Mul(params.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(sdk.NewDec(int64(params.BlocksPerYear)))

	// adjust the new annual inflation for this next cycle
	inflation := currentInflation.Add(inflationRateChange) // note inflationRateChange may be negative
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}
	if inflation.LT(params.InflationMin) {
		inflation = params.InflationMin
	}

	return inflation
}
