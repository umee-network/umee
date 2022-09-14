package upgrades

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingKeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

var (
	minCommissionRate = sdk.MustNewDecFromStr("0.05")
)

// UpdateMinimumCommissionRateParam is update the minimum commission rate param of staking.
func UpdateMinimumCommissionRateParam(ctx sdk.Context, keeper *stakingKeeper.Keeper) (sdk.Dec, error) {
	params := keeper.GetParams(ctx)
	// update the minCommissionRate param
	params.MinCommissionRate = minCommissionRate

	keeper.SetParams(ctx, params)

	return minCommissionRate, nil
}

// SetMinimumCommissionRateToValidatros is update the minimum commission rate to the validators rate
// whose commission rate is below the minimum commission rate.
func SetMinimumCommissionRateToValidatros(
	ctx sdk.Context, keeper *stakingKeeper.Keeper, minCommissionRate sdk.Dec) error {
	validators := keeper.GetAllValidators(ctx)

	for _, validator := range validators {
		if validator.Commission.Rate.GTE(minCommissionRate) {
			continue
		}

		if err := keeper.BeforeValidatorModified(ctx, validator.GetOperator()); err != nil {
			return err
		}

		validator.Commission.Rate = minCommissionRate
		validator.Commission.UpdateTime = ctx.BlockTime()

		keeper.SetValidator(ctx, validator)
	}

	return nil
}
