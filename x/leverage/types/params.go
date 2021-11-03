package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v3"
)

var _ paramtypes.ParamSet = &Params{}

var (
	KeyInterestEpoch = []byte("InterestEpoch")
)

var (
	defaultInterestEpoch = int64(100)
)

func NewParams(epoch int64) Params {
	return Params{InterestEpoch: epoch}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of x/leverage module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyInterestEpoch, &p.InterestEpoch, validateInterestEpoch),
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
		InterestEpoch: defaultInterestEpoch,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateInterestEpoch(p.InterestEpoch); err != nil {
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
		return fmt.Errorf("interest epoch be positive: %d", v)
	}
	return nil
}
