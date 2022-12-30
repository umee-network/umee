package uibc

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Default ibc-transfer quota is disabled
	DefaultIBCPause = IBCTransferStatus_DISABLED
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

func NewParams(ibcPause IBCTransferStatus, totalQuota, quotaPerDenom sdk.Dec, quotaDurationPerDenom int64) Params {
	return Params{
		IbcPause:      ibcPause,
		TotalQuota:    totalQuota,
		TokenQuota:    quotaPerDenom,
		QuotaDuration: time.Second * time.Duration(quotaDurationPerDenom),
	}
}

// DefaultParams returns default genesis params
func DefaultParams() Params {
	return Params{
		IbcPause:      DefaultIBCPause,
		TotalQuota:    DefaultTotalQuota,
		TokenQuota:    DefaultQuotaPerIBCDenom,
		QuotaDuration: DefaultQuotaDurationPerDenom,
	}
}

func (p Params) Validate() error {
	if err := validateIBCTransferStatus(p.IbcPause); err != nil {
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
		return fmt.Errorf("token quota shouldn't be less than quota per denom")
	}

	return nil
}

func validateIBCTransferStatus(status IBCTransferStatus) error {
	if status == IBCTransferStatus_DISABLED || status == IBCTransferStatus_ENABLED || status == IBCTransferStatus_PAUSED {
		return nil
	}

	return fmt.Errorf("invalid ibc-transfer status : %s", status.String())
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
