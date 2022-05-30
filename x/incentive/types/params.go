package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v3"
)

var _ paramtypes.ParamSet = &Params{}

var (
	KeyMaxUnlocks         = []byte("MaxUnlocks")
	KeyLockDurationLong   = []byte("KeyLockDurationLong")
	KeyLockDurationMiddle = []byte("KeyLockDurationMiddle")
	KeyLockDurationShort  = []byte("KeyLockDurationShort")
	KeyTierWeightMiddle   = []byte("KeyTierWeightMiddle")
	KeyTierWeightShort    = []byte("KeyTierWeightShort")
)

var (
	defaultMaxUnlocks         = uint32(20)
	defaultLockDurationShort  = uint64(86400)
	defaultLockDurationMiddle = uint64(604800)
	defaultLockDurationLong   = uint64(1209600)
	defaultTierWeightShort    = sdk.MustNewDecFromStr("0.5")
	defaultTierWeightMiddle   = sdk.MustNewDecFromStr("0.8")
)

func NewParams() Params {
	return Params{}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of x/incentive module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyMaxUnlocks,
			&p.MaxUnlocks,
			validateMaxUnlocks,
		),
		paramtypes.NewParamSetPair(
			KeyLockDurationLong,
			&p.LockDurationLong,
			validateLockDuration,
		),
		paramtypes.NewParamSetPair(
			KeyLockDurationLong,
			&p.LockDurationLong,
			validateLockDuration,
		),
		paramtypes.NewParamSetPair(
			KeyLockDurationLong,
			&p.LockDurationLong,
			validateLockDuration,
		),
		paramtypes.NewParamSetPair(
			KeyTierWeightMiddle,
			&p.TierWeightMiddle,
			validateTierWeight,
		),
		paramtypes.NewParamSetPair(
			KeyTierWeightShort,
			&p.TierWeightShort,
			validateTierWeight,
		),
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
		MaxUnlocks:         defaultMaxUnlocks,
		LockDurationLong:   defaultLockDurationLong,
		LockDurationMiddle: defaultLockDurationMiddle,
		LockDurationShort:  defaultLockDurationShort,
		TierWeightMiddle:   defaultTierWeightMiddle,
		TierWeightShort:    defaultTierWeightShort,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateMaxUnlocks(p.MaxUnlocks); err != nil {
		return err
	}
	if err := validateLockDuration(p.LockDurationLong); err != nil {
		return err
	}
	if err := validateLockDuration(p.LockDurationMiddle); err != nil {
		return err
	}
	if err := validateLockDuration(p.LockDurationShort); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightMiddle); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightShort); err != nil {
		return err
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

func validateMaxUnlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max unlocks cannot be zero")
	}

	return nil
}
