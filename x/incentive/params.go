package incentive

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
)

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxUnbondings:      10,
		UnbondingDuration:  0,
		EmergencyUnbondFee: sdkmath.LegacyMustNewDecFromStr("0.01"),
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateMaxUnbondings(p.MaxUnbondings); err != nil {
		return err
	}
	if err := validateUnbondingDuration(p.UnbondingDuration); err != nil {
		return err
	}
	return validateEmergencyUnbondFee(p.EmergencyUnbondFee)
}

func validateUnbondingDuration(v int64) error {
	// non-negative durations, including zero (instant unbond), are allowed
	if v < 0 {
		return fmt.Errorf("invalid unbonding duration: %d", v)
	}
	return nil
}

func validateEmergencyUnbondFee(v sdkmath.LegacyDec) error {
	if v.IsNil() || v.IsNegative() || v.GTE(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("invalid emergency unbonding fee: %s, valid values: [0, 1)", v)
	}

	return nil
}

func validateMaxUnbondings(_ uint32) error {
	// max unbondings can be any positive number, or zero for unlimited
	return nil
}
