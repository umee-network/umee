package incentive

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const secondsPerDay = 86400

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxUnbondings:      5,
		UnbondingDuration:  secondsPerDay * 1,
		EmergencyUnbondFee: sdk.MustNewDecFromStr("0.01"),
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

func validateEmergencyUnbondFee(v sdk.Dec) error {
	if v.IsNil() || v.IsNegative() || v.GTE(sdk.OneDec()) {
		return fmt.Errorf("invalid emergency unbonding fee: %s, valid values: [0, 1)", v)
	}

	return nil
}

func validateMaxUnbondings(_ uint32) error {
	// max unbondings can be any positive number, or zero for unlimited
	return nil
}
