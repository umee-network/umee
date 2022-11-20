package incentive

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const secondsPerDay = 86400

var (
	defaultMaxUnbondings           = uint32(20)
	defaultUnbondingDurationShort  = uint64(secondsPerDay)
	defaultUnbondingDurationMiddle = uint64(secondsPerDay * 7)
	defaultUnbondingDurationLong   = uint64(secondsPerDay * 14)
	defaultTierWeightShort         = sdk.MustNewDecFromStr("0.5")
	defaultTierWeightMiddle        = sdk.MustNewDecFromStr("0.8")
)

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxUnbondings:           defaultMaxUnbondings,
		UnbondingDurationLong:   defaultUnbondingDurationLong,
		UnbondingDurationMiddle: defaultUnbondingDurationMiddle,
		UnbondingDurationShort:  defaultUnbondingDurationShort,
		TierWeightMiddle:        defaultTierWeightMiddle,
		TierWeightShort:         defaultTierWeightShort,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateMaxUnbondings(p.MaxUnbondings); err != nil {
		return err
	}
	if err := validateLockDuration(p.UnbondingDurationLong); err != nil {
		return err
	}
	if err := validateLockDuration(p.UnbondingDurationMiddle); err != nil {
		return err
	}
	if err := validateLockDuration(p.UnbondingDurationShort); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightMiddle); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightShort); err != nil {
		return err
	}
	if p.UnbondingDurationLong < p.UnbondingDurationMiddle || p.UnbondingDurationMiddle < p.UnbondingDurationShort {
		return ErrUnbondingTierOrder
	}
	if p.TierWeightMiddle.LT(p.TierWeightShort) {
		return ErrUnbondingWeightOrder
	}
	return nil
}

func validateTierWeight(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("tier weight cannot be negative: %d", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("tier weight cannot exceed 1: %d", v)
	}

	return nil
}

func validateLockDuration(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("lock duration cannot be zero")
	}

	return nil
}

func validateMaxUnbondings(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max unbondings cannot be zero")
	}

	return nil
}
