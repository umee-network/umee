package incentive

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const secondsPerDay = 86400

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxUnbondings:           20,
		UnbondingDurationLong:   secondsPerDay * 14,
		UnbondingDurationMiddle: secondsPerDay * 7,
		UnbondingDurationShort:  secondsPerDay,
		TierWeightMiddle:        sdk.MustNewDecFromStr("0.8"),
		TierWeightShort:         sdk.MustNewDecFromStr("0.5"),
		CommunityFundAddress:    "",
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateMaxUnbondings(p.MaxUnbondings); err != nil {
		return err
	}
	if err := validateUnbondingDuration(p.UnbondingDurationLong); err != nil {
		return err
	}
	if err := validateUnbondingDuration(p.UnbondingDurationMiddle); err != nil {
		return err
	}
	if err := validateUnbondingDuration(p.UnbondingDurationShort); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightMiddle); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightShort); err != nil {
		return err
	}
	if err := validateCommunityFundAddress(p.CommunityFundAddress); err != nil {
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

func validateTierWeight(v sdk.Dec) error {
	if v.IsNegative() {
		return fmt.Errorf("tier weight cannot be negative: %d", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("tier weight cannot exceed 1: %d", v)
	}

	return nil
}

func validateUnbondingDuration(v uint64) error {
	if v == 0 {
		return fmt.Errorf("unbonding duration cannot be zero")
	}

	return nil
}

func validateMaxUnbondings(v uint32) error {
	if v == 0 {
		return fmt.Errorf("max unbondings cannot be zero")
	}

	return nil
}

func validateCommunityFundAddress(addr string) error {
	// Address must be either empty or fully valid
	if addr != "" {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return err
		}
	}

	return nil
}
