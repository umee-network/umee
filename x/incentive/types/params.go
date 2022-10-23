package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v3"
)

const blocksPerDay = 86400 / 5

var _ paramtypes.ParamSet = &Params{}

var (
	KeyMaxUnbondings           = []byte("MaxUnbondings")
	KeyUnbondingDurationLong   = []byte("KeyUnbondingDurationLong")
	KeyUnbondingDurationMiddle = []byte("KeyUnbondingDurationMiddle")
	KeyUnbondingDurationShort  = []byte("KeyUnbondingDurationShort")
	KeyTierWeightMiddle        = []byte("KeyTierWeightMiddle")
	KeyTierWeightShort         = []byte("KeyTierWeightShort")
)

var (
	defaultMaxUnbondings           = uint32(20)
	defaultUnbondingDurationShort  = uint64(blocksPerDay)
	defaultUnbondingDurationMiddle = uint64(blocksPerDay * 7)
	defaultUnbondingDurationLong   = uint64(blocksPerDay * 14)
	defaultTierWeightShort         = sdk.MustNewDecFromStr("0.5")
	defaultTierWeightMiddle        = sdk.MustNewDecFromStr("0.8")
)

func NewParams() Params {
	return Params{}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value
// pairs pairs of x/incentive module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyMaxUnbondings,
			&p.MaxUnbondings,
			validateMaxUnbondings,
		),
		paramtypes.NewParamSetPair(
			KeyUnbondingDurationLong,
			&p.UnbondingDurationLong,
			validateLockDuration,
		),
		paramtypes.NewParamSetPair(
			KeyUnbondingDurationMiddle,
			&p.UnbondingDurationMiddle,
			validateLockDuration,
		),
		paramtypes.NewParamSetPair(
			KeyUnbondingDurationShort,
			&p.UnbondingDurationShort,
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
		MaxUnbondings:           defaultMaxUnbondings,
		UnbondingDurationLong:   defaultUnbondingDurationLong,
		UnbondingDurationMiddle: defaultUnbondingDurationMiddle,
		UnbondingDurationShort:  defaultUnbondingDurationShort,
		TierWeightMiddle:        defaultTierWeightMiddle,
		TierWeightShort:         defaultTierWeightShort,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateMaxUnbondings(p.MaxUnbondings); err != nil {
		return err
	}
	if err := validateLockDuration(p.UnbondingDurationLong); err != nil {
		return err
	}
	if err := validateLockDuration(p.UnbondingDurationMiddle); err != nil {
		return err
	}
	if err := validateLockDuration(p.UnbondingDurationShort); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightMiddle); err != nil {
		return err
	}
	if err := validateTierWeight(p.TierWeightShort); err != nil {
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

func validateMaxUnbondings(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max unbondings cannot be zero")
	}

	return nil
}
