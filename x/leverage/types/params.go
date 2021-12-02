package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v3"
)

var _ paramtypes.ParamSet = &Params{}

var (
	KeyInterestEpoch                = []byte("InterestEpoch")
	KeyCompleteLiquidationThreshold = []byte("CompleteLiquidationThreshold")
	KeyMinimumCloseFactor           = []byte("MinimumCloseFactor")
)

var (
	defaultInterestEpoch                = int64(100)
	defaultCompleteLiquidationThreshold = sdk.MustNewDecFromStr("0.1")
	defaultMinimumCloseFactor           = sdk.MustNewDecFromStr("0.01")
)

func NewParams(epoch int64) Params {
	return Params{InterestEpoch: epoch}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of x/leverage module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyInterestEpoch, &p.InterestEpoch, validateInterestEpoch),
		paramtypes.NewParamSetPair(KeyCompleteLiquidationThreshold, &p.CompleteLiquidationThreshold,
			validateLiquidationThreshold),
		paramtypes.NewParamSetPair(KeyMinimumCloseFactor, &p.MinimumCloseFactor, validateMinimumCloseFactor),
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
		InterestEpoch:                defaultInterestEpoch,
		CompleteLiquidationThreshold: defaultCompleteLiquidationThreshold,
		MinimumCloseFactor:           defaultMinimumCloseFactor,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateInterestEpoch(p.InterestEpoch); err != nil {
		return err
	}
	if err := validateLiquidationThreshold(p.CompleteLiquidationThreshold); err != nil {
		return err
	}
	if err := validateMinimumCloseFactor(p.MinimumCloseFactor); err != nil {
		return err
	}
	return nil
}

func validateInterestEpoch(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("interest epoch must be positive: %d", v)
	}
	return nil
}

func validateLiquidationThreshold(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("complete liquidation threshold cannot be negative: %d", v)
	}
	if v.GT(sdk.MustNewDecFromStr("0.1")) {
		return fmt.Errorf("complete liquidation threshold cannot exceed 0.1: %d", v)
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
	if v.GT(sdk.MustNewDecFromStr("1")) {
		return fmt.Errorf("minimum close factor cannot exceed 1: %d", v)
	}
	return nil
}
