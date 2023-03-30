package incentive

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const secondsPerDay = 86400

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxUnbondings:        5,
		UnbondingDuration:    secondsPerDay * 1,
		CommunityFundAddress: "",
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
	if err := validateCommunityFundAddress(p.CommunityFundAddress); err != nil {
		return err
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
