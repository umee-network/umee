package uibc

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Default ibc-transfer quota is disabled
	DefaultIBCPause = IBCTransferStatus_IBC_TRANSFER_STATUS_DISABLED
	// 24 hours time interval for ibc-transfer quota limit
	DefaultQuotaDurationPerDenom = 60 * 60 * 24
)

var (
	// 1M USD daily limit for all denoms
	DefaultTotalQuota = sdk.MustNewDecFromStr("1000000")
	// 600K USD daily limit for each denom
	DefaultQuotaPerIBCDenom = sdk.MustNewDecFromStr("600000")
)

// DefaultParams returns default genesis params
func DefaultParams() Params {
	return Params{
		IbcPause:      DefaultIBCPause,
		TotalQuota:    DefaultTotalQuota,
		TokenQuota:    DefaultQuotaPerIBCDenom,
		QuotaDuration: time.Second * DefaultQuotaDurationPerDenom,
	}
}

func (p Params) Validate() error {
	if err := validateIBCTransferStatus(p.IbcPause); err != nil {
		return err
	}

	if err := validateQuotaDuration(p.QuotaDuration); err != nil {
		return err
	}

	if err := validateQuota(p.TotalQuota, "total quota"); err != nil {
		return err
	}

	if err := validateQuota(p.TokenQuota, "quota per token"); err != nil {
		return err
	}

	if p.TotalQuota.LT(p.TokenQuota) {
		return fmt.Errorf("token quota shouldn't be less than quota per denom")
	}

	return nil
}

func validateIBCTransferStatus(status IBCTransferStatus) error {
	if status == IBCTransferStatus_IBC_TRANSFER_STATUS_DISABLED ||
		status == IBCTransferStatus_IBC_TRANSFER_STATUS_ENABLED ||
		status == IBCTransferStatus_IBC_TRANSFER_STATUS_PAUSED {
		return nil
	}

	return fmt.Errorf("invalid ibc-transfer status : %s", status.String())
}

func validateQuotaDuration(d time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("quota duration time must be positive: %d", d)
	}

	return nil
}

func validateQuota(q sdk.Dec, typ string) error {
	if q.IsNil() || q.IsNegative() {
		return fmt.Errorf("%s must be not negative: %s", typ, q)
	}
	return nil
}
