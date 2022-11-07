package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultIBCPause = false
)

var (
	// KeyHostEnabled is the store key for HostEnabled Params
	KeyIBCPause = []byte("IBCPause")
)

func NewParams(ibcPause bool) Params {
	return Params{IbcPause: ibcPause}
}

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() *Params {
	return &Params{IbcPause: DefaultIBCPause}
}

func (p *Params) Validate() error {
	if err := validateBoolean(p.IbcPause); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIBCPause, p.IbcPause, validateBoolean),
	}
}

func validateBoolean(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
