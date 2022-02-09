package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v3"
)

var _ paramtypes.ParamSet = &Params{}

var (
	KeyCompleteLiquidationThreshold = []byte("CompleteLiquidationThreshold")
	KeyMinimumCloseFactor           = []byte("MinimumCloseFactor")
	KeyOracleRewardFactor           = []byte("OracleRewardFactor")
)

var (
	defaultCompleteLiquidationThreshold = sdk.MustNewDecFromStr("0.1")
	defaultMinimumCloseFactor           = sdk.MustNewDecFromStr("0.01")
	defaultOracleRewardFactor           = sdk.MustNewDecFromStr("0.01")
)

func NewParams() Params {
	return Params{}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of x/leverage module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyCompleteLiquidationThreshold, &p.CompleteLiquidationThreshold,
			validateLiquidationThreshold),
		paramtypes.NewParamSetPair(KeyMinimumCloseFactor, &p.MinimumCloseFactor, validateMinimumCloseFactor),
		paramtypes.NewParamSetPair(KeyOracleRewardFactor, &p.OracleRewardFactor, validateOracleRewardFactor),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamKeyTable returns the x/leverage module's parameter KeyTable expected by
// the x/params module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		CompleteLiquidationThreshold: defaultCompleteLiquidationThreshold,
		MinimumCloseFactor:           defaultMinimumCloseFactor,
		OracleRewardFactor:           defaultOracleRewardFactor,
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
	return nil
}

func validateLiquidationThreshold(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !v.IsPositive() {
		return fmt.Errorf("complete liquidation threshold must be positive: %d", v)
	}
	return nil
}

func validateMinimumCloseFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum close factor cannot be negative: %d", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("minimum close factor cannot exceed 1: %d", v)
	}
	return nil
}

func validateOracleRewardFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("oracle reward factor cannot be negative: %d", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("oracle reward factor cannot exceed 1: %d", v)
	}
	return nil
}
