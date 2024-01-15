package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
)

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		CompleteLiquidationThreshold: sdkmath.LegacyMustNewDecFromStr("0.4"),
		MinimumCloseFactor:           sdkmath.LegacyMustNewDecFromStr("0.05"),
		OracleRewardFactor:           sdkmath.LegacyMustNewDecFromStr("0.01"),
		SmallLiquidationSize:         sdkmath.LegacyMustNewDecFromStr("500.00"),
		DirectLiquidationFee:         sdkmath.LegacyMustNewDecFromStr("0.05"),
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateLiquidationThreshold(p.CompleteLiquidationThreshold); err != nil {
		return err
	}
	if err := validateMinimumCloseFactor(p.MinimumCloseFactor); err != nil {
		return err
	}
	if err := validateOracleRewardFactor(p.OracleRewardFactor); err != nil {
		return err
	}
	if err := validateSmallLiquidationSize(p.SmallLiquidationSize); err != nil {
		return err
	}
	return validateDirectLiquidationFee(p.DirectLiquidationFee)
}

func validateLiquidationThreshold(v sdkmath.LegacyDec) error {
	if !v.IsPositive() {
		return fmt.Errorf("complete liquidation threshold must be positive: %d", v)
	}

	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("complete liquidation threshold cannot exceed 1: %d", v)
	}

	return nil
}

func validateMinimumCloseFactor(v sdkmath.LegacyDec) error {
	if v.IsNegative() {
		return fmt.Errorf("minimum close factor cannot be negative: %d", v)
	}
	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("minimum close factor cannot exceed 1: %d", v)
	}

	return nil
}

func validateOracleRewardFactor(v sdkmath.LegacyDec) error {
	if v.IsNegative() {
		return fmt.Errorf("oracle reward factor cannot be negative: %d", v)
	}
	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("oracle reward factor cannot exceed 1: %d", v)
	}

	return nil
}

func validateSmallLiquidationSize(v sdkmath.LegacyDec) error {
	if v.IsNegative() {
		return fmt.Errorf("small liquidation size cannot be negative: %d", v)
	}

	return nil
}

func validateDirectLiquidationFee(v sdkmath.LegacyDec) error {
	if v.IsNegative() {
		return fmt.Errorf("direct liquidation fee cannot be negative: %d", v)
	}
	if v.GTE(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("direct liquidation fee must be less than 1: %d", v)
	}

	return nil
}
