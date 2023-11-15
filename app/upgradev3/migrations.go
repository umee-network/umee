package upgradev3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var minCommissionRate = sdk.MustNewDecFromStr("0.05")

// UpdateMinimumCommissionRateParam is update the minimum commission rate param of staking.
func UpdateMinimumCommissionRateParam(ctx sdk.Context, keeper StakingKeeper) (sdk.Dec, error) {
	params := keeper.GetParams(ctx)
	params.MinCommissionRate = minCommissionRate
	keeper.SetParams(ctx, params)

	return minCommissionRate, nil
}

// SetMinimumCommissionRateToValidators is update the minimum commission rate to the validators rate
// whose commission rate is below the minimum commission rate.
func SetMinimumCommissionRateToValidators(ctx sdk.Context, keeper StakingKeeper, minCommissionRate sdk.Dec) error {
	validators := keeper.GetAllValidators(ctx)

	for _, validator := range validators {
		if validator.Commission.Rate.IsNil() || validator.Commission.Rate.LT(minCommissionRate) {
			if err := keeper.BeforeValidatorModified(ctx, validator.GetOperator()); err != nil {
				return err
			}

			validator.Commission.Rate = minCommissionRate
			validator.Commission.UpdateTime = ctx.BlockTime()

			keeper.SetValidator(ctx, validator)
		}
	}

	return nil
}
