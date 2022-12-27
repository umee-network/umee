package ibctransfer

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = &Params{}

const (
	DefaultIBCPause = false
	// 24 hours time interval for ibc-transfer quota limit
	DefaultQuotaDurationPerDenom = time.Minute * 60 * 24
)

var (
	// 1M USD daily limit for all denoms
	DefaultTotalQuota = sdk.MustNewDecFromStr("1000000")
	// 600K USD dail limit for each denom
	DefaultQuotaPerIBCDenom = sdk.MustNewDecFromStr("600000")

	KeyIBCPause              = []byte("IBCPause")
	KeyTotalQuota            = []byte("KeyTotalQuota")
	KeyQuotaPerIBCDenom      = []byte("KeyQuotaPerIBCDenom")
	KeyQuotaDurationPerDenom = []byte("KeyQuotaDurationPerDenom")
)

func NewParams(ibcPause bool, totalQuota, quotaPerDenom sdk.Dec, quotaDurationPerDenom int64) Params {
	return Params{
		IbcPause:      ibcPause,
		TotalQuota:    totalQuota,
		TokenQuota:    quotaPerDenom,
		QuotaDuration: time.Second * time.Duration(quotaDurationPerDenom),
	}
}

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default genesis params
func DefaultParams() *Params {
	return &Params{
		IbcPause:      DefaultIBCPause,
		TotalQuota:    DefaultTotalQuota,
		TokenQuota:    DefaultQuotaPerIBCDenom,
		QuotaDuration: DefaultQuotaDurationPerDenom,
	}
}

func (p Params) Validate() error {
	if err := validateBoolean(p.IbcPause); err != nil {
		return err
	}

	if err := validateQuotaDuration(p.QuotaDuration); err != nil {
		return err
	}

	if err := validateTotalQuota(p.TotalQuota); err != nil {
		return err
	}

	if err := validateQuotaPerToken(p.TotalQuota); err != nil {
		return err
	}

	if p.TotalQuota.LT(p.TokenQuota) {
		return fmt.Errorf("token_quota shouldn't be less than token_quota")
	}

	return nil
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIBCPause, &p.IbcPause, validateBoolean),
		paramtypes.NewParamSetPair(KeyTotalQuota, &p.TotalQuota, validateTotalQuota),
		paramtypes.NewParamSetPair(KeyQuotaPerIBCDenom, &p.TokenQuota, validateQuotaPerToken),
		paramtypes.NewParamSetPair(KeyQuotaDurationPerDenom, &p.QuotaDuration, validateQuotaDuration),
	}
}

func validateBoolean(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateQuotaDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("quota duration time must be positive: %d", v)
	}

	return nil
}

func validateTotalQuota(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("total quota cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("total quota cannot be negative: %s", v)
	}

	return nil
}

func validateQuotaPerToken(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("quota per token cannot be nil: %s", v)
	}
	if v.IsNegative() {
		return fmt.Errorf("quota per token cannot be negative: %s", v)
	}

	return nil
}
